import type { RoomSettings, RoomRolePermission } from '../../../../services/api/rooms';

export interface RoomSettingsModalProps {
    isOpen: boolean;
    onClose: () => void;
    roomId: number;
    roomName: string;
    roomDescription: string;
    myRole: 'owner' | 'admin' | 'member' | string;
}

export type TabKey =
    | 'overview'
    | 'access'
    | 'permissions'
    | 'messages'
    | 'safety'
    | 'join'
    | 'invites'
    | 'stats';

export interface BaseTabProps {
    roomId: number;
    canManageSettings: boolean;
    canManageMembers: boolean;
    canManagePermissions: boolean;
    isOwner: boolean;
}

// 通用的文本多行列表和字符串互转帮助函数
export const listToText = (value?: string[] | null) => {
    if (!value || value.length === 0) return '';
    return value.join('\n');
};

export const parseList = (value: string) => {
    const items = value
        .split(/\r?\n|,/g)
        .map(v => v.trim())
        .filter(Boolean);
    return items.length > 0 ? items : [];
};

export const defaultSettings = (): RoomSettings => ({
    is_public: true,
    require_approval: false,
    read_only: false,
    archived: false,
    topic: '',
    avatar_url: '',
    slow_mode_seconds: 0,
    message_retention_days: 0,
    content_filter_mode: 'off',
    blocked_keywords: [],
    allowed_link_domains: [],
    blocked_link_domains: [],
    allowed_file_types: [],
    max_file_size_mb: 10,
    pin_limit: 50,
    notify_mode: 'all',
    notify_keywords: [],
    dnd_start: null,
    dnd_end: null,
    anti_spam_rate: 8,
    anti_spam_window_sec: 10,
    anti_repeat: true,
    stats_enabled: true,
});

export const defaultPermissions = (): RoomRolePermission[] => ([
    { role: 'owner', can_send: true, can_upload: true, can_pin: true, can_manage_members: true, can_manage_settings: true, can_manage_messages: true, can_mention_all: true },
    { role: 'admin', can_send: true, can_upload: true, can_pin: true, can_manage_members: true, can_manage_settings: true, can_manage_messages: true, can_mention_all: true },
    { role: 'member', can_send: true, can_upload: true, can_pin: true, can_manage_members: false, can_manage_settings: false, can_manage_messages: false, can_mention_all: false },
]);
