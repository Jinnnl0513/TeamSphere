import apiClient from '../../api/client';
import type { User } from '../../types/models';

export interface FriendInfo {
    friendship_id: number;
    user: User;
}

export interface FriendRequest {
    id: number;
    user_id: number;
    friend_id: number;
    status: string;
    created_at: string;
    from: User;
}

export const friendsApi = {
    listFriends: () => apiClient.get<FriendInfo[]>('/friends'),
    listRequests: () => apiClient.get<FriendRequest[]>('/friends/requests'),
    sendRequest: (username: string) => apiClient.post('/friends/request', { username }),
    respondRequest: (requestId: number, action: 'accept' | 'reject') =>
        apiClient.put(`/friends/requests/${requestId}`, { action }),
    removeFriend: (friendshipId: number) => apiClient.delete(`/friends/${friendshipId}`),
};
