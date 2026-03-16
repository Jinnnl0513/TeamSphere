import { create } from 'zustand';
import { friendsApi, type FriendInfo, type FriendRequest } from '../services/api/friends';

export type { FriendInfo };

export interface FriendStoreState {
    friends: FriendInfo[];
    pendingRequests: FriendRequest[];
    onlineFriendIds: Set<number>;
    isLoading: boolean;

    fetchFriends: () => Promise<void>;
    fetchRequests: () => Promise<void>;
    sendRequest: (username: string) => Promise<void>;
    respondRequest: (requestId: number, action: 'accept' | 'reject') => Promise<void>;
    removeFriend: (friendshipId: number) => Promise<void>;
    setFriendOnline: (userId: number) => void;
    setFriendOffline: (userId: number) => void;
    setOnlineFriendIds: (userIds: number[]) => void;
}

export const useFriendStore = create<FriendStoreState>((set, get) => ({
    friends: [],
    pendingRequests: [],
    onlineFriendIds: new Set(),
    isLoading: false,

    fetchFriends: async () => {
        set({ isLoading: true });
        try {
            const data = await friendsApi.listFriends();
            set({ friends: data || [] });
        } catch (err) {
            console.error('Failed to fetch friends', err);
        } finally {
            set({ isLoading: false });
        }
    },

    fetchRequests: async () => {
        try {
            const data = await friendsApi.listRequests();
            set({ pendingRequests: data || [] });
        } catch (err) {
            console.error('Failed to fetch friend requests', err);
        }
    },

    sendRequest: async (username: string) => {
        try {
            await friendsApi.sendRequest(username);
        } catch (err) {
            console.error('Failed to send friend request', err);
            throw err;
        }
    },

    respondRequest: async (requestId: number, action: 'accept' | 'reject') => {
        try {
            await friendsApi.respondRequest(requestId, action);
            await get().fetchRequests();
            if (action === 'accept') {
                await get().fetchFriends();
            }
        } catch (err) {
            console.error('Failed to respond to friend request', err);
            throw err;
        }
    },

    removeFriend: async (friendshipId: number) => {
        try {
            await friendsApi.removeFriend(friendshipId);
            await get().fetchFriends();
        } catch (err) {
            console.error('Failed to remove friend', err);
            throw err;
        }
    },

    setFriendOnline: (userId: number) => {
        set(state => {
            const next = new Set(state.onlineFriendIds);
            next.add(userId);
            return { onlineFriendIds: next };
        });
    },

    setFriendOffline: (userId: number) => {
        set(state => {
            const next = new Set(state.onlineFriendIds);
            next.delete(userId);
            return { onlineFriendIds: next };
        });
    },

    setOnlineFriendIds: (userIds: number[]) => {
        set({ onlineFriendIds: new Set(userIds) });
    },
}));
