import { useEffect, useState } from 'react';
import UserProfilePanel from './UserProfilePanel';
import { useChatStore } from '../../../stores/chatStore';
import { useFriendStore } from '../../../stores/friendStore';
import { Users } from 'lucide-react';
import UserProfileModal from './UserProfileModal';
import type { User } from '../../../types/models';

import { useNavigate } from 'react-router-dom';

export default function HomeSidebar() {
    const navigate = useNavigate();
    const { conversations, activeDmUserId, fetchConversations, activeView, unreadDmCounts } = useChatStore();
    const { onlineFriendIds } = useFriendStore();
    const [selectedUser, setSelectedUser] = useState<User | null>(null);

    useEffect(() => {
        fetchConversations();
    }, [fetchConversations]);

    return (
        <div className="w-[240px] bg-[var(--bg-secondary)] flex-shrink-0 flex flex-col z-10 border-r border-black/20">
            <div className="h-12 border-b border-[var(--bg-sidebar)] flex items-center px-4 shrink-0 shadow-sm">
                <button className="w-full bg-[var(--bg-sidebar)] text-[var(--text-muted)] text-[13px] py-1.5 px-3 rounded-md text-left transition-colors duration-200 hover:bg-[var(--bg-main)]">
                    查找或开始新的对话
                </button>
            </div>

            <div className="flex-1 overflow-y-auto px-2 py-3 space-y-1 custom-scrollbar">
                <button
                    onClick={() => navigate('/chat/home')}
                    className={`w-full flex items-center px-3 py-2.5 rounded-lg transition-colors group ${activeView === 'home' ? 'bg-[#3b414a]/80 text-white shadow-sm' : 'text-[var(--text-main)] hover:bg-[var(--bg-main)]/60'}`}
                >
                    <Users className="w-5 h-5 mr-3 opacity-90 group-hover:opacity-100 transition-opacity" />
                    <span className="font-semibold text-[15px]">好友</span>
                </button>

                <div className="text-xs font-bold text-[var(--text-muted)] mt-5 mb-2 pl-3 uppercase tracking-wide hover:text-[var(--text-main)] transition-colors cursor-pointer group flex justify-between items-center pr-2">
                    <span>私信</span>
                    <span className="opacity-0 group-hover:opacity-100 text-[18px] leading-none mb-0.5">+</span>
                </div>

                {conversations.length === 0 ? (
                    <div className="px-3 py-4 text-[13px] text-[var(--text-muted)]/60 text-center italic mt-2 border border-dashed border-[var(--bg-sidebar)] rounded-lg mx-2">
                        世界很安静，快去结识新朋友吧
                    </div>
                ) : (
                    <div className="space-y-0.5">
                        {conversations.map(conv => {
                            const peer = conv.user;
                            const isActive = activeView === 'dm' && activeDmUserId === peer.id;
                            const totalUnread = (conv.unread_count || 0) + (unreadDmCounts[peer.id] || 0);
                            return (
                                <button
                                    key={peer.id}
                                    onClick={() => navigate(`/chat/dm/${peer.id}`)}
                                    className={`w-full flex items-center px-3 py-2 rounded-lg transition-all duration-200 ${isActive ? 'bg-[#3b414a] text-white shadow-sm scale-[1.02]' : 'text-[var(--text-muted)] hover:bg-[var(--bg-main)] hover:text-[var(--text-main)]'}`}
                                >
                                    <div
                                        className="relative cursor-pointer hover:opacity-80 transition-opacity"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            setSelectedUser(peer);
                                        }}
                                    >
                                        <div className="w-8 h-8 rounded-full bg-gradient-to-br from-[var(--bg-sidebar)] to-[var(--bg-main)] flex items-center justify-center font-bold overflow-hidden shadow-inner ring-1 ring-white/5">
                                            <img src={peer.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${peer.username}`} alt={peer.username} className="w-full h-full object-cover" />
                                        </div>
                                        {onlineFriendIds.has(peer.id) && (
                                            <div className="absolute -bottom-0.5 -right-0.5 w-[14px] h-[14px] border-[2.5px] border-[var(--bg-secondary)] group-hover:border-[#3b414a] rounded-full bg-[#20c997] z-10"></div>
                                        )}
                                    </div>
                                    <div className="ml-3 flex-1 text-left min-w-0 flex justify-between items-center">
                                        <span className={`font-medium text-[15px] truncate pr-2 ${isActive ? 'text-white' : 'text-[var(--text-main)]'} flex items-baseline`}>
                                            <span>{peer.username}</span>
                                        </span>
                                        {totalUnread > 0 && !isActive && (
                                            <span className="bg-[#ff6b6b] text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold shadow-sm whitespace-nowrap leading-none shrink-0 min-w-[20px] text-center">
                                                {totalUnread > 99 ? '99+' : totalUnread}
                                            </span>
                                        )}
                                    </div>
                                </button>
                            );
                        })}
                    </div>
                )}
            </div>

            <UserProfilePanel />

            <UserProfileModal
                user={selectedUser}
                isOpen={!!selectedUser}
                onClose={() => setSelectedUser(null)}
            />
        </div>
    );
}
