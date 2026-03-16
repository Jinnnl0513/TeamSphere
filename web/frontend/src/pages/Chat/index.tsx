import React, { useEffect, useState } from 'react';
import { Routes, Route, Navigate, useLocation } from 'react-router-dom';
import ServerSidebar from './components/ServerSidebar';
import ChannelSidebar from './components/ChannelSidebar';
import HomeSidebar from './components/HomeSidebar';
import ChatArea from './components/ChatArea';
import HomeArea from './components/HomeArea';
import ExploreArea from './components/ExploreArea';
import SettingsOverlay from './components/SettingsOverlay';
import { useChatStore } from '../../stores/chatStore';
import { announcementApi } from '../../services/api/announcement';

export default function ChatLayout({ children }: { children?: React.ReactNode }) {
    const connect = useChatStore(state => state.connect);
    const disconnect = useChatStore(state => state.disconnect);
    const fetchRooms = useChatStore(state => state.fetchRooms);
    const unreadCountsByRoom = useChatStore(state => state.unreadCountsByRoom);
    const unreadDmCounts = useChatStore(state => state.unreadDmCounts);

    // UI state actions to sync with URL
    const setActiveView = useChatStore(state => state.setActiveView);
    const setActiveRoomId = useChatStore(state => state.setActiveRoom);
    const setActiveDmUser = useChatStore(state => state.setActiveDmUser);

    // Fallback UI State getters for rendering the sidebars correctly
    const activeView = useChatStore(state => state.activeView);

    const [isSettingsOpen, setIsSettingsOpen] = useState(false);
    const [announcement, setAnnouncement] = useState<string | null>(null);

    const location = useLocation();

    // Sync URL paths back to global state so that existing Sidebar logic (like glowing icons) works out-of-the-box
    useEffect(() => {
        const path = location.pathname;
        let match: RegExpMatchArray | null;
        if (path.startsWith('/chat/home')) {
            setActiveView('home');
            setActiveRoomId(null);
            setActiveDmUser(null);
        } else if (path.startsWith('/chat/explore')) {
            setActiveView('rooms');
            setActiveRoomId(null);
        } else if ((match = path.match(/\/chat\/rooms\/(\d+)/))) {
            setActiveView('rooms');
            setActiveRoomId(Number(match[1]));
        } else if ((match = path.match(/\/chat\/dm\/(\d+)/))) {
            setActiveView('dm');
            setActiveDmUser(Number(match[1]));
        } else if (path === '/chat' || path === '/chat/') {
            // It will redirect to home, see <Routes> below
        }
    }, [location.pathname, setActiveView, setActiveRoomId, setActiveDmUser]);

    useEffect(() => {
        fetchRooms();
        connect();

        // Fetch Phase 2 system announcement if available
        announcementApi.getAnnouncement()
            .then(res => setAnnouncement(res.content))
            .catch(() => { }); // gracefully ignore if not implemented yet

        // Ensure WS close frame is sent on page refresh / tab close
        const handleBeforeUnload = () => {
            disconnect();
        };
        window.addEventListener('beforeunload', handleBeforeUnload);

        return () => {
            window.removeEventListener('beforeunload', handleBeforeUnload);
            disconnect();
        };
    }, [connect, disconnect, fetchRooms]);

    useEffect(() => {
        const totalUnreadRooms = Object.values(unreadCountsByRoom || {}).reduce((sum, n) => sum + (n || 0), 0);
        const totalUnreadDm = Object.values(unreadDmCounts || {}).reduce((sum, n) => sum + (n || 0), 0);
        const total = totalUnreadRooms + totalUnreadDm;
        document.title = total > 0 ? `(${total}) TeamSphere` : 'TeamSphere';
        const favicon = document.querySelector<HTMLLinkElement>('link[rel=\"icon\"]');
        if (favicon) {
            favicon.href = total > 0 ? '/favicon-unread.svg' : '/favicon.svg';
        }
    }, [unreadCountsByRoom, unreadDmCounts]);

    return (
        <div className="flex flex-col h-screen w-full overflow-hidden bg-[var(--bg-main)] text-[var(--text-main)] relative">
            {announcement && (
                <div className="w-full bg-gradient-to-r from-indigo-500 to-purple-600 text-white text-sm font-medium py-1.5 px-4 flex justify-between items-center z-50 shadow-sm shrink-0">
                    <div className="flex items-center">
                        <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" /></svg>
                        {announcement}
                    </div>
                    <button onClick={() => setAnnouncement(null)} className="hover:bg-black/20 rounded-full p-0.5 transition-colors text-white/80 hover:text-white" title="关闭">
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                    </button>
                </div>
            )}
            <div className="flex h-full w-full overflow-hidden relative">
                <ServerSidebar onOpenSettings={() => setIsSettingsOpen(true)} />
                {(activeView === 'home' || activeView === 'dm') ? <HomeSidebar /> : <ChannelSidebar />}

                {/* React Router nested routes for the main content area */}
                {children ? children : (
                    <Routes>
                        <Route path="/" element={<Navigate to="home" replace />} />
                        <Route path="home" element={<HomeArea />} />
                        <Route path="explore" element={<ExploreArea />} />
                        <Route path="rooms/:roomId" element={<ChatArea />} />
                        <Route path="dm/:userId" element={<ChatArea />} />
                    </Routes>
                )}

                <SettingsOverlay isOpen={isSettingsOpen} onClose={() => setIsSettingsOpen(false)} />
            </div>
        </div>
    );
}

