import { create } from 'zustand';
import { AxiosError } from 'axios';
import type { User } from '../types/models';
import { ApiError } from '../api/client';
import { usersApi } from '../services/api/users';

interface AuthState {
    user: User | null;
    token: string | null;
    refreshToken: string | null;
    setAuth: (user: User, token: string, refreshToken?: string) => void;
    setTokens: (token: string, refreshToken: string) => void;
    clearAuth: () => void;
    loadUser: () => Promise<void>;
}

const AUTH_ERROR_CODES = new Set([40100, 40101, 40102, 40300, 40301]);

const clearPersistedAuth = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('refresh_token');
};

const resetChatState = () => {
    void import('./chatStore')
        .then(({ useChatStore }) => {
            useChatStore.getState()._reset();
        })
        .catch(() => {
            // Ignore chat store reset failures during auth cleanup.
        });
};

const resetAuthState = (setAuthState: (state: Partial<AuthState>) => void) => {
    clearPersistedAuth();
    setAuthState({ user: null, token: null, refreshToken: null });
    resetChatState();
};

const isAuthFailure = (error: unknown) => {
    if (error instanceof AxiosError) {
        const status = error.response?.status;
        return status === 401 || status === 403;
    }

    return error instanceof ApiError && AUTH_ERROR_CODES.has(error.code);
};

export const useAuthStore = create<AuthState>((set) => ({
    user: null,
    token: localStorage.getItem('token'),
    refreshToken: localStorage.getItem('refresh_token'),
    setAuth: (user, token, refreshToken) => {
        localStorage.setItem('token', token);
        if (refreshToken) {
            localStorage.setItem('refresh_token', refreshToken);
        }
        set({ user, token, refreshToken: refreshToken || null });
    },
    setTokens: (token, refreshToken) => {
        localStorage.setItem('token', token);
        localStorage.setItem('refresh_token', refreshToken);
        set({ token, refreshToken });
    },
    clearAuth: () => {
        resetAuthState(set);
    },
    loadUser: async () => {
        try {
            const user = await usersApi.getMe();
            set({ user });
        } catch (error) {
            if (isAuthFailure(error)) {
                resetAuthState(set);
                return;
            }

            throw error;
        }
    },
}));
