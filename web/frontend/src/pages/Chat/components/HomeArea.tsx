import { useEffect, useState, useRef, useCallback } from 'react';
import { useFriendStore, type FriendInfo } from '../../../stores/friendStore';
import { useChatStore } from '../../../stores/chatStore';
import { Users } from 'lucide-react';
import UserProfileModal from './UserProfileModal';
import type { User } from '../../../types/models';
import { useNavigate } from 'react-router-dom';
import { roomsApi, type RoomInvite } from '../../../services/api/rooms';

type FriendTab = 'online' | 'all' | 'pending' | 'add' | 'invites';

type RoomInviteItem = RoomInvite;

function normalizeRoomInvite(invite: any): RoomInviteItem {
    return {
        ...invite,
        room: invite?.room ?? (invite?.room_name ? { name: invite.room_name } : undefined),
        inviter: invite?.inviter ?? (invite?.inviter_username ? { username: invite.inviter_username } : undefined),
    };
}

export default function HomeArea() {
    const navigate = useNavigate();
    const [activeTab, setActiveTab] = useState<FriendTab>('all');
    const { friends, pendingRequests, onlineFriendIds, fetchFriends, fetchRequests, sendRequest, respondRequest } = useFriendStore();
    const { setActiveDmUser, fetchRooms } = useChatStore();
    const [addUsername, setAddUsername] = useState('');
    const [addError, setAddError] = useState('');
    const [addSuccess, setAddSuccess] = useState('');
    const [selectedUser, setSelectedUser] = useState<User | null>(null);

    const [roomInvites, setRoomInvites] = useState<RoomInviteItem[]>([]);
    const fetchRoomInvites = useCallback(async () => {
        try {
            const res = await roomsApi.listInvites();
            setRoomInvites(Array.isArray(res) ? res.map(normalizeRoomInvite) : []);
        } catch (err) {
            console.error("Failed to fetch room invites", err);
        }
    }, []);

    useEffect(() => {
        fetchFriends();
        fetchRequests();
        const timer = window.setTimeout(() => {
            roomsApi.listInvites()
                .then((res) => setRoomInvites(Array.isArray(res) ? res.map(normalizeRoomInvite) : []))
                .catch((err) => console.error("Failed to fetch room invites", err));
        }, 0);
        return () => window.clearTimeout(timer);
    }, [fetchFriends, fetchRequests]);

    const handleAddFriend = async (e: React.FormEvent) => {
        e.preventDefault();
        setAddError('');
        setAddSuccess('');
        if (!addUsername.trim()) return;

        try {
            await sendRequest(addUsername.trim());
            setAddSuccess(`成功发送好友请求给 ${addUsername}`);
            setAddUsername('');
        } catch (err: any) {
            setAddError(err.response?.data?.message || '发送失败，请检查用户名是否正确');
        }
    };

    const handleStartDm = useCallback((userId: number) => {
        setActiveDmUser(userId);
        navigate(`/chat/dm/${userId}`);
    }, [navigate, setActiveDmUser]);

    return (
        <div className="flex-1 flex flex-col min-w-0 bg-[var(--bg-main)] relative transition-colors duration-300">
            {/* Topbar Re-imagined: Floating Pill Header */}
            <div className="pt-6 px-8 pb-4 shrink-0 z-10 sticky top-0 bg-gradient-to-b from-[var(--bg-main)] to-transparent">
                <div className="flex items-center justify-between text-[15px] p-2 rounded-2xl bg-[var(--bg-secondary)] border border-[var(--bg-sidebar)] shadow-sm backdrop-blur-md">
                    <div className="flex items-center ml-2 space-x-2">
                        <div className="flex items-center justify-center w-8 h-8 rounded-full bg-indigo-500 text-white shadow-sm animate-pulse" style={{ animationDuration: '3s' }}>
                            <Users className="w-4 h-4" />
                        </div>
                        <span className="font-bold tracking-wide text-[var(--text-main)] hidden sm:block">好友管理</span>
                        <div className="w-[1px] h-4 bg-[var(--text-muted)] opacity-20 mx-2 hidden sm:block"></div>
                    </div>

                    <div className="flex space-x-1 pr-1 overflow-x-auto hide-scrollbar">
                        <TabButton active={activeTab === 'online'} onClick={() => setActiveTab('online')}>在线</TabButton>
                        <TabButton active={activeTab === 'all'} onClick={() => setActiveTab('all')}>全部</TabButton>
                        <TabButton active={activeTab === 'pending'} onClick={() => setActiveTab('pending')}>
                            待处理
                            {pendingRequests.length > 0 && <span className="ml-1.5 px-1.5 py-0.5 rounded-full bg-[#ff6b6b] text-white text-[10px] font-bold leading-none">{pendingRequests.length}</span>}
                        </TabButton>
                        <TabButton active={activeTab === 'invites'} onClick={() => setActiveTab('invites')}>
                            频道邀请
                            {roomInvites.length > 0 && <span className="ml-1.5 px-1.5 py-0.5 rounded-full bg-[var(--accent)] text-white text-[10px] font-bold leading-none">{roomInvites.length}</span>}
                        </TabButton>
                        <button
                            onClick={() => setActiveTab('add')}
                            className={`px-4 py-1.5 ml-1 rounded-xl text-sm font-semibold transition-all duration-300 shadow-sm whitespace-nowrap ${activeTab === 'add' ? 'bg-[#20c997]/10 text-[#20c997] ring-1 ring-[#20c997]/50' : 'bg-[#20c997] text-white hover:bg-[#1bb386] hover:shadow-md hover:-translate-y-0.5'}`}
                        >
                            添加好友
                        </button>
                    </div>
                </div>
            </div>

            {/* Main Content Area */}
            <div className="flex-1 overflow-y-auto px-8 py-2 custom-scrollbar flex flex-col min-w-0 pb-16">

                {activeTab === 'online' && (
                    <div className="animate-in fade-in slide-in-from-bottom-2 duration-500">
                        {(() => {
                            const onlineFriends = friends.filter(f => onlineFriendIds.has(f.user?.id));
                            return (
                                <>
                                    <div className="text-xs font-bold text-[var(--text-muted)] tracking-wider uppercase mb-6 pl-2">
                                        在线联络人 ({onlineFriends.length})
                                    </div>
                                    {onlineFriends.length === 0 ? (
                                        <EmptyState icon="🪐" title="宇宙寂寥" subtitle="此刻没有好友处于在线状态。也许他们正在现实世界中探险？" />
                                    ) : (
                                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                            {onlineFriends.map(f => (
                                                <FriendCard key={f.friendship_id} friend={f} isOnline={true} onMessage={handleStartDm} onViewProfile={(u) => setSelectedUser(u)} />
                                            ))}
                                        </div>
                                    )}
                                </>
                            );
                        })()}
                    </div>
                )}

                {activeTab === 'all' && (
                    <div className="animate-in fade-in slide-in-from-bottom-2 duration-500">
                        <div className="text-xs font-bold text-[var(--text-muted)] tracking-wider uppercase mb-6 pl-2">
                            通讯录 ({friends.length})
                        </div>
                        {friends.length === 0 ? (
                            <EmptyState icon="🌵" title="一片荒芜" subtitle="你的好友列表目前空空如也。点击右上角的“添加好友”去探索新的人类吧！" />
                        ) : (
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                {friends.map(f => (
                                    <FriendCard key={f.friendship_id} friend={f} isOnline={onlineFriendIds.has(f.user?.id)} onMessage={handleStartDm} onViewProfile={(u) => setSelectedUser(u)} />
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {activeTab === 'pending' && (
                    <div className="animate-in fade-in slide-in-from-bottom-2 duration-500">
                        <div className="text-xs font-bold text-[var(--text-muted)] tracking-wider uppercase mb-6 pl-2">
                            待处理请求 ({pendingRequests.length})
                        </div>

                        {pendingRequests.length === 0 ? (
                            <EmptyState icon="🍃" title="风平浪静" subtitle="没有待处理的请求。享受这片刻的宁静吧！" />
                        ) : (
                            <div className="space-y-3 max-w-3xl">
                                {pendingRequests.map(r => (
                                    <div key={r.id} className="flex flex-col sm:flex-row sm:items-center justify-between p-4 rounded-2xl bg-[var(--bg-secondary)] border border-[var(--bg-sidebar)] hover:border-[var(--accent)]/30 hover:shadow-md transition-all duration-300">
                                        <div className="flex items-center space-x-4 mb-4 sm:mb-0">
                                            <div className="w-12 h-12 rounded-full bg-gradient-to-br from-[var(--accent)] to-[#4752c4] flex items-center justify-center text-white text-lg font-bold shadow-sm overflow-hidden">
                                                <img src={`https://api.dicebear.com/7.x/initials/svg?seed=${r.from?.username}`} alt="avatar" className="w-full h-full object-cover" />
                                            </div>
                                            <div>
                                                <div className="font-semibold text-lg leading-tight text-[var(--text-main)]">
                                                    {r.from?.username}
                                                </div>
                                                <div className="text-sm text-[var(--text-muted)] mt-1 flex items-center">
                                                    <span className="inline-block w-1.5 h-1.5 rounded-full bg-[var(--accent)] mr-2 animate-pulse"></span>
                                                    向您发送了好友请求
                                                </div>
                                            </div>
                                        </div>
                                        <div className="flex space-x-3 sm:ml-4">
                                            <button
                                                onClick={() => respondRequest(r.id, 'accept')}
                                                className="flex-1 sm:flex-none px-4 py-2 rounded-xl bg-[var(--bg-hover)] flex items-center justify-center text-[var(--text-main)] hover:text-white hover:bg-[#20c997] transition-all duration-300 font-medium whitespace-nowrap"
                                                title="接受"
                                            >
                                                接受请求
                                            </button>
                                            <button
                                                onClick={() => respondRequest(r.id, 'reject')}
                                                className="flex-1 sm:flex-none px-4 py-2 rounded-xl bg-[var(--bg-hover)] flex items-center justify-center text-[var(--text-muted)] hover:text-white hover:bg-[#ff6b6b] transition-all duration-300 font-medium whitespace-nowrap"
                                                title="忽略"
                                            >
                                                忽略
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {activeTab === 'invites' && (
                    <div className="animate-in fade-in slide-in-from-bottom-2 duration-500">
                        <div className="text-xs font-bold text-[var(--text-muted)] tracking-wider uppercase mb-6 pl-2">
                            频道邀请 ({roomInvites.length})
                        </div>

                        {roomInvites.length === 0 ? (
                            <EmptyState icon="📬" title="信箱空空" subtitle="目前没有收到任何频道邀请。" />
                        ) : (
                            <div className="space-y-3 max-w-3xl">
                                {roomInvites.map(invite => (
                                    <div key={invite.id} className="flex flex-col sm:flex-row sm:items-center justify-between p-4 rounded-2xl bg-[var(--bg-secondary)] border border-[var(--bg-sidebar)] hover:border-[var(--accent)]/30 hover:shadow-md transition-all duration-300">
                                        <div className="flex items-center space-x-4 mb-4 sm:mb-0">
                                            <div className="w-12 h-12 rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white text-lg font-bold shadow-sm">
                                                #
                                            </div>
                                            <div>
                                                <div className="font-semibold text-lg leading-tight text-[var(--text-main)]">
                                                    {invite.room?.name || '未知频道'}
                                                </div>
                                                <div className="text-sm text-[var(--text-muted)] mt-1 flex items-center">
                                                    <span className="font-bold mr-1 text-[var(--text-main)]">@{invite.inviter?.username || '未知的人'}</span> 邀请你加入频道
                                                </div>
                                            </div>
                                        </div>
                                        <div className="flex space-x-3 sm:ml-4">
                                            <button
                                                onClick={async () => {
                                                    try {
                                                        await roomsApi.respondInvite(invite.id, 'accept');
                                                        await fetchRoomInvites();
                                                        fetchRooms();
                                                    } catch (err: any) { alert("失败: " + err.message); }
                                                }}
                                                className="flex-1 sm:flex-none px-4 py-2 rounded-xl bg-[var(--bg-hover)] flex items-center justify-center text-[var(--text-main)] hover:text-white hover:bg-[#20c997] transition-all duration-300 font-medium whitespace-nowrap"
                                                title="加入"
                                            >
                                                加入频道
                                            </button>
                                            <button
                                                onClick={async () => {
                                                    try {
                                                        await roomsApi.respondInvite(invite.id, 'decline');
                                                        await fetchRoomInvites();
                                                    } catch (err: any) { alert("失败: " + err.message); }
                                                }}
                                                className="flex-1 sm:flex-none px-4 py-2 rounded-xl bg-[var(--bg-hover)] flex items-center justify-center text-[var(--text-muted)] hover:text-white hover:bg-[#ff6b6b] transition-all duration-300 font-medium whitespace-nowrap"
                                                title="拒绝"
                                            >
                                                拒绝
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {activeTab === 'add' && (
                    <div className="animate-in fade-in slide-in-from-bottom-2 duration-500 flex justify-center py-10">
                        <div className="w-full max-w-xl p-8 rounded-3xl bg-[var(--bg-secondary)] border border-[var(--bg-sidebar)] shadow-xl relative overflow-hidden">
                            {/* Decorative background shapes for flavor */}
                            <div className="absolute -top-10 -right-10 w-40 h-40 bg-[var(--accent)] opacity-10 rounded-full blur-3xl pointer-events-none"></div>
                            <div className="absolute -bottom-10 -left-10 w-40 h-40 bg-[#f7c116] opacity-10 rounded-full blur-3xl pointer-events-none"></div>

                            <div className="relative z-10">
                                <h2 className="text-2xl font-bold mb-3">连接新世界 🚀</h2>
                                <p className="text-[var(--text-muted)] mb-8 leading-relaxed">
                                    输入对方独一无二的用户名，建立你们之间的引力场。世界因此改变。
                                </p>

                                <form onSubmit={handleAddFriend} className="space-y-4">
                                    <div className={`relative flex items-center rounded-2xl bg-[var(--bg-main)] border-2 transition-all duration-300 ${addError ? 'border-[#ff6b6b]' : addSuccess ? 'border-[#20c997]' : 'border-[var(--bg-sidebar)] focus-within:border-[var(--accent)] shadow-sm'}`}>
                                        <div className="pl-4 text-[var(--text-muted)]">@</div>
                                        <input
                                            type="text"
                                            value={addUsername}
                                            onChange={e => setAddUsername(e.target.value)}
                                            placeholder="输入对方用户名"
                                            className="flex-1 bg-transparent border-none outline-none px-3 py-4 text-[16px] text-[var(--text-main)] placeholder-[var(--text-muted)]/50"
                                        />
                                    </div>

                                    <button
                                        type="submit"
                                        disabled={!addUsername.trim()}
                                        className="w-full bg-[var(--accent)] hover:bg-[#5b4eb3] disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-4 rounded-2xl text-[16px] font-bold transition-all duration-300 shadow-md hover:shadow-lg hover:-translate-y-0.5 active:translate-y-0"
                                    >
                                        发送好友请求
                                    </button>
                                </form>

                                {addError && (
                                    <div className="mt-4 p-3 rounded-xl bg-[#ff6b6b]/10 border border-[#ff6b6b]/20 text-[#ff6b6b] text-sm animate-in fade-in slide-in-from-top-1">
                                        ⚠️ {addError}
                                    </div>
                                )}
                                {addSuccess && (
                                    <div className="mt-4 p-3 rounded-xl bg-[#20c997]/10 border border-[#20c997]/20 text-[#20c997] text-sm animate-in fade-in slide-in-from-top-1">
                                        🎉 {addSuccess}
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                )}

            </div>

            <UserProfileModal
                user={selectedUser}
                isOpen={!!selectedUser}
                onClose={() => setSelectedUser(null)}
            />
        </div>
    );
}

function TabButton({ children, active, onClick }: { children: React.ReactNode, active: boolean, onClick: () => void }) {
    return (
        <button
            onClick={onClick}
            className={`px-4 py-1.5 rounded-xl text-sm font-semibold transition-all duration-300 ${active ? 'bg-[var(--bg-main)] text-[var(--text-main)] shadow-sm' : 'text-[var(--text-muted)] hover:bg-[var(--bg-main)]/50 hover:text-[var(--text-main)]'}`}
        >
            {children}
        </button>
    );
}

function FriendCard({ friend: f, isOnline, onMessage, onViewProfile }: { friend: FriendInfo, isOnline: boolean, onMessage?: (id: number) => void, onViewProfile: (user: User) => void }) {
    const [showMenu, setShowMenu] = useState(false);
    const menuRef = useRef<HTMLDivElement>(null);
    const { removeFriend } = useFriendStore();

    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
                setShowMenu(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleRemove = async (e: React.MouseEvent) => {
        e.stopPropagation();
        if (window.confirm(`确定要移除好友 ${f.user.username} 吗？`)) {
            await removeFriend(f.friendship_id);
            setShowMenu(false);
        }
    };

    return (
        <div className="relative group p-4 rounded-2xl bg-[var(--bg-secondary)] border border-[var(--bg-sidebar)] hover:border-[var(--accent)]/40 hover:shadow-lg transition-all duration-300 flex items-center justify-between cursor-pointer">
            {/* Clickable area for chatting */}
            <div className="absolute inset-0 z-0 rounded-2xl overflow-hidden pointer-events-none">
                {/* Subtle background glow effect on hover */}
                <div className={`absolute -top-10 -right-10 w-32 h-32 blur-2xl opacity-0 group-hover:opacity-10 transition-opacity duration-500 ${isOnline ? 'bg-[#20c997]' : 'bg-[var(--accent)]'}`}></div>
            </div>

            <div className="absolute inset-0 z-10" onClick={() => onMessage?.(f.user.id)}></div>

            <div className="flex items-center space-x-4 relative z-10 w-full">
                <div
                    className="relative shrink-0 cursor-pointer hover:opacity-80 transition-opacity pointer-events-auto z-20"
                    onClick={(e) => { e.stopPropagation(); onViewProfile(f.user); }}
                    title="查看资料"
                >
                    <div className="w-12 h-12 rounded-full bg-gradient-to-br from-[var(--bg-sidebar)] to-[var(--bg-input)] flex items-center justify-center text-[var(--text-main)] text-xl font-bold shadow-inner ring-1 ring-white/5 overflow-hidden">
                        <img src={f.user?.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${f.user?.username}`} alt={f.user?.username} className="w-full h-full object-cover" />
                    </div>
                    {/* Status Dot */}
                    <div className={`absolute bottom-0 right-0 w-3.5 h-3.5 border-2 border-[var(--bg-secondary)] group-hover:border-[var(--bg-main)] rounded-full transition-colors duration-300 ${isOnline ? 'bg-[#20c997]' : 'bg-[var(--text-muted)]'}`}></div>
                </div>

                <div className="min-w-0 flex-1 pr-2 pointer-events-none">
                    <div className="font-bold text-[16px] truncate text-[var(--text-main)] flex items-baseline">
                        <span>{f.user?.username}</span>
                    </div>
                    <div className={`text-xs mt-1 transition-colors ${isOnline ? 'text-[#20c997] font-medium' : 'text-[var(--text-muted)]'}`}>
                        {isOnline ? '在线' : '离线'}
                    </div>
                </div>

                {/* Quick actions reveal on hover */}
                <div className="absolute right-0 flex opacity-0 group-hover:opacity-100 transform translate-x-4 group-hover:translate-x-0 space-x-2 transition-all duration-300 ease-out z-20" ref={menuRef}>
                    <button
                        title="发消息"
                        onClick={(e) => { e.stopPropagation(); onMessage?.(f.user.id); }}
                        className="w-10 h-10 rounded-full bg-[var(--bg-main)] shadow-sm flex items-center justify-center text-[var(--text-muted)] hover:text-[#6c5dd3] hover:ring-2 ring-[var(--accent)]/30 transition-all cursor-pointer relative pointer-events-auto">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M7.9 20A9 9 0 1 0 4 16.1L2 22Z"></path></svg>
                    </button>
                    <div className="relative">
                        <button
                            title="更多选项"
                            onClick={(e) => { e.stopPropagation(); setShowMenu(!showMenu); }}
                            className={`w-10 h-10 rounded-full bg-[var(--bg-main)] shadow-sm flex items-center justify-center text-[var(--text-muted)] hover:text-[var(--text-main)] hover:bg-[var(--bg-hover)] transition-all cursor-pointer relative pointer-events-auto ${showMenu ? 'bg-[var(--bg-hover)] text-[var(--text-main)] opacity-100' : ''}`}>
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="1"></circle><circle cx="12" cy="5" r="1"></circle><circle cx="12" cy="19" r="1"></circle></svg>
                        </button>

                        {/* Dropdown Menu */}
                        {showMenu && (
                            <div className="absolute right-0 top-12 w-36 bg-[var(--bg-main)] border border-[var(--bg-sidebar)] rounded-lg shadow-xl py-1 z-50 animate-in fade-in slide-in-from-top-2 duration-200">
                                <button
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        navigator.clipboard.writeText(f.user.username);
                                        setShowMenu(false);
                                    }}
                                    className="w-full text-left px-3 py-2 text-sm text-[var(--text-main)] hover:bg-[var(--bg-secondary)] transition-colors flex items-center"
                                >
                                    <svg className="w-4 h-4 mr-2 text-[var(--text-muted)]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path></svg>
                                    复制用户名
                                </button>
                                <div className="h-[1px] bg-[var(--bg-sidebar)] my-1"></div>
                                <button
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        onViewProfile(f.user);
                                        setShowMenu(false);
                                    }}
                                    className="w-full text-left px-3 py-2 text-sm text-[var(--text-main)] hover:bg-[var(--bg-secondary)] transition-colors flex items-center"
                                >
                                    👤 查看资料
                                </button>
                                <div className="h-[1px] bg-[var(--bg-sidebar)] my-1"></div>
                                <button
                                    onClick={handleRemove}
                                    className="w-full text-left px-3 py-2 text-sm text-[#ff6b6b] hover:bg-[#ff6b6b]/10 transition-colors flex items-center"
                                >
                                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
                                    移除好友
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}

function EmptyState({ icon, title, subtitle }: { icon: string, title: string, subtitle: string }) {
    return (
        <div className="flex flex-col items-center justify-center py-24 text-center">
            <div className="text-7xl mb-6 filter drop-shadow-md animate-bounce" style={{ animationDuration: '3s' }}>{icon}</div>
            <h3 className="text-2xl font-bold text-[var(--text-main)] mb-3">{title}</h3>
            <p className="text-[15px] text-[var(--text-muted)] max-w-sm leading-relaxed">{subtitle}</p>
        </div>
    );
}
