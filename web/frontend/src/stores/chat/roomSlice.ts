import type { ChatSliceCreator, RoomSlice, ChatMessage } from './chatStore.types';
import { chatApi } from '../../services/api/chat';
import { roomsApi } from '../../services/api/rooms';
import { upsertMessage, genClientMsgId } from './helpers';
import { useAuthStore } from '../authStore';
import { MESSAGE_PAGE_SIZE } from '../../constants';
import { trackEvent } from '../../utils/metrics';

export const createRoomSlice: ChatSliceCreator<RoomSlice> = (set, get) => ({
    rooms: [],
    messagesByRoom: {},
    onlineUsersByRoom: {},
    hasMoreByRoom: {},
    isLoadingOlderByRoom: {},
    activeThreadMsgId: null,
    threadMessagesByMsgId: {},
    pinnedMessages: {},

    fetchRooms: async () => {
        try {
            const rooms = await roomsApi.list();
            set(state => {
                let currentActive = state.activeRoomId;
                if (currentActive !== null && !rooms.some(r => r.id === currentActive)) {
                    currentActive = null;
                }
                return { rooms, activeRoomId: currentActive };
            });
            if (rooms.length > 0) {
                const pairs = await Promise.all(rooms.map(async (r) => {
                    try {
                        const res = await roomsApi.getUnreadCount(r.id);
                        return [r.id, res?.unread_count ?? 0] as const;
                    } catch {
                        return [r.id, 0] as const;
                    }
                }));
                set(state => ({
                    unreadCountsByRoom: pairs.reduce((acc, [id, cnt]) => {
                        acc[id] = cnt;
                        return acc;
                    }, { ...(state.unreadCountsByRoom || {}) } as Record<number, number>)
                }));
            }
        } catch (e) {
            console.error('Failed to fetch rooms', e);
        }
    },

    setActiveThreadMsgId: (msgId) => {
        set({ activeThreadMsgId: msgId });
    },

    pinMessage: async (roomId, msgId) => {
        try {
            await chatApi.pinMessage(roomId, msgId);
            await get().fetchPinnedMessages(roomId);
        } catch (e) {
            console.error('Failed to pin message', e);
            throw e;
        }
    },

    unpinMessage: async (roomId, msgId) => {
        try {
            await chatApi.unpinMessage(roomId, msgId);
            await get().fetchPinnedMessages(roomId);
        } catch (e) {
            console.error('Failed to unpin message', e);
            throw e;
        }
    },

    fetchPinnedMessages: async (roomId) => {
        try {
            const res = await chatApi.getPinnedMessages(roomId);
            const msgs = Array.isArray(res) ? res : ((res as any).items || []);
            set(state => ({
                pinnedMessages: { ...state.pinnedMessages, [roomId]: msgs }
            }));
        } catch (e) {
            console.error(`Failed to fetch pinned messages for room: ${roomId}`, e);
        }
    },

    fetchThreadMessages: async (roomId, msgId) => {
        try {
            const res = await chatApi.getThreadMessages(roomId, msgId, MESSAGE_PAGE_SIZE);
            const msgs = Array.isArray(res) ? res : ((res as any).items || []);
            set(state => ({
                threadMessagesByMsgId: { ...state.threadMessagesByMsgId, [msgId]: msgs }
            }));
        } catch (e) {
            console.error(`Failed to fetch thread messages: ${msgId}`, e);
        }
    },
    batchDeleteMessages: async (roomId, msgIds) => {
        if (!msgIds.length) return;
        try {
            await chatApi.deleteRoomMessagesBatch(roomId, msgIds);
            const now = new Date().toISOString();
            set(state => ({
                messagesByRoom: {
                    ...state.messagesByRoom,
                    [roomId]: (state.messagesByRoom[roomId] || []).map(m =>
                        msgIds.includes(m.id) ? { ...m, deleted_at: now } : m
                    )
                }
            }));
        } catch (e) {
            console.error('Failed to batch delete messages', e);
            throw e;
        }
    },

    fetchHistory: async (roomId) => {
        try {
            const res = await chatApi.getRoomMessages(roomId, MESSAGE_PAGE_SIZE);
            const msgs = Array.isArray(res) ? res : ((res as any).items || []);
            set(state => {
                const existing = state.messagesByRoom[roomId] || [];
                let merged = [...existing];
                for (const m of msgs) {
                    merged = upsertMessage(merged, m);
                }
                merged.sort((a, b) => a.id - b.id);
                return {
                    messagesByRoom: { ...state.messagesByRoom, [roomId]: merged },
                    hasMoreByRoom: { ...state.hasMoreByRoom, [roomId]: state.hasMoreByRoom[roomId] !== false ? msgs.length >= MESSAGE_PAGE_SIZE : false },
                };
            });
            trackEvent('message_load', { room_id: roomId, count: msgs.length, hasMore: msgs.length >= MESSAGE_PAGE_SIZE });
        } catch (e) {
            console.error(`Failed to fetch history for room: ${roomId}`, e);
        }
    },

    fetchOlderHistory: async (roomId) => {
        const state = get();
        if (state.isLoadingOlderByRoom[roomId] || !state.hasMoreByRoom[roomId]) return;
        const msgs = state.messagesByRoom[roomId] || [];
        const oldestId = msgs.find((m: ChatMessage) => m.id > 0)?.id;
        if (!oldestId) return;

        set(s => ({ isLoadingOlderByRoom: { ...s.isLoadingOlderByRoom, [roomId]: true } }));
        try {
            const res = await chatApi.getRoomMessagesBefore(roomId, oldestId, MESSAGE_PAGE_SIZE);
            const older = Array.isArray(res) ? res : ((res as any).items || []);
            set(s => ({
                messagesByRoom: { ...s.messagesByRoom, [roomId]: [...older, ...(s.messagesByRoom[roomId] || [])] },
                hasMoreByRoom: { ...s.hasMoreByRoom, [roomId]: older.length >= MESSAGE_PAGE_SIZE },
                isLoadingOlderByRoom: { ...s.isLoadingOlderByRoom, [roomId]: false },
            }));
            trackEvent('scroll_load_older', { room_id: roomId, success: true, count: older.length });
        } catch (e) {
            console.error(`Failed to fetch older history for room: ${roomId}`, e);
            set(s => ({ isLoadingOlderByRoom: { ...s.isLoadingOlderByRoom, [roomId]: false } }));
            trackEvent('scroll_load_older', { room_id: roomId, success: false });
        }
    },

    sendMessage: (roomId, content, msgType = 'text', replyToId, fileMeta) => {
        const { socket, isConnected } = get();
        if (!socket || !isConnected) return;

        const client_msg_id = genClientMsgId();
        socket.send(JSON.stringify({
            type: 'chat',
            data: {
                room_id: roomId,
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
                room_id: roomId,
                mentions: [],
                file_size: fileMeta?.file_size,
                mime_type: fileMeta?.mime_type,
                created_at: new Date().toISOString()
            };
            set(state => ({
                messagesByRoom: {
                    ...state.messagesByRoom,
                    [roomId]: upsertMessage(state.messagesByRoom[roomId] || [], tempMsg)
                }
            }));
        }
    },

    sendTyping: (roomId) => {
        const { socket, isConnected } = get();
        if (!socket || !isConnected) return;
        socket.send(JSON.stringify({ type: 'typing', data: { room_id: roomId } }));
    },

    joinRoom: (roomId) => {
        const { socket, isConnected } = get();
        if (!socket || !isConnected) return;
        socket.send(JSON.stringify({ type: 'join_room', data: { room_id: roomId } }));
    },

    leaveRoom: (roomId) => {
        const { socket, isConnected } = get();
        if (!socket || !isConnected) return;
        socket.send(JSON.stringify({ type: 'leave_room', data: { room_id: roomId } }));
    },

    retractMessage: async (msgId, roomId, dmUserId) => {
        try {
            if (roomId) {
                await chatApi.deleteRoomMessage(roomId, msgId);
                set(state => ({
                    messagesByRoom: {
                        ...state.messagesByRoom,
                        [roomId]: (state.messagesByRoom[roomId] || []).map(m =>
                            m.id === msgId ? { ...m, deleted_at: new Date().toISOString() } : m
                        )
                    }
                }));
            } else if (dmUserId) {
                await chatApi.deleteDmMessage(msgId);
                set(state => ({
                    messagesByDm: {
                        ...state.messagesByDm,
                        [dmUserId]: (state.messagesByDm[dmUserId] || []).map(m =>
                            m.id === msgId ? { ...m, deleted_at: new Date().toISOString() } : m
                        )
                    }
                }));
            }
        } catch (e: any) {
            console.error('Failed to retract message', e);
            throw e;
        }
    },

    editMessage: async (msgId, content, roomId, dmUserId) => {
        const nextContent = content.trim();
        if (!nextContent) {
            throw new Error('消息内容不能为空');
        }
        const updatedAt = new Date().toISOString();
        try {
            if (roomId) {
                await chatApi.updateRoomMessage(roomId, msgId, nextContent);
                trackEvent('edit_message', { msg_type: 'text', char_count: Array.from(nextContent).length });
                set(state => ({
                    messagesByRoom: {
                        ...state.messagesByRoom,
                        [roomId]: (state.messagesByRoom[roomId] || []).map(m =>
                            m.id === msgId ? { ...m, content: nextContent, updated_at: updatedAt } : m
                        )
                    }
                }));
            } else if (dmUserId) {
                await chatApi.updateDmMessage(msgId, nextContent);
                trackEvent('edit_message', { msg_type: 'text', char_count: Array.from(nextContent).length });
                set(state => ({
                    messagesByDm: {
                        ...state.messagesByDm,
                        [dmUserId]: (state.messagesByDm[dmUserId] || []).map(m =>
                            m.id === msgId ? { ...m, content: nextContent, updated_at: updatedAt } : m
                        )
                    }
                }));
            }
        } catch (e: any) {
            console.error('Failed to edit message', e);
            throw e;
        }
    },

    setMessageReactions: (msgId, reactions, isDm, dmUserId) => {
        if (!isDm) {
            set(state => {
                const newMsgsByRoom = { ...state.messagesByRoom };
                for (const roomIdStr of Object.keys(newMsgsByRoom)) {
                    const roomId = Number(roomIdStr);
                    if (newMsgsByRoom[roomId]?.some(m => m.id === msgId)) {
                        newMsgsByRoom[roomId] = newMsgsByRoom[roomId].map(m =>
                            m.id === msgId ? { ...m, reactions } : m
                        );
                        break;
                    }
                }
                return { messagesByRoom: newMsgsByRoom };
            });
        } else if (dmUserId !== null) {
            set(state => ({
                messagesByDm: {
                    ...state.messagesByDm,
                    [dmUserId]: (state.messagesByDm[dmUserId] || []).map(m =>
                        m.id === msgId ? { ...m, reactions } : m
                    ),
                },
            }));
        }
    },
});
