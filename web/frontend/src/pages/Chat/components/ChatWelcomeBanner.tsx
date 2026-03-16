import type { User, Room } from '../../../types/models';

interface ChatWelcomeBannerProps {
    isDm: boolean;
    titleName: string;
    dmPeerInfo: User | undefined | null;
    activeRoom: Room | undefined | null;
    myRole: 'owner' | 'admin' | 'member';
    setIsSettingsOpen: React.Dispatch<React.SetStateAction<boolean>>;
}

export default function ChatWelcomeBanner({
    isDm,
    titleName,
    dmPeerInfo,
    activeRoom,
    myRole,
    setIsSettingsOpen
}: ChatWelcomeBannerProps) {
    return (
        <div className="px-4 pt-4 pb-0 space-y-2">
            {isDm && (
                <>
                    <div className="w-16 h-16 rounded-full bg-[var(--bg-secondary)] flex items-center justify-center overflow-hidden relative">
                        <img src={dmPeerInfo?.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${dmPeerInfo?.username}`} className="w-full h-full object-cover" />
                    </div>
                    <h1 className="text-3xl font-bold break-words">
                        与 {titleName} 的私聊
                    </h1>
                </>
            )}
            <p className="text-[var(--text-muted)]">
                {isDm ? `这是你与 ${titleName} 私聊记录的开始。` : (activeRoom?.description || `这是 #${titleName} 频道的开始。`)}
            </p>
            {!isDm && (myRole === 'owner' || myRole === 'admin') && (
                <button onClick={() => setIsSettingsOpen(true)} className="text-[var(--accent)] hover:underline flex items-center shrink-0">
                    <svg className="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                    </svg>
                    编辑频道
                </button>
            )}
            <hr className="border-[var(--bg-secondary)] mt-2" />
        </div>
    );
}
