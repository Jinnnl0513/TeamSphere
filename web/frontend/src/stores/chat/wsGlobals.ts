let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let reconnectAttempts = 0;
let typingCleanupTimer: ReturnType<typeof setInterval> | null = null;

export const wsGlobals = {
    getReconnectTimer: () => reconnectTimer,
    setReconnectTimer: (timer: ReturnType<typeof setTimeout> | null) => { reconnectTimer = timer; },
    getReconnectAttempts: () => reconnectAttempts,
    incReconnectAttempts: () => { reconnectAttempts++; },
    resetReconnectAttempts: () => { reconnectAttempts = 0; },

    getTypingCleanupTimer: () => typingCleanupTimer,
    setTypingCleanupTimer: (timer: ReturnType<typeof setInterval> | null) => { typingCleanupTimer = timer; }
};
