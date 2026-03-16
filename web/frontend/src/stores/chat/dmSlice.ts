import type { ChatSliceCreator, DmSlice, ChatMessage } from './chatStore.types';
import { chatApi } from '../../services/api/chat';
import { upsertMessage, genClientMsgId } from './helpers';
import { useAuthStore } from '../authStore';
import { MESSAGE_PAGE_SIZE } from '../../constants';

export const createDmSlice: ChatSliceCreator<DmSlice> = (set, get) => ({
    messagesByDm: {},
    conversations: [],
    unreadDmCounts: {},
    hasMoreByDm: {},
    isLoadingOlderByDm: {},

    fetchConversations: async () => {
        try {
            const res = await chatApi.getDmConversations();
            const conversations = Array.isArray(res) ? res : ((res as any).items || []);
            set({ conversations });
        } catch (e) {
            console.error('Failed to fetch conversations', e);
        }
    },

    clearUnreadDmCount: (userId) => {
        set(state => ({ unreadDmCounts: { ...state.unreadDmCounts, [userId]: 0 } }));
    },

    fetchDmHistory: async (userId) => {
        try {
            const res = await chatApi.getDmMessages(userId, MESSAGE_PAGE_SIZE);
            const msgs = Array.isArray(res) ? res : ((res as any).items || []);
            set(state => {
                const existing = state.messagesByDm[userId] || [];
                let merged = [...existing];
                for (const m of msgs) {
                    merged = upsertMessage(merged, m);
                }
                merged.sort((a, b) => a.id - b.id);
                return {
                    messagesByDm: { ...state.messagesByDm, [userId]: merged },
                    hasMoreByDm: { ...state.hasMoreByDm, [userId]: state.hasMoreByDm[userId] !== false ? msgs.length >= MESSAGE_PAGE_SIZE : false },
                };
            });
        } catch (e) {
            console.error(`Failed to fetch DM history for user: ${userId}`, e);
        }
    },

    fetchOlderDmHistory: async (userId) => {
        const state = get();
        if (state.isLoadingOlderByDm[userId] || !state.hasMoreByDm[userId]) return;
        const msgs = state.messagesByDm[userId] || [];
        const oldestId = msgs.find((m: ChatMessage) => m.id > 0)?.id;
        if (!oldestId) return;

        set(s => ({ isLoadingOlderByDm: { ...s.isLoadingOlderByDm, [userId]: true } }));
        try {
            const res = await chatApi.getDmMessagesBefore(userId, oldestId, MESSAGE_PAGE_SIZE);
            const older = Array.isArray(res) ? res : ((res as any).items || []);
            set(s => ({
                messagesByDm: { ...s.messagesByDm, [userId]: [...older, ...(s.messagesByDm[userId] || [])] },
                hasMoreByDm: { ...s.hasMoreByDm, [userId]: older.length >= MESSAGE_PAGE_SIZE },
                isLoadingOlderByDm: { ...s.isLoadingOlderByDm, [userId]: false },
            }));
        } catch (e) {
            console.error(`Failed to fetch older DM history for user: ${userId}`, e);
            set(s => ({ isLoadingOlderByDm: { ...s.isLoadingOlderByDm, [userId]: false } }));
        }
    },

    sendDmMessage: (userId, content, msgType = 'text', replyToId, fileMeta) => {
        const { socket, isConnected } = get();
        if (!socket || !isConnected) return;

        const client_msg_id = genClientMsgId();
        socket.send(JSON.stringify({
            type: 'dm',
            data: {
                target_user_id: userId,
                content,
                client_msg_id,
                msg_type: msgType,
                ...(replyToId ? { reply_to_id: replyToId } : {}),
                ...(fileMeta ? { file_size: fileMeta.file_size, mime_type: fileMeta.mime_type } : {}),
            }
        }));

        const user = useAuthStore.getState().user;
        if (user) {
            const tempMsg: ChatMessage = {
                id: -Date.now(),
                client_msg_id,
                content,
                msg_type: msgType,
                user,
                file_size: fileMeta?.file_size,
                mime_type: fileMeta?.mime_type,
                created_at: new Date().toISOString()
            };
            set(state => ({
                messagesByDm: {
                    ...state.messagesByDm,
                    [userId]: upsertMessage(state.messagesByDm[userId] || [], tempMsg)
                }
            }));
        }
    },

    sendDmTyping: (userId) => {
        const { socket, isConnected } = get();
        if (!socket || !isConnected) return;
        socket.send(JSON.stringify({ type: 'dm_typing', data: { target_user_id: userId } }));
    },
});
