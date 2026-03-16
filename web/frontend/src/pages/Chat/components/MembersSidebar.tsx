import { useEffect, useState, useRef, useCallback } from 'react';
import { useChatStore } from '../../../stores/chatStore';
import { useAuthStore } from '../../../stores/authStore';
import Swal from 'sweetalert2';
import UserProfileModal from './UserProfileModal';
import type { User } from '../../../types/models';
import { roomsApi, type RoomMember } from '../../../services/api/rooms';

export default function MembersSidebar({ isVisible }: { isVisible: boolean }) {
    const { activeRoomId, onlineUsersByRoom } = useChatStore();
    const { user: currentUser } = useAuthStore();
    const [members, setMembers] = useState<RoomMember[]>([]);
    const [loading, setLoading] = useState(false);
    const [selectedUser, setSelectedUser] = useState<User | null>(null);
    const [nowTs, setNowTs] = useState(() => Date.now());

    const fetchMembers = useCallback(async () => {
        if (!activeRoomId) return;
        setLoading(true);
        try {
            const res = await roomsApi.listMembers(activeRoomId);
            setMembers(res || []);
        } catch (err) {
            console.error("Failed to fetch members", err);
        } finally {
            setLoading(false);
        }
    }, [activeRoomId]);

    useEffect(() => {
        if (!activeRoomId || !isVisible) {
            setMembers([]);
            return;
        }
        void fetchMembers();
    }, [activeRoomId, isVisible, fetchMembers]);

    useEffect(() => {
        const timer = window.setInterval(() => setNowTs(Date.now()), 30 * 1000);
        return () => window.clearInterval(timer);
    }, []);

    const handleAction = async (action: string, targetId: number) => {
        if (!activeRoomId) {
            return;
        }
        try {
            switch (action) {
                case 'kick': {
                    const result = await Swal.fire({
                        title: '确定踢出此人？',
                        icon: 'warning',
                        showCancelButton: true,
                        confirmButtonColor: '#d33',
                        cancelButtonColor: '#3085d6',
                        confirmButtonText: '确定踢出',
                        cancelButtonText: '取消'
                    });
                    if (result.isConfirmed) {
                        await roomsApi.kickMember(activeRoomId, targetId);
                    }
                    break;
                }
                case 'mute': {
                    const { value: durationStr } = await Swal.fire({
                        title: '禁言时长',
                        input: 'text',
                        inputLabel: '请输入禁言时长（分钟）：',
                        inputValue: '60',
                        showCancelButton: true,
                        confirmButtonText: '确定',
                        cancelButtonText: '取消',
                        inputValidator: (value: string) => {
                            if (!value) {
                                return '请输入时长';
                            }
                            const duration = parseInt(value, 10);
                            if (isNaN(duration) || duration <= 0) {
                                return '禁言时长无效，请输入大于 0 的整数数字（代表分钟）。';
                            }
                            return null;
                        }
                    });

                    if (durationStr) {
                        const duration = parseInt(durationStr as string, 10);
                        const result = await Swal.fire({
                            title: '确认禁言?',
                            text: `确定要禁言此人 ${duration} 分钟吗？`,
                            icon: 'warning',
                            showCancelButton: true,
                            confirmButtonColor: '#d33',
                            cancelButtonColor: '#3085d6',
                            confirmButtonText: '确定禁言',
                            cancelButtonText: '取消'
                        });

                        if (result.isConfirmed) {
                            await roomsApi.muteMember(activeRoomId, targetId, duration);
                        }
                    }
                    break;
                }
                case 'unmute':
                    await roomsApi.unmuteMember(activeRoomId, targetId);
                    break;
                case 'set_admin':
                    await roomsApi.updateMemberRole(activeRoomId, targetId, 'admin');
                    break;
                case 'set_member':
                    await roomsApi.updateMemberRole(activeRoomId, targetId, 'member');
                    break;
                case 'transfer': {
                    const result = await Swal.fire({
                        title: '转移群主?',
                        text: '确定将频道群主转移给此人吗？您将成为普通管理员。',
                        icon: 'warning',
                        showCancelButton: true,
                        confirmButtonColor: '#d33',
                        cancelButtonColor: '#3085d6',
                        confirmButtonText: '确定转移',
                        cancelButtonText: '取消'
                    });
                    if (result.isConfirmed) {
                        // Fix #12: Use the dedicated /transfer endpoint, not UpdateMemberRole.
                        // UpdateMemberRole rejects 'owner' role explicitly.
                        await roomsApi.transferOwner(activeRoomId, targetId);
                    }
                    break;
                }
            }
            await fetchMembers();
        } catch (err: any) {
            Swal.fire('操作失败', err.response?.data?.message || err.message, 'error');
        }
    };

    if (!isVisible) return null;

    const myRole = members.find(m => m.user_id === currentUser?.id)?.role || 'member';

    // Group members by role
    const onlineUsers = activeRoomId ? (onlineUsersByRoom[activeRoomId] || []) : [];
    const onlineSet = new Set(onlineUsers.map(u => u.id));

    const owners = members.filter(m => m.role === 'owner');
    const admins = members.filter(m => m.role === 'admin');
    const normals = members.filter(m => m.role === 'member');

    const onlineNormals = normals.filter(m => onlineSet.has(m.user_id));
    const offlineNormals = normals.filter(m => !onlineSet.has(m.user_id));

    const renderItem = (m: RoomMember) => (
        <MemberItem
            key={m.user_id}
            member={m}
            isOnline={onlineSet.has(m.user_id)}
            myRole={myRole}
            currentUserId={currentUser?.id}
            onAction={handleAction}
            nowTs={nowTs}
            onViewProfile={(u) => setSelectedUser(u)}
        />
    );

    return (
        <div className="w-[240px] bg-[var(--bg-secondary)] flex-shrink-0 flex flex-col z-10 border-l border-[var(--bg-sidebar)] overflow-y-auto custom-scrollbar">
            {loading && (
                <div className="flex h-12 items-center justify-center text-xs text-[var(--text-muted)] p-4">
                    加载成员中...
                </div>
            )}

            {!loading && members.length === 0 && (
                <div className="p-4 text-center text-sm text-[var(--text-muted)] italic">
                    暂无成员信息
                </div>
            )}

            {!loading && members.length > 0 && (
                <div className="py-4 space-y-4">
                    {owners.length > 0 && (
                        <div className="px-2">
                            <h3 className="text-xs font-semibold text-[var(--text-muted)] uppercase px-2 mb-1">
                                群主 — {owners.length}
                            </h3>
                            <div className="space-y-[2px]">
                                {owners.map(renderItem)}
                            </div>
                        </div>
                    )}

                    {admins.length > 0 && (
                        <div className="px-2">
                            <h3 className="text-xs font-semibold text-[var(--text-muted)] uppercase px-2 mb-1">
                                管理员 — {admins.length}
                            </h3>
                            <div className="space-y-[2px]">
                                {admins.map(renderItem)}
                            </div>
                        </div>
                    )}

                    {onlineNormals.length > 0 && (
                        <div className="px-2">
                            <h3 className="text-xs font-semibold text-[var(--text-muted)] uppercase px-2 mb-1">
                                在线 — {onlineNormals.length}
                            </h3>
                            <div className="space-y-[2px]">
                                {onlineNormals.map(renderItem)}
                            </div>
                        </div>
                    )}

                    {offlineNormals.length > 0 && (
                        <div className="px-2">
                            <h3 className="text-xs font-semibold text-[var(--text-muted)] uppercase px-2 mb-1">
                                离线 — {offlineNormals.length}
                            </h3>
                            <div className="space-y-[2px]">
                                {offlineNormals.map(renderItem)}
                            </div>
                        </div>
                    )}
                </div>
            )}

            <UserProfileModal
                user={selectedUser}
                isOpen={!!selectedUser}
                onClose={() => setSelectedUser(null)}
            />
        </div>
    );
}

