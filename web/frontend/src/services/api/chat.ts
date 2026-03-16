import apiClient from '../../api/client';
import type { ChatMessage, Conversation, ReactionSummary } from '../../stores/chat/chatStore.types';

export const chatApi = {
    getDmConversations: () => apiClient.get<Conversation[]>('/dm/conversations'),
    getDmMessages: (userId: number, limit: number) =>
        apiClient.get<ChatMessage[]>(`/dm/${userId}/messages?limit=${limit}`),
    getDmMessagesBefore: (userId: number, beforeId: number, limit: number) =>
        apiClient.get<ChatMessage[]>(`/dm/${userId}/messages?before_id=${beforeId}&limit=${limit}`),
    markDmRead: (userId: number, lastReadMsgId: number) =>
        apiClient.post(`/dm/${userId}/read`, { last_read_msg_id: lastReadMsgId }),
    deleteDmMessage: (msgId: number) => apiClient.delete(`/dm/messages/${msgId}`),
    updateDmMessage: (msgId: number, content: string) =>
        apiClient.put(`/dm/messages/${msgId}`, { content }),

    getRoomMessages: (roomId: number, limit: number) =>
        apiClient.get<ChatMessage[]>(`/rooms/${roomId}/messages?limit=${limit}`),
    getRoomMessagesBefore: (roomId: number, beforeId: number, limit: number) =>
        apiClient.get<ChatMessage[]>(`/rooms/${roomId}/messages?before_id=${beforeId}&limit=${limit}`),
    deleteRoomMessage: (roomId: number, msgId: number) =>
        apiClient.delete(`/rooms/${roomId}/messages/${msgId}`),
    deleteRoomMessagesBatch: (roomId: number, msgIds: number[]) =>
        apiClient.delete(`/rooms/${roomId}/messages/batch`, { data: { msg_ids: msgIds } }),
    updateRoomMessage: (roomId: number, msgId: number, content: string) =>
        apiClient.put(`/rooms/${roomId}/messages/${msgId}`, { content }),
    getThreadMessages: (roomId: number, msgId: number, limit = 50) =>
        apiClient.get<ChatMessage[]>(`/rooms/${roomId}/messages/${msgId}/thread?limit=${limit}`),
    getPinnedMessages: (roomId: number) =>
        apiClient.get<ChatMessage[]>(`/rooms/${roomId}/pinned-messages`),
    pinMessage: (roomId: number, msgId: number) =>
        apiClient.post(`/rooms/${roomId}/messages/${msgId}/pin`),
    unpinMessage: (roomId: number, msgId: number) =>
        apiClient.delete(`/rooms/${roomId}/messages/${msgId}/pin`),
    searchMessages: (params: { q: string; roomId: number; senderId?: number; from?: string; to?: string; limit?: number }) => {
        const sp = new URLSearchParams();
        sp.set('q', params.q);
        sp.set('room_id', String(params.roomId));
        if (params.senderId) sp.set('sender_id', String(params.senderId));
        if (params.from) sp.set('from', params.from);
        if (params.to) sp.set('to', params.to);
        if (params.limit) sp.set('limit', String(params.limit));
        return apiClient.get<ChatMessage[]>(`/search/messages?${sp.toString()}`);
    },

    // Reactions
    addReaction: (msgId: number, emoji: string, messageType: 'room' | 'dm', roomId?: number, peerUserId?: number) =>
        apiClient.post<ReactionSummary[]>(`/messages/${msgId}/reactions`, {
            emoji,
            message_type: messageType,
            room_id: roomId ?? 0,
            peer_user_id: peerUserId ?? 0,
        }),
    removeReaction: (msgId: number, emoji: string, messageType: 'room' | 'dm', roomId?: number, peerUserId?: number) =>
        apiClient.delete<ReactionSummary[]>(`/messages/${msgId}/reactions/${encodeURIComponent(emoji)}`, {
            data: {
                message_type: messageType,
                room_id: roomId ?? 0,
                peer_user_id: peerUserId ?? 0,
            },
        }),
    getReactions: (msgId: number, messageType: 'room' | 'dm') =>
        apiClient.get<ReactionSummary[]>(`/messages/${msgId}/reactions?type=${messageType}`),

    forwardMessage: (msgId: number, payload: { message_type: 'room' | 'dm'; target_room_id?: number; target_user_id?: number; comment?: string }) =>
        apiClient.post(`/messages/${msgId}/forward`, payload),
};
