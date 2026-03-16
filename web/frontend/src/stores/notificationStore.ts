import { create } from 'zustand';

interface NotificationState {
    desktopEnabled: boolean;
    soundEnabled: boolean;
    unreadBadge: boolean;
    permissionState: NotificationPermission | 'unsupported';

    requestPermission: () => Promise<void>;
    setDesktopEnabled: (v: boolean) => void;
    setSoundEnabled: (v: boolean) => void;
    setUnreadBadge: (v: boolean) => void;
    init: () => void;
}

export const useNotificationStore = create<NotificationState>((set, get) => ({
    desktopEnabled: localStorage.getItem('notify-desktop') !== 'false',
    soundEnabled: localStorage.getItem('notify-sound') !== 'false',
    unreadBadge: localStorage.getItem('notify-badge') !== 'false',
    permissionState: 'default',

    init: () => {
        if (!('Notification' in window)) {
            set({ permissionState: 'unsupported' });
        } else {
            set({ permissionState: Notification.permission });
        }
    },

    requestPermission: async () => {
        if (!('Notification' in window)) {
            set({ permissionState: 'unsupported' });
            return;
        }
        const permission = await Notification.requestPermission();
        set({ permissionState: permission });
        if (permission === 'granted') {
            set({ desktopEnabled: true });
            localStorage.setItem('notify-desktop', 'true');
        }
    },

    setDesktopEnabled: (v) => {
        localStorage.setItem('notify-desktop', String(v));
        set({ desktopEnabled: v });
        if (v && 'Notification' in window && Notification.permission === 'default') {
            get().requestPermission();
        }
    },

    setSoundEnabled: (v) => {
        localStorage.setItem('notify-sound', String(v));
        set({ soundEnabled: v });
    },

    setUnreadBadge: (v) => {
        localStorage.setItem('notify-badge', String(v));
        set({ unreadBadge: v });
    },
}));

/** Trigger a desktop notification if permission is granted and enabled. */
export function sendDesktopNotification(title: string, body: string, icon?: string) {
    const { desktopEnabled } = useNotificationStore.getState();

    // 如果没有开启桌面通知设置、浏览器不支持，或未授权则不发送
    if (!desktopEnabled || typeof window === 'undefined' || !('Notification' in window) || Notification.permission !== 'granted') {
        return;
    }

    try {
        const notification = new Notification(title, {
            body,
            icon: icon || '/favicon.ico'
        });

        notification.onclick = () => {
            window.focus();
            notification.close();
        };
    } catch (err) {
        console.error('Failed to send desktop notification:', err);
    }
}
