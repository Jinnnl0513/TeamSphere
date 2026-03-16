import type { ReactNode, RefObject } from 'react';

export default function MessageListContainer({
    listRef,
    onScroll,
    isLoadingOlder,
    hasMore,
    messagesLength,
    isVirtualized,
    children,
}: {
    listRef: RefObject<HTMLDivElement | null>;
    onScroll: () => void;
    isLoadingOlder: boolean;
    hasMore: boolean;
    messagesLength: number;
    isVirtualized?: boolean;
    children: ReactNode;
}) {
    return (
        <div
            ref={listRef}
            onScroll={onScroll}
            className="h-full overflow-y-auto flex flex-col p-4 custom-scrollbar select-text pb-8"
        >
            <div className={`space-y-1 flex flex-col min-h-min w-full${isVirtualized ? '' : ' mt-auto'}`}>
                {isLoadingOlder && (
                    <div className="flex items-center justify-center py-3 text-[var(--text-muted)] text-sm gap-2">
                        <svg className="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4l3-3-3-3v4a8 8 0 00-8 8h4l-3 3 3 3H4z" />
                        </svg>
                        正在加载更早消息...
                    </div>
                )}
                {!isLoadingOlder && !hasMore && messagesLength > 0 && (
                    <div className="text-center text-[var(--text-muted)] text-xs py-2 opacity-60">已到最早消息</div>
                )}
                {messagesLength === 0 && (
                    <div className="text-center text-[var(--text-muted)] italic py-2">暂无消息...</div>
                )}
                {children}
            </div>
        </div>
    );
}
