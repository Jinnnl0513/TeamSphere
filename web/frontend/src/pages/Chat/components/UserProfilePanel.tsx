import { useState } from 'react';
import { useAuthStore } from '../../../stores/authStore';
import { useNavigate } from 'react-router-dom';
import UserProfileModal from './UserProfileModal';

export default function UserProfilePanel() {
    const { user, clearAuth } = useAuthStore();
    const navigate = useNavigate();
    const [isProfileModalOpen, setIsProfileModalOpen] = useState(false);

    const handleLogout = () => {
        clearAuth();
        navigate('/login');
    };

    return (
        <div className="h-[52px] bg-[var(--bg-sidebar)]/50 px-2 flex items-center shrink-0">
            <div className="flex items-center space-x-2 w-full hover:bg-[var(--bg-secondary)]/80 p-1 rounded-md transition-colors duration-150 relative group">
                {/* User Avatar */}
                <div 
                    className="w-8 h-8 rounded-full bg-[var(--accent)] flex-shrink-0 flex items-center justify-center text-white font-bold cursor-pointer relative"
                    onClick={() => setIsProfileModalOpen(true)}
                >
                    <img src={user?.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${user?.username}`} alt="avatar" className="w-full h-full rounded-full object-cover" />
                    <div className="absolute bottom-0 -right-0.5 w-3 h-3 bg-[var(--color-discord-green-500)] border-[2px] border-[var(--bg-sidebar)] rounded-full z-10 transition-colors"></div>
                </div>

                {/* User Info */}
                <div className="flex flex-col overflow-hidden max-w-[90px] cursor-default">
                    <span className="text-[14px] font-semibold text-[var(--text-main)] truncate leading-none mb-0.5">{user?.username || 'Guest'}</span>
                </div>

                {/* Controls */}
                <div className="flex-1 flex justify-end space-x-0.5">
                    <button className="w-8 h-8 rounded hover:bg-[var(--bg-secondary)] flex items-center justify-center text-[var(--text-muted)] hover:text-[var(--color-discord-red-400)] transition-colors" title="注销" onClick={handleLogout}>
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                        </svg>
                    </button>
                </div>
            </div>

            {/* User Profile Modal */}
            <UserProfileModal
                user={user}
                isOpen={isProfileModalOpen}
                onClose={() => setIsProfileModalOpen(false)}
            />
        </div>
    );
}
