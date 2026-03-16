import { create } from 'zustand';

type Theme = 'system' | 'light' | 'dark';

interface AppearanceState {
    theme: Theme;
    setTheme: (theme: Theme) => void;
    initTheme: () => void;
}

export const useAppearanceStore = create<AppearanceState>((set) => ({
    theme: (localStorage.getItem('app-theme') as Theme) || 'system',

    setTheme: (theme: Theme) => {
        localStorage.setItem('app-theme', theme);
        set({ theme });
        applyTheme(theme);
    },

    initTheme: () => {
        const theme = (localStorage.getItem('app-theme') as Theme) || 'system';
        applyTheme(theme);
    }
}));

function getSystemTheme(): 'light' | 'dark' {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function applyTheme(theme: Theme) {
    const root = document.documentElement;
    root.classList.remove('theme-light', 'theme-dark');
    const resolved = theme === 'system' ? getSystemTheme() : theme;
    root.classList.add(resolved === 'dark' ? 'theme-dark' : 'theme-light');
}

if (typeof window !== 'undefined') {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        const current = useAppearanceStore.getState().theme;
        if (current === 'system') applyTheme('system');
    });
}
