import apiClient from '../../api/client';

export type InviteRoomPreview = {
    id: number;
    name: string;
    description?: string;
};

export type InviteLinkInfo = {
    max_uses: number;
    uses: number;
    expires_at?: string;
};

export type InviteLinkResponse = {
    room: InviteRoomPreview;
    link: InviteLinkInfo;
};

export const inviteApi = {
    getInvite: (code: string) => apiClient.get<InviteLinkResponse>(`/invite-links/${code}`),
    useInvite: (code: string) => apiClient.post(`/invite-links/${code}/use`),
};
