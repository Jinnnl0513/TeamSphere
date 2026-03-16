import type { ChatSliceCreator, UiSlice } from './chatStore.types';

export const createUiSlice: ChatSliceCreator<UiSlice> = (set, get) => ({
    activeView: (localStorage.getItem('chat_activeView') as 'home' | 'rooms' | 'dm') || 'home',
    activeRoomId: localStorage.getItem('chat_activeRoomId')
        ? Number(localStorage.getItem('chat_activeRoomId'))
        : null,
    activeDmUserId: localStorage.getItem('chat_activeDmUserId')
        ? Number(localStorage.getItem('chat_activeDmUserId'))
        : null,
    myMutedUntilByRoom: {},
    typingUsers: {},
    lastReadByRoom: {},
    lastReadByDm: {},
    unreadCountsByRoom: {},
    mentionFilter: false,
    blockedUserIds: (() => {
        try {
            const raw = localStorage.getItem('chat_blockedUserIds');
            if (!raw) return [];
            const parsed = JSON.parse(raw);
            return Array.isArray(parsed) ? parsed.filter((id) => Number.isInteger(id)) : [];
        } catch {
            return [];
        }
    })(),

    setActiveView: (view) => {
        localStorage.setItem('chat_activeView', view);
        set({ activeView: view });
    },
    setActiveRoom: (roomId) => {
        if (roomId === null) {
            localStorage.removeItem('chat_activeRoomId');
            set({ activeRoomId: null });
            return;
        }

        if (get().activeRoomId !== roomId) {
            localStorage.setItem('chat_activeRoomId', String(roomId));
            localStorage.setItem('chat_activeView', 'rooms');
            localStorage.removeItem('chat_activeDmUserId');
            set({ activeRoomId: roomId, activeView: 'rooms', activeDmUserId: null });
            get().fetchHistory(roomId);
        } else {
            localStorage.setItem('chat_activeView', 'rooms');
            localStorage.removeItem('chat_activeDmUserId');
            set({ activeView: 'rooms', activeDmUserId: null });
        }
    },
    setActiveDmUser: (userId) => {
        if (userId === null) {
            localStorage.removeItem('chat_activeDmUserId');
            set({ activeDmUserId: null });
            return;
        }
        if (get().activeDmUserId !== userId) {
            localStorage.setItem('chat_activeDmUserId', String(userId));
            localStorage.setItem('chat_activeView', 'dm');
            localStorage.removeItem('chat_activeRoomId');
            set(state => ({
                activeDmUserId: userId,
                activeRoomId: null,
                activeView: 'dm',
                unreadDmCounts: { ...state.unreadDmCounts, [userId]: 0 }
            }));
            get().fetchDmHistory(userId);
        } else {
            localStorage.setItem('chat_activeView', 'dm');
            localStorage.removeItem('chat_activeRoomId');
            set(state => ({
                activeView: 'dm',
                activeRoomId: null,
                unreadDmCounts: { ...state.unreadDmCounts, [userId]: 0 }
            }));
        }
    },
    setMyMutedUntil: (roomId, mutedUntil) => {
        set(state => ({
            myMutedUntilByRoom: { ...state.myMutedUntilByRoom, [roomId]: mutedUntil }
        }));
    },

    markAsRead: (isDm, roomId, dmId, lastMsgId) => {
        if (isDm && dmId) {
            set(state => ({ lastReadByDm: { ...state.lastReadByDm, [dmId]: lastMsgId } }));
        } else if (!isDm && roomId) {
            set(state => ({ lastReadByRoom: { ...state.lastReadByRoom, [roomId]: lastMsgId } }));
        }
    },

    setUnreadCount: (roomId, count) => {
        set(state => ({ unreadCountsByRoom: { ...state.unreadCountsByRoom, [roomId]: count } }));
    },

    setMentionFilter: (val) => set({ mentionFilter: val }),

    blockUser: (userId) => {
        if (userId <= 0) return;
        if (userId === get().activeDmUserId) {
            // Allow blocking even if current DM, no-op if needed.
        }
        set(state => {
            if (state.blockedUserIds.includes(userId)) return state;
            const next = [...state.blockedUserIds, userId];
            localStorage.setItem('chat_blockedUserIds', JSON.stringify(next));
            return { blockedUserIds: next };
        });
    },
    unblockUser: (userId) => {
        set(state => {
            const next = state.blockedUserIds.filter(id => id !== userId);
            localStorage.setItem('chat_blockedUserIds', JSON.stringify(next));
            return { blockedUserIds: next };
        });
    },
});
