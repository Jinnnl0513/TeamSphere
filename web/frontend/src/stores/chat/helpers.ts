import type { ChatMessage } from './chatStore.types';
import { wsGlobals } from './wsGlobals';

export function genClientMsgId(): string {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
        return crypto.randomUUID();
    }
    return `${Date.now()}-${Math.random().toString(36).substring(2)}`;
}

export function upsertMessage(msgs: ChatMessage[], newMsg: ChatMessage): ChatMessage[] {
    if (newMsg.client_msg_id) {
        const idx = msgs.findIndex(m => m.client_msg_id === newMsg.client_msg_id);
        if (idx !== -1) {
            const next = [...msgs];
            next[idx] = { ...next[idx], ...newMsg };
            return next;
        }
    }
    if (newMsg.id > 0) {
        const idIdx = msgs.findIndex(m => m.id === newMsg.id);
        if (idIdx !== -1) {
            const next = [...msgs];
            next[idIdx] = { ...next[idIdx], ...newMsg };
            return next;
        }
    }
    return [...msgs, newMsg];
}

export function startTypingCleanup(get: any, set: any) {
    if (wsGlobals.getTypingCleanupTimer()) return;
    const timer = setInterval(() => {
        const now = Date.now();
        const state = get();
        const updated: typeof state.typingUsers = {};
        let changed = false;
        for (const key in state.typingUsers) {
            const filtered = state.typingUsers[key].filter((u: any) => u.expiresAt > now);
            if (filtered.length !== state.typingUsers[key].length) changed = true;
            if (filtered.length > 0) updated[key] = filtered;
        }
        if (changed) set({ typingUsers: updated });
    }, 2000);
    wsGlobals.setTypingCleanupTimer(timer);
}

export function stopTypingCleanup() {
    const timer = wsGlobals.getTypingCleanupTimer();
    if (timer) {
        clearInterval(timer);
        wsGlobals.setTypingCleanupTimer(null);
    }
}
