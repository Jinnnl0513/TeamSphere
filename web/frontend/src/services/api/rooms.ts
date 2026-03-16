import apiClient from '../../api/client';
import type { Room, User } from '../../types/models';

export type RoomMember = {
  user_id: number;
  room_id: number;
  role: 'owner' | 'admin' | 'member';
  user: User;
  muted_until?: string | null;
};

export type RoomInvite = {
  id: number;
  room?: { name?: string };
  inviter?: { username?: string };
  room_name?: string;
  inviter_username?: string;
};

export type InviteLink = {
  id: number;
  code: string;
  max_uses: number;
  uses: number;
  expires_at?: string;
  created_at: string;
  creator_name: string;
};

export type RoomSettings = {
  is_public: boolean;
  require_approval: boolean;
  read_only: boolean;
  archived: boolean;
  topic?: string | null;
  avatar_url?: string | null;
  slow_mode_seconds: number;
  message_retention_days: number;
  content_filter_mode: 'off' | 'block_log';
  blocked_keywords?: string[] | null;
  allowed_link_domains?: string[] | null;
  blocked_link_domains?: string[] | null;
  allowed_file_types?: string[] | null;
  max_file_size_mb: number;
  pin_limit: number;
  notify_mode: 'all' | 'mentions' | 'none';
  notify_keywords?: string[] | null;
  dnd_start?: string | null;
  dnd_end?: string | null;
  anti_spam_rate: number;
  anti_spam_window_sec: number;
  anti_repeat: boolean;
  stats_enabled: boolean;
};

export type RoomRolePermission = {
  role: 'owner' | 'admin' | 'member';
  can_send: boolean;
  can_upload: boolean;
  can_pin: boolean;
  can_manage_members: boolean;
  can_manage_settings: boolean;
  can_manage_messages: boolean;
  can_mention_all: boolean;
};

export type RoomJoinRequest = {
  id: number;
  room_id: number;
  user_id: number;
  status: 'pending' | 'approved' | 'rejected';
  reason?: string | null;
  reviewer_id?: number | null;
  created_at: string;
  updated_at: string;
};

export type RoomStatsSummary = {
  total_messages: number;
  active_users: number;
};

export const roomsApi = {
  list: () => apiClient.get<Room[]>('/rooms'),
  discover: () => apiClient.get<Room[]>('/rooms/discover'),
  create: (payload: { name: string; description?: string }) => apiClient.post<Room>('/rooms', payload),
  join: (roomId: number) => apiClient.post(`/rooms/${roomId}/join`),
  leave: (roomId: number) => apiClient.post(`/rooms/${roomId}/leave`),
  update: (roomId: number, payload: { name: string; description?: string }) =>
    apiClient.put(`/rooms/${roomId}`, payload),
  remove: (roomId: number) => apiClient.delete(`/rooms/${roomId}`),
  listMembers: (roomId: number) => apiClient.get<RoomMember[]>(`/rooms/${roomId}/members`),
  invite: (roomId: number, userId: number) => apiClient.post(`/rooms/${roomId}/invite`, { user_id: userId }),
  listInvites: () => apiClient.get<RoomInvite[]>('/rooms/invites'),
  respondInvite: (inviteId: number, action: 'accept' | 'decline') =>
    apiClient.put(`/rooms/invites/${inviteId}`, { action }),
  transferOwner: (roomId: number, newOwnerId: number) =>
    apiClient.put(`/rooms/${roomId}/transfer`, { new_owner_id: newOwnerId }),
  kickMember: (roomId: number, userId: number) => apiClient.delete(`/rooms/${roomId}/members/${userId}`),
  muteMember: (roomId: number, userId: number, duration: number) =>
    apiClient.post(`/rooms/${roomId}/members/${userId}/mute`, { duration }),
  unmuteMember: (roomId: number, userId: number) => apiClient.delete(`/rooms/${roomId}/members/${userId}/mute`),
  updateMemberRole: (roomId: number, userId: number, role: 'admin' | 'member') =>
    apiClient.put(`/rooms/${roomId}/members/${userId}`, { role }),
  listInviteLinks: (roomId: number) => apiClient.get<InviteLink[]>(`/rooms/${roomId}/invite-links`),
  createInviteLink: (roomId: number, payload: { max_uses: number; expires_hours: number }) =>
    apiClient.post(`/rooms/${roomId}/invite-links`, payload),
  deleteInviteLink: (roomId: number, linkId: number) =>
    apiClient.delete(`/rooms/${roomId}/invite-links/${linkId}`),
  markRead: (roomId: number, lastReadMsgId?: number) =>
    apiClient.post(`/rooms/${roomId}/read`, lastReadMsgId ? { last_read_msg_id: lastReadMsgId } : {}),
  getUnreadCount: (roomId: number) =>
    apiClient.get<{ unread_count: number }>(`/rooms/${roomId}/unread-count`),
  getSettings: (roomId: number) => apiClient.get<RoomSettings>(`/rooms/${roomId}/settings`),
  updateSettings: (roomId: number, settings: RoomSettings) =>
    apiClient.put(`/rooms/${roomId}/settings`, { settings }),
  getPermissions: (roomId: number) => apiClient.get<RoomRolePermission[]>(`/rooms/${roomId}/permissions`),
  updatePermissions: (roomId: number, permissions: RoomRolePermission[]) =>
    apiClient.put(`/rooms/${roomId}/permissions`, { permissions }),
  listJoinRequests: (roomId: number) => apiClient.get<RoomJoinRequest[]>(`/rooms/${roomId}/join-requests`),
  approveJoinRequest: (roomId: number, reqId: number) =>
    apiClient.post(`/rooms/${roomId}/join-requests/${reqId}/approve`),
  rejectJoinRequest: (roomId: number, reqId: number) =>
    apiClient.post(`/rooms/${roomId}/join-requests/${reqId}/reject`),
  getStatsSummary: (roomId: number, days = 7) =>
    apiClient.get<RoomStatsSummary>(`/rooms/${roomId}/stats/summary?days=${days}`),
};
