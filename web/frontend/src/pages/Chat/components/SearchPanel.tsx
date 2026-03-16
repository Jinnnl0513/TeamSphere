import { useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import type { ChatMessage } from '../../../stores/chat/chatStore.types';
import { chatApi } from '../../../services/api/chat';

function highlightText(text: string, keyword: string) {
    if (!keyword) return text;
    const escaped = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const regex = new RegExp(escaped, 'ig');
    const parts = text.split(regex);
    const matches = text.match(regex) || [];
    return parts.reduce<ReactNode[]>((acc, part, idx) => {
        acc.push(<span key={`t-${idx}`}>{part}</span>);
        const match = matches[idx];
        if (match) {
            acc.push(
                <mark key={`m-${idx}`} className="bg-yellow-300/60 text-[var(--text-main)] px-0.5 rounded">
                    {match}
                </mark>
            );
        }
        return acc;
    }, []);
}

export default function SearchPanel({
    isOpen,
    roomId,
    onClose,
    onJumpToMessage,
}: {
    isOpen: boolean;
    roomId: number;
    onClose: () => void;
    onJumpToMessage: (msgId: number) => void;
}) {
    const [query, setQuery] = useState('');
    const [senderId, setSenderId] = useState('');
    const [from, setFrom] = useState('');
    const [to, setTo] = useState('');
    const [results, setResults] = useState<ChatMessage[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const trimmedQuery = useMemo(() => query.trim(), [query]);

    useEffect(() => {
        if (!isOpen) return;
        setResults([]);
        setError('');
    }, [isOpen, roomId]);

    useEffect(() => {
        if (!isOpen) return;
        if (!trimmedQuery) {
            setResults([]);
            setError('');
            return;
        }
        const timer = window.setTimeout(async () => {
            setLoading(true);
            setError('');
            try {
                const res = await chatApi.searchMessages({
                    q: trimmedQuery,
                    roomId,
                    senderId: senderId ? Number(senderId) : undefined,
                    from: from || undefined,
                    to: to || undefined,
                    limit: 50,
                });
                setResults(Array.isArray(res) ? res : []);
            } catch (e: any) {
                setError(e?.message || '搜索失败');
            } finally {
                setLoading(false);
            }
        }, 300);
        return () => window.clearTimeout(timer);
    }, [trimmedQuery, senderId, from, to, roomId, isOpen]);

    if (!isOpen) return null;

    return (
        <div className="fixed right-0 top-0 h-full w-[360px] max-w-[90vw] bg-[var(--bg-main)] border-l border-[var(--bg-secondary)] shadow-xl z-50 flex flex-col">
            <div className="h-12 flex items-center justify-between px-4 border-b border-[var(--bg-sidebar)]">
                <span className="font-semibold text-[var(--text-main)]">搜索消息</span>
                <button className="text-[var(--text-muted)] hover:text-[var(--text-main)]" onClick={onClose}>×</button>
            </div>

            <div className="p-4 space-y-3">
                <input
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    className="w-full bg-[var(--bg-secondary)] text-[var(--text-main)] rounded-md px-3 py-2 text-sm outline-none"
                    placeholder="输入关键词..."
                />
                <div className="grid grid-cols-2 gap-2">
                    <input
                        value={senderId}
                        onChange={(e) => setSenderId(e.target.value.replace(/\D/g, ''))}
                        className="bg-[var(--bg-secondary)] text-[var(--text-main)] rounded-md px-3 py-2 text-sm outline-none"
                        placeholder="发送者ID"
                    />
                    <div className="text-xs text-[var(--text-muted)] flex items-center">可选筛选</div>
                </div>
                <div className="grid grid-cols-2 gap-2">
                    <input
                        type="date"
                        value={from}
                        onChange={(e) => setFrom(e.target.value)}
                        className="bg-[var(--bg-secondary)] text-[var(--text-main)] rounded-md px-3 py-2 text-sm outline-none"
                    />
                    <input
                        type="date"
                        value={to}
                        onChange={(e) => setTo(e.target.value)}
                        className="bg-[var(--bg-secondary)] text-[var(--text-main)] rounded-md px-3 py-2 text-sm outline-none"
                    />
                </div>
            </div>

            <div className="flex-1 overflow-y-auto custom-scrollbar px-4 pb-4">
                {loading && <div className="text-[var(--text-muted)] text-sm">搜索中...</div>}
                {error && <div className="text-red-400 text-sm">{error}</div>}
                {!loading && !error && results.length === 0 && trimmedQuery && (
                    <div className="text-[var(--text-muted)] text-sm italic">没有找到匹配消息</div>
                )}
                <div className="space-y-3">
                    {results.map((msg) => (
                        <button
                            key={msg.id}
                            onClick={() => onJumpToMessage(msg.id)}
                            className="w-full text-left p-3 rounded-md border border-[var(--bg-secondary)] hover:bg-[var(--bg-secondary)] transition-colors"
                        >
                            <div className="text-xs text-[var(--text-muted)]">{msg.user?.username || 'System'}</div>
                            <div className="text-sm text-[var(--text-main)] mt-1 line-clamp-2">
                                {highlightText(msg.content || '', trimmedQuery)}
                            </div>
                        </button>
                    ))}
                </div>
            </div>
        </div>
    );
}
