import apiClient from '../../api/client';

export type AdminUser = {
  id: number;
  username: string;
  avatar_url: string;
  bio: string;
  profile_color: string;
  role: string;
  deleted_at: unknown;
  created_at: string;
};

export type UsersResponse = {
  users: AdminUser[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
};

export type AdminRoom = {
  id: number;
  name: string;
  description: string;
  creator_id: number;
  created_at: string;
  member_count: number;
};

export type RoomsResponse = {
  rooms: AdminRoom[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
};

export type Stats = {
  total_users: number;
  active_users: number;
  total_rooms: number;
  total_messages: number;
  total_dms: number;
  online_users: number;
};

export type SettingsData = {
  setup_completed: string;
  [key: string]: string;
};

export type EmailSettings = {
  enabled: boolean;
  smtp_host: string;
  smtp_port: number;
  username: string;
  password?: string;
  from_address: string;
  from_name: string;
};

export type AnnouncementResponse = {
  content: string;
};

export const adminApi = {
  getStats: () => apiClient.get<Stats>('/admin/stats'),
  listUsers: (page = 1, pageSize = 20) =>
    apiClient.get<UsersResponse>(`/admin/users?page=${page}&page_size=${pageSize}`),
  updateUserRole: (userId: number, role: string) =>
    apiClient.put(`/admin/users/${userId}/role`, { role }),
  deleteUser: (userId: number) => apiClient.delete(`/admin/users/${userId}`),
  listRooms: (page = 1, pageSize = 20) =>
    apiClient.get<RoomsResponse>(`/admin/rooms?page=${page}&page_size=${pageSize}`),
  deleteRoom: (roomId: number) => apiClient.delete(`/admin/rooms/${roomId}`),
  getSettings: () => apiClient.get<SettingsData>('/admin/settings'),
  updateSettings: (payload: Record<string, string>) =>
    apiClient.put('/admin/settings', { settings: payload }),
  getEmailSettings: () => apiClient.get<EmailSettings>('/admin/email'),
  updateEmailSettings: (payload: EmailSettings) => apiClient.put('/admin/email', payload),
  getAnnouncement: () => apiClient.get<AnnouncementResponse>('/admin/announcement'),
  setAnnouncement: (content: string) => apiClient.post('/admin/announcement', { content }),
};
