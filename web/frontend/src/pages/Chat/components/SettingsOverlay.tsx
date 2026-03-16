import { useEffect, useState } from 'react';
import { useAuthStore } from '../../../stores/authStore';
import { useChatStore } from '../../../stores/chatStore';
import { authApi } from '../../../services/api/auth';
import AccountSettings from './settings/AccountSettings';
import ProfileSettings from './settings/ProfileSettings';
import AppearanceSettings from './settings/AppearanceSettings';
import NotificationSettings from './settings/NotificationSettings';

interface SettingsOverlayProps {
    isOpen: boolean;
    onClose: () => void;
}

export default function SettingsOverlay({ isOpen, onClose }: SettingsOverlayProps) {
    const { clearAuth } = useAuthStore();
    const { disconnect } = useChatStore();
    const [activeTab, setActiveTab] = useState('account');

    // Handle escape key to close settings
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'Escape' && isOpen) {
                onClose();
            }
        };
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [isOpen, onClose]);

    if (!isOpen) return null;

    const handleLogout = async () => {
        disconnect();
        try {
            const rt = localStorage.getItem('refresh_token');
            await authApi.logout(rt ? { refresh_token: rt } : {});
        } catch { /* token might already be invalid */ }
        clearAuth();
        onClose();
    };

    return (
        <div className="fixed inset-0 z-50 flex bg-[var(--bg-main)] overflow-hidden animate-in fade-in duration-200">
            {/* Sidebar Left */}
            <div className="flex-1 flex justify-end bg-[var(--bg-secondary)] pb-10">
                <div className="w-56 mt-16 flex flex-col items-start px-2 py-4 space-y-1">
                    <div className="w-full text-xs font-bold text-[var(--text-muted)] mt-2 mb-1 px-3">用户设置</div>

                    <button
                        className={`w-full text-left px-3 py-1.5 rounded-md text-base ${activeTab === 'account' ? 'bg-[var(--bg-hover)] text-[var(--text-main)] font-semibold' : 'text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-main)] font-medium'}`}
                        onClick={() => setActiveTab('account')}
                    >
                        我的账号
                    </button>
                    <button
                        className={`w-full text-left px-3 py-1.5 rounded-md text-base ${activeTab === 'profile' ? 'bg-[var(--bg-hover)] text-[var(--text-main)] font-semibold' : 'text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-main)] font-medium'}`}
                        onClick={() => setActiveTab('profile')}
                    >
                        用户资料
                    </button>

                    <div className="w-full text-xs font-bold text-[var(--text-muted)] mt-5 mb-1 px-3">应用设置</div>

                    <button
                        className={`w-full text-left px-3 py-1.5 rounded-md text-base ${activeTab === 'appearance' ? 'bg-[var(--bg-hover)] text-[var(--text-main)] font-semibold' : 'text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-main)] font-medium'}`}
                        onClick={() => setActiveTab('appearance')}
                    >
                        外观
                    </button>
                    <button
                        className={`w-full text-left px-3 py-1.5 rounded-md text-base ${activeTab === 'notifications' ? 'bg-[var(--bg-hover)] text-[var(--text-main)] font-semibold' : 'text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-main)] font-medium'}`}
                        onClick={() => setActiveTab('notifications')}
                    >
                        通知
                    </button>

                    <div className="w-10/12 h-[1px] bg-[var(--text-muted)]/10 mx-auto my-3"></div>

                    <button
                        className={`w-full text-left px-3 py-1.5 rounded-md text-base text-[var(--color-discord-red-400)] hover:bg-[var(--color-discord-red-500)]/10 hover:text-[var(--color-discord-red-400)]`}
                        onClick={handleLogout}
                    >
                        退出登录
                    </button>
                </div>
            </div>

            {/* Main Content Right */}
            <div className="flex-[2_2_0%] bg-[var(--bg-main)] relative pb-10">
                <button
                    onClick={onClose}
                    className="absolute right-10 top-16 w-10 h-10 flex flex-col items-center justify-center text-[var(--text-muted)] hover:text-[var(--text-main)] transition-colors group z-10"
                >
                    <div className="w-8 h-8 rounded-full border-2 border-[var(--text-muted)] group-hover:bg-white/10 group-hover:border-[var(--text-main)] flex items-center justify-center mb-1 transition-all">
                        <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M18.4 4L12 10.4 5.6 4 4 5.6l6.4 6.4L4 18.4 5.6 20l6.4-6.4 6.4 6.4 1.6-1.6-6.4-6.4L20 5.6 18.4 4z" /></svg>
                    </div>
                    <span className="text-xs font-semibold">ESC</span>
                </button>

                <div className="max-w-[700px] mt-16 px-10 h-full overflow-y-auto custom-scrollbar pb-10">
                    {activeTab === 'account' && <AccountSettings />}
                    {activeTab === 'profile' && <ProfileSettings />}
                    {activeTab === 'appearance' && <AppearanceSettings />}
                    {activeTab === 'notifications' && <NotificationSettings />}
                </div>
            </div>
        </div>
    );
}
