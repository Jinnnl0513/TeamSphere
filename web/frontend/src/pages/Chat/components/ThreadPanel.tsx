import { useEffect } from 'react';
import { useChatStore } from '../../../stores/chatStore';
import type { ChatMessage } from '../../../stores/chat/chatStore.types';
import { resolveFileUrl } from '../../../utils/urls';

export default function ThreadPanel({ roomId }: { roomId: number | null }) {
    const activeThreadMsgId = useChatStore(s => s.activeThreadMsgId);
    const setActiveThreadMsgId = useChatStore(s => s.setActiveThreadMsgId);
    const threadMessagesByMsgId = useChatStore(s => s.threadMessagesByMsgId);
    const fetchThreadMessages = useChatStore(s => s.fetchThreadMessages);
    const messagesByRoom = useChatStore(s => s.messagesByRoom);

    useEffect(() => {
        if (!roomId || !activeThreadMsgId) return;
        fetchThreadMessages(roomId, activeThreadMsgId);
    }, [roomId, activeThreadMsgId, fetchThreadMessages]);

    if (!roomId || !activeThreadMsgId) return null;

    const rootMsg = (messagesByRoom[roomId] || []).find(m => m.id === activeThreadMsgId) as ChatMessage | undefined;
    const threadMessages = threadMessagesByMsgId[activeThreadMsgId] || [];

    return (
        <div className="w-80 max-w-[40vw] border-l border-[var(--bg-secondary)] bg-[var(--bg-main)] h-full flex flex-col">
            <div className="h-12 flex items-center justify-between px-4 border-b border-[var(--bg-sidebar)]">
                <span className="font-semibold text-[var(--text-main)]">线程</span>
                <button
                    className="text-[var(--text-muted)] hover:text-[var(--text-main)] transition-colors"
                    onClick={() => setActiveThreadMsgId(null)}
                    title="关闭"
                >
                    ×
                </button>
            </div>

            <div className="px-4 py-3 border-b border-[var(--bg-secondary)] text-sm">
                {rootMsg ? (
                    <div>
                        <div className="font-medium text-[var(--text-main)]">{rootMsg.user?.username || 'System'}</div>
                        <div className="text-[var(--text-muted)] mt-1 line-clamp-3">{rootMsg.content}</div>
                    </div>
                ) : (
                    <div className="text-[var(--text-muted)]">线程消息不存在</div>
                )}
            </div>

            <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-3">
                {threadMessages.length === 0 && (
                    <div className="text-[var(--text-muted)] text-sm italic">暂无回复</div>
                )}
                {threadMessages.map((msg) => (
                    <div key={msg.id} className="flex gap-3">
                        <img
                            className="w-8 h-8 rounded-full object-cover"
                            src={msg.user?.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${msg.user?.username}`}
                            alt={msg.user?.username}
                        />
                        <div className="flex-1">
                            <div className="text-sm font-medium text-[var(--text-main)]">{msg.user?.username}</div>
                            {msg.msg_type === 'image' ? (
                                <img
                                    src={resolveFileUrl(msg.content)}
                                    alt="Thread image"
                                    className="mt-1 max-h-48 rounded"
                                />
                            ) : msg.msg_type === 'file' ? (
                                <a
                                    href={resolveFileUrl(msg.content)}
                                    target="_blank"
                                    rel="noreferrer"
                                    className="mt-1 inline-block text-sm text-[var(--accent)] hover:underline"
                                >
                                    查看文件
                                </a>
                            ) : (
                                <div className="text-sm text-[var(--text-main)] mt-1 whitespace-pre-wrap">{msg.content}</div>
                            )}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}
