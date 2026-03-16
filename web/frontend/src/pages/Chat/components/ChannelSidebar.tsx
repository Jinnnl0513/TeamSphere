import { useChatStore } from '../../../stores/chatStore';
import UserProfilePanel from './UserProfilePanel';
import { Compass } from 'lucide-react';

import { useNavigate } from 'react-router-dom';

export default function ChannelSidebar() {
    const navigate = useNavigate();
    const { rooms, activeRoomId, unreadCountsByRoom } = useChatStore();

    return (
        <div className="w-[240px] bg-[var(--bg-secondary)] flex-shrink-0 flex flex-col z-10">
            {/* Header (Server Name / 'Direct Messages') */}
            <div className="h-12 border-b border-[var(--bg-sidebar)] flex items-center px-4 font-semibold text-[var(--text-main)] shadow-sm shrink-0 truncate">
                <span className="flex-1 truncate">我的服务器</span>
            </div>

            {/* Channel List Area */}
            <div className="flex-1 overflow-y-auto px-2 py-4 space-y-[2px] custom-scrollbar">
                {/* Discover/Explore Section */}
                <div
                    onClick={() => navigate('/chat/explore')}
                    className={`px-2 py-2 mb-4 rounded-md cursor-pointer flex items-center space-x-3 transition-all duration-200 shadow-sm ${activeRoomId === null
                        ? 'bg-[#20c997]/10 text-[#20c997] ring-1 ring-[#20c997]/50'
                        : 'text-[var(--text-main)] hover:bg-[var(--bg-main)]/60'
                        }`}
                >
                    <div className={`w-8 h-8 rounded-full flex items-center justify-center text-lg ${activeRoomId === null ? 'bg-[#20c997] text-white' : 'bg-[var(--bg-main)] text-[var(--text-muted)]'}`}>
                        <Compass className="w-5 h-5" />
                    </div>
                    <span className="font-bold text-[15px]">发现频道</span>
                </div>

                {/* Category Header */}
                <div className="flex items-center justify-between text-xs font-semibold text-[var(--text-muted)] uppercase px-1 mb-1 group">
                    <div className="flex items-center">
                        文本频道
                    </div>
                    <button
                        onClick={() => navigate('/chat/explore')}
                        title="浏览或创建频道"
                        className="opacity-0 group-hover:opacity-100 hover:text-[var(--text-main)] transition-all font-bold text-lg leading-none"
                    >
                        +
                    </button>
                </div>

                {rooms.map((room) => {
                    const isActive = room.id === activeRoomId;
                    const unread = unreadCountsByRoom[room.id] || 0;
                    return (
                        <div
                            key={room.id}
                            onClick={() => navigate(`/chat/rooms/${room.id}`)}
                            className={`px-2 py-1.5 rounded-md cursor-pointer flex items-center space-x-2 transition-colors duration-150 ${isActive
                                ? 'bg-[var(--bg-main)]/80 text-[var(--text-main)]'
                                : 'text-[var(--text-muted)] hover:text-[var(--text-main)] hover:bg-[var(--bg-main)]/40'
                                }`}
                        >
                            <span className="text-xl leading-none opacity-60">#</span>
                            <span className="font-medium text-[15px] truncate">{room.name}</span>
                            {unread > 0 && (
                                <span className="ml-auto text-[11px] px-2 py-0.5 rounded-full bg-[#f04747] text-white font-semibold">
                                    {unread > 99 ? '99+' : unread}
                                </span>
                            )}
                        </div>
                    );
                })}

                {rooms.length === 0 && (
                    <div className="px-2 py-6 text-center text-[13px] text-[var(--text-muted)] border border-dashed border-[var(--bg-sidebar)] rounded-lg mx-2 mt-4 cursor-pointer hover:border-[#20c997]/50 hover:text-[#20c997] transition-colors" onClick={() => navigate('/chat/explore')}>
                        还没有加入任何频道<br /><span className="font-bold mt-1 inline-block">点击去探索世界！</span>
                    </div>
                )}
            </div>

            <UserProfilePanel />
        </div>
    );
}
