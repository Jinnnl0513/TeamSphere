import { useState, useEffect } from 'react';
import type { User } from '../../../types/models';
import { usersApi } from '../../../services/api/users';

interface UserProfileModalProps {
    user: User | null;
    isOpen: boolean;
    onClose: () => void;
}

export default function UserProfileModal({ user, isOpen, onClose }: UserProfileModalProps) {
    const [fetchedUser, setFetchedUser] = useState<User | null>(null);

    useEffect(() => {
        if (!(isOpen && user?.id)) return;

        let cancelled = false;
        usersApi.getProfile(user.id).then((res) => {
            if (!cancelled && res) {
                setFetchedUser(res);
            }
        }).catch(console.error);

        return () => {
            cancelled = true;
        };
    }, [isOpen, user?.id]);

    if (!isOpen || !user) return null;

    const displayUser = fetchedUser?.id === user.id ? fetchedUser : user;

    const bannerColor = displayUser.profile_color || 'var(--bg-secondary)';
    const avatarUrl = displayUser.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${displayUser.username}`;
    const displayName = displayUser.username;

    return (
        <div className="fixed inset-0 z-[1000] flex items-center justify-center p-4">
            {/* Backdrop */}
            <div
                className="absolute inset-0 bg-black/50 backdrop-blur-sm animate-in fade-in duration-200"
                onClick={onClose}
            ></div>

            {/* Modal Content */}
            <div
                className="relative bg-[var(--bg-main)] rounded-lg shadow-2xl overflow-hidden w-full max-w-[340px] animate-in zoom-in-95 duration-200"
                onClick={(e) => e.stopPropagation()}
            >
                {/* Banner */}
                <div
                    className="h-24 w-full"
                    style={{ backgroundColor: bannerColor }}
                ></div>

                {/* Avatar */}
                <div className="absolute top-[52px] left-4">
                    <div className="w-[88px] h-[88px] rounded-full border-[6px] border-[var(--bg-main)] bg-[var(--bg-main)] overflow-hidden">
                        <img
                            src={avatarUrl}
                            alt={displayUser.username}
                            className="w-full h-full object-cover bg-white"
                        />
                    </div>
                </div>

                {/* Close Button */}
                <button
                    onClick={onClose}
                    className="absolute top-3 right-3 w-8 h-8 flex items-center justify-center rounded-full bg-black/20 text-white hover:bg-black/40 transition-colors"
                >
                    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>

                <div className="pt-14 px-4 pb-4">
                    <div className="bg-[var(--bg-secondary)]/30 rounded-lg p-3 border border-[var(--bg-secondary)]">
                        {/* Headers */}
                        <div className="mb-3 border-b border-[var(--bg-sidebar)] pb-3">
                            <h2 className="text-xl font-bold text-[var(--text-main)] leading-none break-words">
                                {displayName}
                            </h2>
                            <div className="text-[14px] text-[var(--text-main)] mt-1.5 flex items-center">
                                <span>{displayUser.username}</span>
                            </div>
                        </div>

                        {/* Bio */}
                        <div>
                            <h3 className="text-xs font-bold text-[var(--text-muted)] uppercase mb-2">
                                个人简介
                            </h3>
                            {displayUser.bio ? (
                                <p className="text-[14px] text-[var(--text-main)] break-words whitespace-pre-wrap leading-relaxed">
                                    {displayUser.bio}
                                </p>
                            ) : (
                                <p className="text-[14px] text-[var(--text-muted)] italic">
                                    这个人很懒，什么都没写~
                                </p>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
