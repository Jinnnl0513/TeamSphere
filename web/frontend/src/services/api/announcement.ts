import apiClient from '../../api/client';

export type AnnouncementResponse = {
    content: string;
};

export const announcementApi = {
    getAnnouncement: () => apiClient.get<AnnouncementResponse>('/announcement'),
};
