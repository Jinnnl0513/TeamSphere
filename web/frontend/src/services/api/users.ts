import apiClient from '../../api/client';
import type { User } from '../../types/models';

export type UpdateProfilePayload = {
    bio?: string;
    profile_color?: string;
};

export type UpdatePasswordPayload = {
    old_password: string;
    new_password: string;
};

export const usersApi = {
    getMe: () => apiClient.get<User>('/users/me'),
    updateProfile: (payload: UpdateProfilePayload) => apiClient.put('/users/me/profile', payload),
    updatePassword: (payload: UpdatePasswordPayload) => apiClient.put('/users/me/password', payload),
    uploadAvatar: (file: File) => {
        const formData = new FormData();
        formData.append('file', file);
        return apiClient.post('/users/me/avatar', formData, {
            headers: {
                'Content-Type': 'multipart/form-data',
            },
        });
    },
    deleteMe: () => apiClient.delete('/users/me'),
    getProfile: (userId: number) => apiClient.get<User>(`/users/profile/${userId}`),
};
