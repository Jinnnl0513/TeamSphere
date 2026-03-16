import { useChatStore } from '../../../stores/chatStore';
import { MessageCircle, Settings } from 'lucide-react';

import { useNavigate } from 'react-router-dom';

export default function ServerSidebar({ onOpenSettings }: { onOpenSettings?: () => void }) {
    const navigate = useNavigate();
    const { activeView, conversations, unreadDmCounts } = useChatStore();

    const isHomeOrDm = activeView === 'home' || activeView === 'dm';

    const totalUnread = conversations.reduce((acc, conv) => {
        return acc + (conv.unread_count || 0) + (unreadDmCounts[conv.user.id] || 0);
    }, 0);

    return (
        <div className="w-[72px] bg-[var(--bg-sidebar)] flex-shrink-0 flex flex-col items-center py-3 space-y-2 z-20 overflow-y-auto hide-scrollbar">
            {/* Home / Direct Messages */}
            <div className="relative group flex justify-center w-full" title="主页">
                <div className={`absolute left-0 w-1 bg-[var(--text-main)] rounded-r-md top-1/2 -translate-y-1/2 transition-all duration-200 ${isHomeOrDm ? 'h-10' : 'h-0 group-hover:h-5'}`}></div>
                <div
                    className={`w-12 h-12 rounded-[16px] flex items-center justify-center transition-all duration-200 cursor-pointer shadow-sm relative ${isHomeOrDm ? 'bg-[var(--accent)] text-white' : 'bg-[var(--bg-main)] hover:bg-[var(--accent)] text-[var(--text-main)] hover:text-white hover:rounded-[16px] rounded-[24px]'}`}
                    onClick={() => navigate('/chat/home')}
                >
                    <MessageCircle className="w-6 h-6" />
                    {totalUnread > 0 && (
                        <div className="absolute -bottom-1 -right-1 bg-[#ff6b6b] text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold shadow-sm border-4 border-[var(--bg-sidebar)] break-normal whitespace-nowrap leading-none">
                            {totalUnread > 99 ? '99+' : totalUnread}
                        </div>
                    )}
                </div>
            </div>

            <div className="w-8 h-[2px] bg-[var(--bg-secondary)] rounded-full mx-auto my-2 shrink-0"></div>

            {/* Default Server / Chat Rooms */}
            <div className="relative group flex justify-center w-full" title="公共频道">
                <div className={`absolute left-0 w-1 bg-[var(--text-main)] rounded-r-md top-1/2 -translate-y-1/2 transition-all duration-200 ${!isHomeOrDm ? 'h-10' : 'h-0 group-hover:h-5'}`}></div>
                <div
                    className={`w-12 h-12 flex items-center justify-center transition-all duration-200 cursor-pointer shadow-sm bg-gradient-to-br from-indigo-500 to-purple-600 text-white font-bold text-lg ${!isHomeOrDm ? 'rounded-[16px]' : 'hover:rounded-[16px] rounded-[24px]'}`}
                    onClick={() => navigate('/chat/explore')}
                >
                    #
                </div>
            </div>

            {/* Settings Placeholder */}
            <div className="relative group flex justify-center w-full mb-2 mt-auto" title="全局设置">
                <div className="absolute left-0 w-1 bg-[var(--text-main)] rounded-r-md top-1/2 -translate-y-1/2 transition-all duration-200 h-0 group-hover:h-5"></div>
                <div
                    className="w-12 h-12 rounded-[24px] hover:rounded-[16px] bg-[var(--bg-main)] hover:bg-[var(--accent)] text-[var(--text-main)] hover:text-white flex items-center justify-center transition-all duration-200 cursor-pointer"
                    onClick={onOpenSettings}
                >
                    <Settings className="w-6 h-6" />
                </div>
            </div>
        </div>
    );
}
