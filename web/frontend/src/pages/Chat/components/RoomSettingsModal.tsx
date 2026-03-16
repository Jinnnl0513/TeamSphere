import { useState, useEffect } from 'react';
import type { RoomSettingsModalProps, TabKey } from './room-settings/types';
import OverviewTab from './room-settings/tabs/OverviewTab';
import AccessTab from './room-settings/tabs/AccessTab';
import PermissionsTab from './room-settings/tabs/PermissionsTab';
import MessagesTab from './room-settings/tabs/MessagesTab';
import SafetyTab from './room-settings/tabs/SafetyTab';
import JoinRequestsTab from './room-settings/tabs/JoinRequestsTab';
import InvitesTab from './room-settings/tabs/InvitesTab';
import StatsTab from './room-settings/tabs/StatsTab';

export default function RoomSettingsModal({ isOpen, onClose, roomId, roomName, myRole }: RoomSettingsModalProps) {
    const [activeTab, setActiveTab] = useState<TabKey>('overview');

    const isOwner = myRole === 'owner';
    const canManageSettings = myRole === 'owner' || myRole === 'admin';
    const canManagePermissions = myRole === 'owner';
    const canManageMembers = myRole === 'owner' || myRole === 'admin';

    useEffect(() => {
        if (isOpen) {
            setActiveTab('overview');
        }
    }, [isOpen]);

    const tabs: { key: TabKey; label: string }[] = [
        { key: 'overview', label: '基础信息' },
        { key: 'access', label: '加入与可见性' },
        { key: 'permissions', label: '角色权限' },
        { key: 'messages', label: '消息与通知' },
        { key: 'safety', label: '安全与风控' },
        { key: 'join', label: '加入申请' },
        { key: 'invites', label: '邀请与链接' },
        { key: 'stats', label: '数据统计' },
    ];

    if (!isOpen) return null;

    const renderTabContent = () => {
        const props = {
            roomId,
            canManageSettings,
            canManageMembers,
            canManagePermissions,
            isOwner,
        };

        switch (activeTab) {
            case 'overview':
                return <OverviewTab {...props} onClose={onClose} />;
            case 'access':
                return <AccessTab {...props} />;
            case 'permissions':
                return <PermissionsTab {...props} />;
            case 'messages':
                return <MessagesTab {...props} />;
            case 'safety':
                return <SafetyTab {...props} />;
            case 'join':
                return <JoinRequestsTab {...props} />;
            case 'invites':
                return <InvitesTab {...props} />;
            case 'stats':
                return <StatsTab {...props} />;
            default:
                return null;
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
            <div className="bg-[var(--bg-secondary)] w-full max-w-5xl rounded-2xl shadow-2xl overflow-hidden border border-[var(--bg-sidebar)] flex flex-col">
                <div className="px-6 py-5 border-b border-[var(--bg-sidebar)] font-bold text-xl text-[var(--text-main)] flex justify-between items-center">
                    <div className="flex items-center gap-2">
                        <span>频道设置</span>
                        <span className="text-xs text-[var(--text-muted)]">{roomName}</span>
                    </div>
                    <button
                        onClick={onClose}
                        className="text-[var(--text-muted)] hover:text-[#ff6b6b] transition-colors p-1"
                    >
                        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                    </button>
                </div>

                <div className="flex h-[70vh]">
                    <aside className="w-56 border-r border-[var(--bg-sidebar)] p-4 space-y-2 overflow-y-auto">
                        {tabs.map(tab => (
                            <button
                                key={tab.key}
                                onClick={() => setActiveTab(tab.key)}
                                className={`w-full text-left px-3 py-2 rounded-md text-sm font-medium transition-colors ${activeTab === tab.key ? 'bg-[var(--bg-hover)] text-[var(--text-main)]' : 'text-[var(--text-muted)] hover:bg-[var(--bg-hover)]'}`}
                            >
                                {tab.label}
                            </button>
                        ))}
                    </aside>

                    <div className="flex-1 p-6 overflow-y-auto custom-scrollbar relative">
                        {renderTabContent()}
                    </div>
                </div>
            </div>
        </div>
    );
}
