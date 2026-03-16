import { useState, useCallback } from 'react';
import { useAuthStore } from '../../../stores/authStore';
import { chatApi } from '../../../services/api/chat';
import type { ReactionSummary } from '../../../stores/chatStore';


interface MessageReactionsProps {
    msgId: number;
    reactions: ReactionSummary[] | undefined;
    isDm: boolean;
    roomId: number | null;
    dmUserId: number | null;
    // Optimistic update callback
    onReactionChange: (msgId: number, reactions: ReactionSummary[]) => void;
}

export default function MessageReactions({
    msgId,
    reactions,
    isDm,
    roomId,
    dmUserId,
    onReactionChange,
}: MessageReactionsProps) {
    const { user } = useAuthStore();
    const [pending, setPending] = useState<Set<string>>(new Set());

    const myId = user?.id;

    const toggleReaction = useCallback(
        async (emoji: string) => {
            if (!myId || pending.has(emoji)) return;
            const existingReaction = (reactions ?? []).find(r => r.emoji === emoji);
            const alreadyReacted = existingReaction?.user_ids.includes(myId) ?? false;

            // Optimistic update
            const current = reactions ?? [];
            let updated: ReactionSummary[];
            if (alreadyReacted) {
                updated = current
                    .map(r =>
                        r.emoji === emoji
                            ? { ...r, count: r.count - 1, user_ids: r.user_ids.filter(id => id !== myId) }
                            : r
                    )
                    .filter(r => r.count > 0);
            } else {
                const idx = current.findIndex(r => r.emoji === emoji);
                if (idx === -1) {
                    updated = [...current, { emoji, count: 1, user_ids: [myId] }];
                } else {
                    updated = current.map(r =>
                        r.emoji === emoji
                            ? { ...r, count: r.count + 1, user_ids: [...r.user_ids, myId] }
                            : r
                    );
                }
            }
            onReactionChange(msgId, updated);

            setPending(p => new Set(p).add(emoji));
            try {
                const msgType = isDm ? 'dm' : 'room';
                if (alreadyReacted) {
                    await chatApi.removeReaction(
                        msgId, emoji, msgType,
                        roomId ?? undefined,
                        dmUserId ?? undefined
                    );
                } else {
                    await chatApi.addReaction(
                        msgId, emoji, msgType,
                        roomId ?? undefined,
                        dmUserId ?? undefined
                    );
                }
            } catch {
                // Rollback optimistic update on failure
                onReactionChange(msgId, reactions ?? []);
            } finally {
                setPending(p => {
                    const next = new Set(p);
                    next.delete(emoji);
                    return next;
                });
            }
        },
        [msgId, myId, reactions, isDm, roomId, dmUserId, pending, onReactionChange]
    );

    const reactionList = reactions ?? [];

    return (
        <div className="flex flex-wrap items-center gap-1 mt-1 select-none">
            {reactionList.map(r => {
                const reacted = myId ? r.user_ids.includes(myId) : false;
                const isPending = pending.has(r.emoji);
                return (
                    <button
                        key={r.emoji}
                        type="button"
                        disabled={isPending}
                        onClick={() => void toggleReaction(r.emoji)}
                        title={`${r.count} 人 ${r.emoji}`}
                        className={`
                            inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[13px]
                            border transition-all duration-150
                            ${reacted
                                ? 'border-[#5865F2]/60 bg-[#5865F2]/20 text-[var(--text-main)]'
                                : 'border-[var(--bg-secondary)] bg-[var(--bg-secondary)]/60 text-[var(--text-muted)] hover:border-[#5865F2]/40 hover:bg-[#5865F2]/10 hover:text-[var(--text-main)]'
                            }
                            ${isPending ? 'opacity-50 cursor-wait' : 'cursor-pointer'}
                        `}
                    >
                        <span>{r.emoji}</span>
                        <span className="text-[11px] font-medium tabular-nums">{r.count}</span>
                    </button>
                );
            })}
        </div>
    );
}
