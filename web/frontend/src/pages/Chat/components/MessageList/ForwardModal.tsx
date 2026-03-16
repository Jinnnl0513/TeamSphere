import { useMemo, useState } from 'react';
import type { ChatMessage } from '../../../../stores/chatStore';
import { useChatStore } from '../../../../stores/chatStore';
import { useFriendStore } from '../../../../stores/friendStore';
import { chatApi } from '../../../../services/api/chat';
import toast from 'react-hot-toast';
import { useApiCall } from '../../../../hooks/useApiCall';

export default function ForwardModal({
    isOpen,
    sourceType,
    message,
    onClose,
}: {
    isOpen: boolean;
    sourceType: 'room' | 'dm';
    message: ChatMessage | null;
    onClose: () => void;
}) {
    const rooms = useChatStore(s => s.rooms);
    const friends = useFriendStore(s => s.friends);
    const [tab, setTab] = useState<'room' | 'dm'>('room');
    const [query, setQuery] = useState('');
    const { loading: isSending, call } = useApiCall<void>();

    const filteredRooms = useMemo(() => {
        const q = query.trim().toLowerCase();
        if (!q) return rooms;
        return rooms.filter(r => r.name.toLowerCase().includes(q));
    }, [rooms, query]);

    const filteredFriends = useMemo(() => {
        const q = query.trim().toLowerCase();
        if (!q) return friends;
        return friends.filter(f => f.user.username.toLowerCase().includes(q));
    }, [friends, query]);

    if (!isOpen || !message) return null;

    const handleForward = async (targetRoomId?: number, targetUserId?: number) => {
        if (!message.id) return;
        try {
            await call(async () => {
                await chatApi.forwardMessage(message.id, {
                    message_type: sourceType,
                    target_room_id: targetRoomId,
                    target_user_id: targetUserId,
                });
            });
            toast.success('已转发');
            onClose();
        } catch {
            // handled by useApiCall
        }
    };

    return (
        <div className="fixed inset-0 z-[1000] bg-black/50 flex items-center justify-center">
            <div className="w-[520px] max-w-[90vw] bg-[var(--bg-main)] border border-[var(--bg-secondary)] rounded-lg shadow-2xl">
                <div className="px-4 py-3 border-b border-[var(--bg-secondary)] flex items-center justify-between">
                    <div className="text-[15px] font-semibold">转发消息</div>
                    <button onClick={onClose} className="text-[var(--text-muted)] hover:text-[var(--text-main)]">×</button>
                </div>

                <div className="px-4 py-3 flex items-center gap-2 border-b border-[var(--bg-secondary)]">
                    <button
                        className={`px-3 py-1.5 rounded-full text-sm ${tab === 'room' ? 'bg-[var(--bg-input)] text-[var(--text-main)]' : 'bg-[var(--bg-secondary)] text-[var(--text-muted)]'}`}
                        onClick={() => setTab('room')}
                    >
                        房间
                    </button>
                    <button
                        className={`px-3 py-1.5 rounded-full text-sm ${tab === 'dm' ? 'bg-[var(--bg-input)] text-[var(--text-main)]' : 'bg-[var(--bg-secondary)] text-[var(--text-muted)]'}`}
                        onClick={() => setTab('dm')}
                    >
                        私信
                    </button>
                    <input
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        placeholder="搜索..."
                        className="ml-auto bg-[var(--bg-input)] px-3 py-1.5 rounded text-sm outline-none text-[var(--text-main)] w-56"
                    />
                </div>

                <div className="max-h-[60vh] overflow-y-auto">
                    {tab === 'room' ? (
                        <div className="p-2">
                            {filteredRooms.map(room => (
                                <button
                                    key={room.id}
                                    onClick={() => handleForward(room.id, undefined)}
                                    disabled={isSending}
                                    className="w-full text-left px-3 py-2 rounded hover:bg-[var(--bg-secondary)] transition-colors flex items-center gap-2"
                                >
                                    <span className="text-[var(--text-muted)]">#</span>
                                    <span className="text-[var(--text-main)]">{room.name}</span>
                                </button>
                            ))}
                            {filteredRooms.length === 0 && (
                                <div className="text-center text-[var(--text-muted)] text-sm py-6">没有匹配的房间</div>
                            )}
                        </div>
                    ) : (
                        <div className="p-2">
                            {filteredFriends.map(f => (
                                <button
                                    key={f.user.id}
                                    onClick={() => handleForward(undefined, f.user.id)}
                                    disabled={isSending}
                                    className="w-full text-left px-3 py-2 rounded hover:bg-[var(--bg-secondary)] transition-colors flex items-center gap-2"
                                >
                                    <div className="w-7 h-7 rounded-full overflow-hidden bg-[var(--bg-secondary)]">
                                        <img src={f.user.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${f.user.username}`} className="w-full h-full object-cover" />
                                    </div>
                                    <span className="text-[var(--text-main)]">{f.user.username}</span>
                                </button>
                            ))}
                            {filteredFriends.length === 0 && (
                                <div className="text-center text-[var(--text-muted)] text-sm py-6">没有匹配的好友</div>
                            )}
                        </div>
                    )}
                </div>

                <div className="px-4 py-3 border-t border-[var(--bg-secondary)] text-xs text-[var(--text-muted)]">
                    点击目标立即转发
                </div>
            </div>
        </div>
    );
}
