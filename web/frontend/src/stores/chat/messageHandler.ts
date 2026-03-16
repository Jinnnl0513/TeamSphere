import type { ChatStoreState, ChatMessage, SystemMessage, ReactionSummary } from './chatStore.types';
import type { User } from '../../types/models';
import { TYPING_EXPIRY_MS } from '../../constants';
import { useAuthStore } from '../authStore';
import { useFriendStore } from '../friendStore';
import { sendDesktopNotification, useNotificationStore } from '../notificationStore';
import { upsertMessage } from './helpers';
import toast from 'react-hot-toast';

type GetFn = () => ChatStoreState;
type SetFn = (partial: Partial<ChatStoreState> | ((s: ChatStoreState) => Partial<ChatStoreState>)) => void;

function getMessagePreview(msg: ChatMessage) {
    if (msg.msg_type === 'image') return '[图片]';
    if (msg.msg_type === 'file') return '[文件]';
    return msg.content;
}

const applyReaction = (reactions: ReactionSummary[] | undefined, emoji: string, count: number, user_ids: number[]) => {
    const existing = reactions ? [...reactions] : [];
    const idx = existing.findIndex(r => r.emoji === emoji);
    if (count === 0) {
        return existing.filter(r => r.emoji !== emoji);
    }
    const updatedReaction = { emoji, count, user_ids: user_ids ?? [] };
    if (idx === -1) return [...existing, updatedReaction];
    const updated = [...existing];
    updated[idx] = updatedReaction;
    return updated;
};

