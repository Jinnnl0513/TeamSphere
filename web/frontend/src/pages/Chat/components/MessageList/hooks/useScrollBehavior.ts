import { useRef, useState, useEffect, useCallback } from 'react';
import type { RefObject } from 'react';
import type { ChatMessage } from '../../../../../stores/chatStore';

export interface UseScrollBehaviorOptions {
    messages: ChatMessage[];
    isDm: boolean;
    roomId: number | null;
    dmId: number | null;
    hasMore: boolean;
    isLoadingOlder: boolean;
    skipInitialScroll?: boolean;
    fetchOlderHistory: (roomId: number) => Promise<void> | void;
    fetchOlderDmHistory: (dmId: number) => Promise<void> | void;
}

export interface UseScrollBehaviorReturn {
    listRef: RefObject<HTMLDivElement | null>;
    showScrollToBottom: boolean;
    handleScroll: () => void;
    scrollToBottom: (behavior?: ScrollBehavior) => void;
}

export function useScrollBehavior({
    messages,
    isDm,
    roomId,
    dmId,
    hasMore,
    isLoadingOlder,
    skipInitialScroll = false,
    fetchOlderHistory,
    fetchOlderDmHistory,
}: UseScrollBehaviorOptions): UseScrollBehaviorReturn {
    const listRef = useRef<HTMLDivElement | null>(null);
    const [showScrollToBottom, setShowScrollToBottom] = useState(false);
    const lastScrollHeightRef = useRef<number>(0);
    const initialScrolledRef = useRef<boolean>(false);

    const isNearBottom = (el: HTMLDivElement) => {
        const distance = el.scrollHeight - (el.scrollTop + el.clientHeight);
        return distance < 120;
    };

    const handleScroll = useCallback(() => {
        const el = listRef.current;
        if (!el) return;

        const nearBottom = isNearBottom(el);
        setShowScrollToBottom(!nearBottom);

        if (el.scrollTop <= 80 && hasMore && !isLoadingOlder) {
            if (isDm && dmId) fetchOlderDmHistory(dmId);
            else if (!isDm && roomId) fetchOlderHistory(roomId);
        }
    }, [hasMore, isLoadingOlder, isDm, dmId, roomId, fetchOlderHistory, fetchOlderDmHistory]);

    const scrollToBottom = useCallback((behavior: ScrollBehavior = 'smooth') => {
        const el = listRef.current;
        if (!el) return;
        el.scrollTo({ top: el.scrollHeight, behavior });
    }, []);

    useEffect(() => {
        const el = listRef.current;
        if (!el) return;

        if (!initialScrolledRef.current) {
            if (!skipInitialScroll) {
                el.scrollTop = el.scrollHeight;
            }
            initialScrolledRef.current = true;
            lastScrollHeightRef.current = el.scrollHeight;
            return;
        }

        if (isLoadingOlder) {
            const prev = lastScrollHeightRef.current;
            const next = el.scrollHeight;
            const delta = next - prev;
            if (delta > 0) {
                el.scrollTop = el.scrollTop + delta;
            }
        } else if (isNearBottom(el)) {
            el.scrollTop = el.scrollHeight;
        }

        lastScrollHeightRef.current = el.scrollHeight;
        setShowScrollToBottom(!isNearBottom(el));
    }, [messages.length, isLoadingOlder, skipInitialScroll]);

    return { listRef, showScrollToBottom, handleScroll, scrollToBottom };
}