function MemberItem({ member, isOnline, myRole, currentUserId, onAction, nowTs, onViewProfile }: {
    member: RoomMember,
    isOnline: boolean,
    myRole: string,
    currentUserId?: number,
    onAction: (action: string, id: number) => void,
    nowTs: number,
    onViewProfile: (user: User) => void
}) {
    const [showMenu, setShowMenu] = useState(false);
    const menuRef = useRef<HTMLDivElement>(null);

    const isSelf = member.user_id === currentUserId;
    const canManage = !isSelf && (myRole === 'owner' || (myRole === 'admin' && member.role === 'member'));
    const canSetAdmin = myRole === 'owner' && !isSelf;
    const isMuted = member.muted_until && new Date(member.muted_until).getTime() > nowTs;

    // click outside to close
    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
                setShowMenu(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    return (
        <div className="relative" ref={menuRef}>
            <div
                onClick={() => setShowMenu(!showMenu)}
                onContextMenu={(e) => {
                    e.preventDefault();
                    setShowMenu(true);
                }}
                className={`flex items-center space-x-2 px-2 py-1.5 rounded-md hover:bg-[var(--bg-main)]/60 cursor-pointer transition-colors duration-150 group ${!isOnline && 'opacity-50 hover:opacity-100'} ${showMenu ? 'bg-[var(--bg-main)]/80' : ''}`}
            >
                <div className="w-8 h-8 rounded-full bg-[var(--accent)] flex-shrink-0 flex items-center justify-center text-white font-bold relative">
                    <img src={member.user.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${member.user.username}`} alt="avatar" className="w-full h-full rounded-full object-cover" />
                    {isOnline && (
                        <div className="absolute bottom-0 -right-0.5 w-3.5 h-3.5 bg-[var(--color-discord-green-500)] border-[2.5px] border-[var(--bg-secondary)] group-hover:border-[#3b414a] rounded-full z-10 transition-colors"></div>
                    )}
                </div>
                <div className="flex-1 truncate min-w-0 flex items-center">
                    <div className="text-[15px] font-medium text-[var(--text-main)] truncate leading-tight flex-1 flex flex-wrap items-center">
                        {member.user.username}
                        {isSelf && <span className="text-xs text-[var(--text-muted)] ml-1">(您)</span>}
                        {isMuted && <span className="text-xs mx-1 cursor-default text-[#ff6b6b]" title={`被禁言至 ${new Date(member.muted_until!).toLocaleString()}`}>🔇</span>}
                    </div>
                </div>
            </div>

            {/* Menu */}
            {showMenu && (
                <div className="absolute top-10 right-4 w-40 bg-[var(--bg-main)] border border-[var(--bg-sidebar)] rounded-md shadow-lg py-1 z-50">
                    <div className="px-3 py-1 text-xs font-bold text-[var(--text-muted)] border-b border-[var(--bg-sidebar)] mb-1">
                        对 {member.user.username} 操作
                    </div>
                    <button onClick={() => { onViewProfile(member.user); setShowMenu(false); }} className="w-full text-left px-3 py-1.5 text-sm hover:bg-[var(--bg-secondary)]">
                        👤 查看资料
                    </button>
                    {canManage && (
                        <>
                            {isMuted ? (
                                <button onClick={() => { onAction('unmute', member.user_id); setShowMenu(false); }} className="w-full text-left px-3 py-1.5 text-sm hover:bg-[var(--bg-secondary)]">
                                    🔊 解除禁言
                                </button>
                            ) : (
                                <button onClick={() => { onAction('mute', member.user_id); setShowMenu(false); }} className="w-full text-left px-3 py-1.5 text-sm hover:bg-[var(--bg-secondary)]">
                                    🤐 禁言
                                </button>
                            )}
                            <button onClick={() => { onAction('kick', member.user_id); setShowMenu(false); }} className="w-full text-left px-3 py-1.5 text-sm hover:bg-[#ff6b6b]/10 text-[#ff6b6b]">
                                👢 移出频道
                            </button>
                        </>
                    )}

                    {canSetAdmin && (
                        <div className="border-t border-[var(--bg-sidebar)] mt-1 pt-1">
                            {member.role === 'member' && (
                                <button onClick={() => { onAction('set_admin', member.user_id); setShowMenu(false); }} className="w-full text-left px-3 py-1.5 text-sm hover:bg-[var(--bg-secondary)] text-[#20c997]">
                                    👑 设为管理员
                                </button>
                            )}
                            {member.role === 'admin' && (
                                <button onClick={() => { onAction('set_member', member.user_id); setShowMenu(false); }} className="w-full text-left px-3 py-1.5 text-sm hover:bg-[var(--bg-secondary)] text-[#f39c12]">
                                    ⬇️ 撤销管理员
                                </button>
                            )}
                            <button onClick={() => { onAction('transfer', member.user_id); setShowMenu(false); }} className="w-full text-left px-3 py-1.5 text-sm hover:bg-[#ff6b6b]/10 text-[#ff6b6b]">
                                ⚠️ 转移群主
                            </button>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

