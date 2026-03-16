import type { StateCreator } from 'zustand';
import type { User, Room } from '../../types/models';
export type { User, Room };

export interface ReplyInfo {
    id: number;
    content: string;
    msg_type: string;
    is_deleted: boolean;
    user: User;
}

export interface ForwardInfo {
    type: 'room' | 'dm';
    id: number;
    msg_type: string;
    content: string;
    is_deleted: boolean;
    user: User;
}

export interface ReactionSummary {
    emoji: string;
    count: number;
    user_ids: number[];
}

export interface FileMeta {
    file_size: number;
    mime_type: string;
}

export interface ChatMessage {
    id: number;
    client_msg_id?: string;
    content: string;
    msg_type: string;
    user: User;
    room_id?: number;
    mentions?: number[];
    reply_to?: ReplyInfo;
    forward_meta?: ForwardInfo;
    reactions?: ReactionSummary[];
    file_size?: number;
    mime_type?: string;
    created_at: string;
    updated_at?: string;
    deleted_at?: string;
    read_at?: string;
}

export interface Conversation {
    user: User;
    last_message: string;
    last_message_at: string;
    unread_count?: number;
}

export interface SystemMessage {
    content: string;
    room_id?: number;
}

export interface ConnectionSlice {
    socket: WebSocket | null;
    isConnected: boolean;
    isConnecting: boolean;
    connect: () => Promise<void>;
    disconnect: () => void;
    _reset: () => void;
}

export interface UiSlice {
    activeView: 'home' | 'rooms' | 'dm';
    setActiveView: (view: 'home' | 'rooms' | 'dm') => void;
    activeRoomId: number | null;
    setActiveRoom: (roomId: number | null) => void;
    activeDmUserId: number | null;
    setActiveDmUser: (userId: number | null) => void;

    myMutedUntilByRoom: Record<number, string | null>;
    setMyMutedUntil: (roomId: number, mutedUntil: string | null) => void;

    typingUsers: Record<string, { username: string; expiresAt: number }[]>;

    // 未读分割线：记录每个房间/DM 进入时的最后已读消息 ID
    lastReadByRoom: Record<number, number>;
    lastReadByDm: Record<number, number>;
    // 标记当前进入时的最后已读位置
    markAsRead: (isDm: boolean, roomId: number | null, dmId: number | null, lastMsgId: number) => void;
    unreadCountsByRoom: Record<number, number>;
    setUnreadCount: (roomId: number, count: number) => void;

    mentionFilter: boolean;
    setMentionFilter: (val: boolean) => void;

    blockedUserIds: number[];
    blockUser: (userId: number) => void;
    unblockUser: (userId: number) => void;
}

export interface RoomSlice {
    rooms: Room[];
    fetchRooms: () => Promise<void>;

    messagesByRoom: Record<number, ChatMessage[]>;
    onlineUsersByRoom: Record<number, User[]>;

    fetchHistory: (roomId: number) => Promise<void>;
    fetchOlderHistory: (roomId: number) => Promise<void>;
    hasMoreByRoom: Record<number, boolean>;
    isLoadingOlderByRoom: Record<number, boolean>;

    sendMessage: (roomId: number, content: string, msgType?: string, replyToId?: number, fileMeta?: FileMeta) => void;
    sendTyping: (roomId: number) => void;
    joinRoom: (roomId: number) => void;
    leaveRoom: (roomId: number) => void;
    retractMessage: (msgId: number, roomId: number | null, dmUserId: number | null) => Promise<void>;
    editMessage: (msgId: number, content: string, roomId: number | null, dmUserId: number | null) => Promise<void>;
    setMessageReactions: (msgId: number, reactions: ReactionSummary[], isDm: boolean, dmUserId: number | null) => void;

    activeThreadMsgId: number | null;
    setActiveThreadMsgId: (msgId: number | null) => void;
    threadMessagesByMsgId: Record<number, ChatMessage[]>;
    fetchThreadMessages: (roomId: number, msgId: number) => Promise<void>;

    pinnedMessages: Record<number, ChatMessage[]>;
    fetchPinnedMessages: (roomId: number) => Promise<void>;
    pinMessage: (roomId: number, msgId: number) => Promise<void>;
    unpinMessage: (roomId: number, msgId: number) => Promise<void>;

    batchDeleteMessages: (roomId: number, msgIds: number[]) => Promise<void>;
}

export interface DmSlice {
    messagesByDm: Record<number, ChatMessage[]>;
    conversations: Conversation[];
    unreadDmCounts: Record<number, number>;

    fetchConversations: () => Promise<void>;
    fetchDmHistory: (userId: number) => Promise<void>;
    fetchOlderDmHistory: (userId: number) => Promise<void>;
    hasMoreByDm: Record<number, boolean>;
    isLoadingOlderByDm: Record<number, boolean>;
    clearUnreadDmCount: (userId: number) => void;

    sendDmMessage: (userId: number, content: string, msgType?: string, replyToId?: number, fileMeta?: FileMeta) => void;
    sendDmTyping: (userId: number) => void;
}

export type ChatStoreState = ConnectionSlice & UiSlice & RoomSlice & DmSlice;

export type ChatSliceCreator<T> = StateCreator<
    ChatStoreState,
    [],
    [],
    T
>;