export function handleIncomingMessage(
    type: string,
    data: any,
    get: GetFn,
    set: SetFn
) {
    const myId = useAuthStore.getState().user?.id;

    switch (type) {
        case 'chat': {
            const newMsg = data as ChatMessage;
            const room_id = newMsg.room_id as number;

            set(state => ({
                messagesByRoom: {
                    ...state.messagesByRoom,
                    [room_id]: upsertMessage(state.messagesByRoom[room_id] || [], newMsg)
                }
            }));

            if (newMsg.user?.id !== myId) {
                const roomName = get().rooms.find(r => r.id === room_id)?.name || '频道';
                const { soundEnabled } = useNotificationStore.getState();
                if (soundEnabled) {
                    const audio = new Audio('/drip.wav');
                    audio.play().catch(console.error);
                }
                const title = `#${roomName} - ${newMsg.user?.username || '某人'}`;
                const body = getMessagePreview(newMsg);

                const state = get();
                const isActiveRoom = state.activeView === 'rooms' && state.activeRoomId === room_id && document.hasFocus();

                if (!isActiveRoom) {
                    if (document.hasFocus()) {
                        toast(title + ': ' + body, { icon: '🔔', duration: 3000, style: { background: 'var(--bg-secondary)', color: 'var(--text-main)' } });
                    } else {
                        sendDesktopNotification(title, body, newMsg.user?.avatar_url);
                    }
                }
            }
            break;
        }

        case 'dm': {
            const newMsg = data as ChatMessage;
            let listKey = newMsg.user?.id;
            if (listKey === myId) listKey = get().activeDmUserId || listKey;
            if (!listKey) break;

            set(state => ({
                messagesByDm: {
                    ...state.messagesByDm,
                    [listKey]: upsertMessage(state.messagesByDm[listKey] || [], newMsg)
                }
            }));

            if (newMsg.user?.id !== myId) {
                const { soundEnabled } = useNotificationStore.getState();
                if (soundEnabled) {
                    const audio = new Audio('/drip.wav');
                    audio.play().catch(console.error);
                }
                const title = `${newMsg.user?.username || '某人'} 发来私信`;
                const body = getMessagePreview(newMsg);

                const state = get();
                const isActiveDm = state.activeView === 'dm' && state.activeDmUserId === listKey && document.hasFocus();

                if (!isActiveDm) {
                    if (document.hasFocus()) {
                        toast(title + ': ' + body, { icon: '✉️', duration: 3000, style: { background: 'var(--bg-secondary)', color: 'var(--text-main)' } });
                    } else {
                        sendDesktopNotification(title, body, newMsg.user?.avatar_url);
                    }
                }

                if (state.activeView !== 'dm' || state.activeDmUserId !== listKey) {
                    set(s => ({
                        unreadDmCounts: {
                            ...s.unreadDmCounts,
                            [listKey]: (s.unreadDmCounts[listKey] || 0) + 1
                        }
                    }));
                }
            }
            get().fetchConversations();
            break;
        }

        case 'system': {
            const sysMsg = data as SystemMessage;
            if (sysMsg.room_id) {
                const sysChatMsg: ChatMessage = {
                    id: -Date.now(),
                    content: sysMsg.content,
                    msg_type: 'system',
                    user: { id: 0, username: 'System', avatar_url: '', role: 'system' } as User,
                    room_id: sysMsg.room_id,
                    created_at: new Date().toISOString()
                };
                set(state => ({
                    messagesByRoom: {
                        ...state.messagesByRoom,
                        [sysMsg.room_id!]: [...(state.messagesByRoom[sysMsg.room_id!] || []), sysChatMsg]
                    }
                }));
            }
            break;
        }

        case 'online_users': {
            const { room_id, users } = data as { room_id: number; users: User[] };
            set(state => ({
                onlineUsersByRoom: { ...state.onlineUsersByRoom, [room_id]: users }
            }));
            break;
        }

        case 'chat_ack':
            break;

        case 'dm_sent':
            get().fetchConversations();
            break;

        case 'friend_online':
            useFriendStore.getState().setFriendOnline(data.user_id);
            break;

        case 'friend_offline':
            useFriendStore.getState().setFriendOffline(data.user_id);
            break;

        case 'friends_online_list':
            useFriendStore.getState().setOnlineFriendIds(data.user_ids || []);
            break;

        case 'msg_recalled': {
            const { msg_id, room_id } = data as { msg_id: number; room_id: number };
            set(state => ({
                messagesByRoom: {
                    ...state.messagesByRoom,
                    [room_id]: (state.messagesByRoom[room_id] || []).map(m =>
                        m.id === msg_id ? { ...m, deleted_at: new Date().toISOString() } : m
                    )
                }
            }));
            break;
        }

        case 'dm_recalled': {
            const { msg_id } = data as { msg_id: number };
            set(state => {
                const newDmMsgs = { ...state.messagesByDm };
                for (const uid in newDmMsgs) {
                    newDmMsgs[uid] = newDmMsgs[uid].map(m =>
                        m.id === msg_id ? { ...m, deleted_at: new Date().toISOString() } : m
                    );
                }
                return { messagesByDm: newDmMsgs };
            });
            get().fetchConversations();
            break;
        }

        case 'msg_edited': {
            const { msg_id, room_id, content, updated_at } = data as { msg_id: number; room_id: number; content: string; updated_at?: string };
            set(state => ({
                messagesByRoom: {
                    ...state.messagesByRoom,
                    [room_id]: (state.messagesByRoom[room_id] || []).map(m =>
                        m.id === msg_id ? { ...m, content, updated_at: updated_at || new Date().toISOString() } : m
                    )
                }
            }));
            break;
        }

        case 'dm_edited': {
            const { msg_id, content, updated_at } = data as { msg_id: number; content: string; updated_at?: string };
            set(state => {
                const newDmMsgs = { ...state.messagesByDm };
                for (const uid in newDmMsgs) {
                    newDmMsgs[uid] = newDmMsgs[uid].map(m =>
                        m.id === msg_id ? { ...m, content, updated_at: updated_at || new Date().toISOString() } : m
                    );
                }
                return { messagesByDm: newDmMsgs };
            });
            break;
        }

        case 'dm_read': {
            const { peer_id, last_read_msg_id, read_at } = data as { peer_id: number; last_read_msg_id: number; read_at?: string };
            if (!peer_id || !last_read_msg_id) break;
            const now = read_at || new Date().toISOString();
            set(state => {
                const newDmMsgs = { ...state.messagesByDm };
                const list = newDmMsgs[peer_id] || [];
                newDmMsgs[peer_id] = list.map(m => {
                    if (m.user?.id === myId && m.id <= last_read_msg_id) {
                        return { ...m, read_at: now };
                    }
                    return m;
                });
                return { messagesByDm: newDmMsgs };
            });
            break;
        }

        case 'room_member_muted': {
            const { room_id, user, muted_until } = data as { room_id: number; user: any; muted_until: string };
            const sysMsg: ChatMessage = {
                id: -Date.now(),
                content: `${user.username} 被禁言至 ${new Date(muted_until).toLocaleString()}`,
                msg_type: 'system',
                user: { id: 0, username: 'System', avatar_url: '', role: 'system' } as User,
                room_id,
                created_at: new Date().toISOString()
            };
            set(state => ({
                messagesByRoom: {
                    ...state.messagesByRoom,
                    [room_id]: [...(state.messagesByRoom[room_id] || []), sysMsg]
                }
            }));
            if (myId === user.id) {
                get().setMyMutedUntil(room_id, muted_until);
                const msg = `你在房间 #${room_id} 被禁言至 ${new Date(muted_until).toLocaleString()}`;
                if (document.hasFocus()) toast.error(msg);
                else sendDesktopNotification('系统通知', msg);
            }
            break;
        }

        case 'room_member_unmuted': {
            const { room_id, user } = data as { room_id: number; user: any };
            const sysMsg: ChatMessage = {
                id: -Date.now(),
                content: `${user.username} 的禁言已解除` ,
                msg_type: 'system',
                user: { id: 0, username: 'System', avatar_url: '', role: 'system' } as User,
                room_id,
                created_at: new Date().toISOString()
            };
            set(state => ({
                messagesByRoom: {
                    ...state.messagesByRoom,
                    [room_id]: [...(state.messagesByRoom[room_id] || []), sysMsg]
                }
            }));
            if (myId === user.id) {
                get().setMyMutedUntil(room_id, null);
                const msg = `你在房间 #${room_id} 的禁言已解除`;
                if (document.hasFocus()) toast.success(msg);
                else sendDesktopNotification('系统通知', msg);
            }
            break;
        }

        case 'msg_pinned': {
            const { room_id, msg_id } = data as { room_id: number; msg_id: number };
            if (!room_id || !msg_id) break;
            const state = get();
            const roomMsgs = state.messagesByRoom[room_id] || [];
            const msg = roomMsgs.find(m => m.id === msg_id);
            if (!msg) break;
            set(s => ({
                pinnedMessages: {
                    ...s.pinnedMessages,
                    [room_id]: [msg, ...(s.pinnedMessages[room_id] || []).filter(m => m.id !== msg_id)]
                }
            }));
            break;
        }

        case 'msg_unpinned': {
            const { room_id, msg_id } = data as { room_id: number; msg_id: number };
            if (!room_id || !msg_id) break;
            set(s => ({
                pinnedMessages: {
                    ...s.pinnedMessages,
                    [room_id]: (s.pinnedMessages[room_id] || []).filter(m => m.id !== msg_id)
                }
            }));
            break;
        }

        case 'unread_update': {
            const { room_id, unread_count } = data as { room_id: number; unread_count: number };
            if (!room_id && room_id !== 0) break;
            get().setUnreadCount(room_id, unread_count);
            break;
        }

        case 'thread_reply': {
            const { root_msg_id, message } = data as { root_msg_id: number; message: ChatMessage };
            if (!root_msg_id || !message) break;
            set(state => ({
                threadMessagesByMsgId: {
                    ...state.threadMessagesByMsgId,
                    [root_msg_id]: upsertMessage(state.threadMessagesByMsgId[root_msg_id] || [], message)
                }
            }));
            break;
        }

        case 'typing': {
            const { room_id, username, user_id } = data as { room_id: number; user_id: number; username: string };
            if (user_id === myId) break;
            const key = `room_${room_id}`;
            set(state => {
                const now = Date.now();
                const existing = (state.typingUsers[key] || []).filter(u => u.expiresAt > now);
                const idx = existing.findIndex(u => u.username === username);
                if (idx !== -1) existing[idx].expiresAt = now + TYPING_EXPIRY_MS;
                else existing.push({ username, expiresAt: now + TYPING_EXPIRY_MS });
                return { typingUsers: { ...state.typingUsers, [key]: existing } };
            });
            break;
        }

        case 'dm_typing': {
            const { user_id, username } = data as { user_id: number; username: string };
            if (user_id === myId) break;
            const key = `dm_${user_id}`;
            set(state => {
                const now = Date.now();
                const existing = (state.typingUsers[key] || []).filter(u => u.expiresAt > now);
                const idx = existing.findIndex(u => u.username === username);
                if (idx !== -1) existing[idx].expiresAt = now + TYPING_EXPIRY_MS;
                else existing.push({ username, expiresAt: now + TYPING_EXPIRY_MS });
                return { typingUsers: { ...state.typingUsers, [key]: existing } };
            });
            break;
        }

        case 'reaction_added':
        case 'reaction_removed': {
            const { message_id, message_type, emoji, count, user_ids } =
                data as { message_id: number; message_type: 'room' | 'dm'; emoji: string; count: number; user_ids: number[] };

            if (message_type === 'room') {
                set(state => {
                    const newMsgsByRoom = { ...state.messagesByRoom };
                    for (const roomIdStr of Object.keys(newMsgsByRoom)) {
                        const roomId = Number(roomIdStr);
                        if (newMsgsByRoom[roomId]?.some(m => m.id === message_id)) {
                            newMsgsByRoom[roomId] = newMsgsByRoom[roomId].map(m =>
                                m.id === message_id ? { ...m, reactions: applyReaction(m.reactions, emoji, count, user_ids) } : m
                            );
                        }
                    }
                    return { messagesByRoom: newMsgsByRoom };
                });
            } else {
                set(state => {
                    const newDmMsgs = { ...state.messagesByDm };
                    for (const uid of Object.keys(newDmMsgs)) {
                        if (newDmMsgs[uid as any]?.some((m: any) => m.id === message_id)) {
                            newDmMsgs[uid as any] = newDmMsgs[uid as any].map((m: any) =>
                                m.id === message_id ? { ...m, reactions: applyReaction(m.reactions, emoji, count, user_ids) } : m
                            );
                        }
                    }
                    return { messagesByDm: newDmMsgs };
                });
            }
            break;
        }

        case 'error':
            console.error('WS logic error from server:', data);
            break;

        default:
            break;
    }
}
