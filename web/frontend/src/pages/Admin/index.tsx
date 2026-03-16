import { Routes, Route, Navigate } from 'react-router-dom';
import AdminSidebar from './components/AdminSidebar';
import StatsView from './views/StatsView';
import UsersView from './views/UsersView';
import RoomsView from './views/RoomsView';
import SettingsView from './views/SettingsView';
import EmailSettingsView from './views/EmailSettingsView';
import AnnouncementView from './views/AnnouncementView';

export default function AdminLayout() {
    return (
        <div className="flex h-screen w-full bg-[var(--bg-main)] text-[var(--text-main)] font-sans overflow-hidden">
            {/* Sidebar */}
            <AdminSidebar />

            {/* Main Content Area */}
            <div className="flex-1 flex flex-col h-full bg-[var(--bg-main)] overflow-y-auto">
                <main className="flex-1 p-8 w-full max-w-6xl mx-auto">
                    <Routes>
                        <Route path="/" element={<Navigate to="stats" replace />} />
                        <Route path="stats" element={<StatsView />} />
                        <Route path="users" element={<UsersView />} />
                        <Route path="rooms" element={<RoomsView />} />
                        <Route path="settings" element={<SettingsView />} />
                        <Route path="email" element={<EmailSettingsView />} />
                        <Route path="announcement" element={<AnnouncementView />} />
                        <Route path="*" element={<Navigate to="stats" replace />} />
                    </Routes>
                </main>
            </div>
        </div>
    );
}
