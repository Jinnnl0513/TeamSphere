import type { ChatSliceCreator, ConnectionSlice } from './chatStore.types';
import { wsApi } from '../../services/api/ws';
import { wsGlobals } from './wsGlobals';
import { handleIncomingMessage } from './messageHandler';
import { startTypingCleanup, stopTypingCleanup } from './helpers';
import { WS_BASE_URL } from '../../config/app';

let currentConnectAttemptId = 0;

export const createConnectionSlice: ChatSliceCreator<ConnectionSlice> = (set, get) => ({
    socket: null,
    isConnected: false,
    isConnecting: false,

    connect: async () => {
        const { isConnected } = get();
        if (isConnected) return;

        currentConnectAttemptId++;
        const myAttemptId = currentConnectAttemptId;

        set({ isConnecting: true });

        try {
            const res = await wsApi.getTicket();
            const ticket = res?.ticket;
            if (!ticket) throw new Error('Could not obtain WS ticket');

            if (myAttemptId !== currentConnectAttemptId) return;

            const ws = new WebSocket(`${WS_BASE_URL}?ticket=${ticket}`);

            ws.onopen = () => {
                if (myAttemptId !== currentConnectAttemptId) {
                    ws.close(1000);
                    return;
                }
                wsGlobals.resetReconnectAttempts();
                set({ isConnected: true, isConnecting: false, socket: ws });
                startTypingCleanup(get, set);

                const state = get();
                state.fetchConversations();

                if (state.activeView === 'dm' && state.activeDmUserId) {
                    state.fetchDmHistory(state.activeDmUserId);
                } else if (state.activeView === 'rooms' && state.activeRoomId) {
                    state.joinRoom(state.activeRoomId);
                    state.fetchHistory(state.activeRoomId);
                }
            };

            ws.onclose = (event) => {
                const { socket } = get();
                if (socket && socket !== ws) return;

                set({ isConnected: false, socket: null, isConnecting: false });
                stopTypingCleanup();

                if (event.code !== 1000 && event.code !== 1001) {
                    const count = wsGlobals.getReconnectAttempts();
                    const delay = Math.min(1000 * (2 ** count), 30000);
                    wsGlobals.incReconnectAttempts();
                    const timer = wsGlobals.getReconnectTimer();
                    if (timer) clearTimeout(timer);
                    wsGlobals.setReconnectTimer(setTimeout(() => get().connect(), delay));
                }
            };

            ws.onerror = (error) => {
                console.error('Chat WS error', error);
            };

            ws.onmessage = (event) => {
                try {
                    const msg = JSON.parse(event.data as string);
                    handleIncomingMessage(msg.type, msg.data, get, set);
                } catch (err) {
                    console.error('Failed to parse WS message', err);
                }
            };
        } catch (error) {
            console.error('WS Connection failed:', error);
            if (myAttemptId === currentConnectAttemptId) {
                set({ isConnecting: false });
            }
        }
    },

    disconnect: () => {
        currentConnectAttemptId++;
        const timer = wsGlobals.getReconnectTimer();
        if (timer) { clearTimeout(timer); wsGlobals.setReconnectTimer(null); }
        stopTypingCleanup();
        const { socket } = get();
        if (socket) socket.close(1000);
        set({ socket: null, isConnected: false, isConnecting: false });
    },

    _reset: () => {
        currentConnectAttemptId++;
        stopTypingCleanup();
        const timer = wsGlobals.getReconnectTimer();
        if (timer) { clearTimeout(timer); wsGlobals.setReconnectTimer(null); }
        wsGlobals.resetReconnectAttempts();
        const { socket } = get();
        if (socket && socket.readyState === WebSocket.OPEN) socket.close(1000);

        localStorage.removeItem('chat_activeRoomId');
        localStorage.removeItem('chat_activeDmUserId');
        localStorage.removeItem('chat_activeView');

        set({
            socket: null,
            isConnected: false,
            isConnecting: false,
            rooms: [],
            messagesByRoom: {},
            messagesByDm: {},
            conversations: [],
            unreadDmCounts: {},
            typingUsers: {},
            onlineUsersByRoom: {},
            myMutedUntilByRoom: {},
            activeView: 'home',
            activeRoomId: null,
            activeDmUserId: null,
            hasMoreByRoom: {},
            isLoadingOlderByRoom: {},
            hasMoreByDm: {},
            isLoadingOlderByDm: {},
            mentionFilter: false,
            activeThreadMsgId: null,
            threadMessagesByMsgId: {},
            pinnedMessages: {},
        });
    }
});
