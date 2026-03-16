import { useAppearanceStore } from '../../../../stores/appearanceStore';

export default function AppearanceSettings() {
    const { theme, setTheme } = useAppearanceStore();

    return (
        <div className="space-y-6 animate-in slide-in-from-right-4 duration-300">
            <h2 className="text-xl font-bold text-[var(--text-main)] mb-6">外观</h2>

            <div className="space-y-4">
                <h3 className="text-xs font-bold text-[var(--text-muted)] uppercase">应用主题</h3>
                <p className="text-sm text-[var(--text-muted)] mb-2">由跟随操作系统首选项(Prefers-Color-Scheme)提供原生的深色与浅色支援。</p>

                <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mt-6">
                    {/* System Theme Toggle */}
                    <div onClick={() => setTheme('system')} className={`border-2 ${theme === 'system' ? 'border-[var(--accent)]' : 'border-[var(--bg-sidebar)]'} bg-[var(--bg-secondary)] rounded-xl p-3 cursor-pointer relative overflow-hidden transition-all hover:bg-[var(--bg-hover)]`}>
                        {theme === 'system' && (
                            <div className="absolute top-2 right-2 w-4 h-4 bg-[var(--accent)] rounded-full border-2 border-[var(--bg-main)] flex items-center justify-center shadow-sm">
                                <svg className="w-2.5 h-2.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" /></svg>
                            </div>
                        )}
                        <div className="h-10 bg-[var(--bg-main)] rounded mb-2 w-full flex space-x-2 p-1 border border-[var(--text-muted)]/10">
                            <div className="w-2/12 h-full bg-[var(--bg-hover)] rounded border border-[var(--text-muted)]/10"></div>
                            <div className="w-10/12 h-full bg-[var(--accent)]/20 rounded border border-[var(--text-muted)]/10"></div>
                        </div>
                        <div className="text-center text-sm font-semibold text-[var(--text-main)]">跟随系统</div>
                    </div>

                    {/* Light Theme Toggle */}
                    <div onClick={() => setTheme('light')} className={`border-2 ${theme === 'light' ? 'border-[var(--accent)]' : 'border-[#e9e7e4]'} bg-[#f5f3f0] rounded-xl p-3 cursor-pointer relative overflow-hidden transition-all hover:bg-[#f1ede8]`}>
                        {theme === 'light' && (
                            <div className="absolute top-2 right-2 w-4 h-4 bg-[var(--accent)] rounded-full border-2 border-white flex items-center justify-center shadow-sm">
                                <svg className="w-2.5 h-2.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" /></svg>
                            </div>
                        )}
                        <div className="h-10 bg-white rounded mb-2 w-full flex space-x-2 p-1 border border-black/5">
                            <div className="w-2/12 h-full bg-[#e9e7e4] rounded border border-black/5"></div>
                            <div className="w-10/12 h-full bg-[var(--accent)]/20 rounded border border-black/5"></div>
                        </div>
                        <div className="text-center text-sm font-semibold text-[#1a1a1a]">纯亮色模式</div>
                    </div>

                    {/* Dark Theme Toggle */}
                    <div onClick={() => setTheme('dark')} className={`border-2 ${theme === 'dark' ? 'border-[var(--accent)]' : 'border-[#141517]'} bg-[#1e1f22] rounded-xl p-3 cursor-pointer relative overflow-hidden transition-all hover:bg-[#2b2d31]`}>
                        {theme === 'dark' && (
                            <div className="absolute top-2 right-2 w-4 h-4 bg-[var(--accent)] rounded-full border-2 border-[#1e1f22] flex items-center justify-center shadow-sm">
                                <svg className="w-2.5 h-2.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" /></svg>
                            </div>
                        )}
                        <div className="h-10 bg-[#2b2d31] rounded mb-2 w-full flex space-x-2 p-1 border border-white/5">
                            <div className="w-2/12 h-full bg-[#141517] rounded border border-white/5"></div>
                            <div className="w-10/12 h-full bg-[var(--accent)]/20 rounded border border-white/5"></div>
                        </div>
                        <div className="text-center text-sm font-semibold text-[#f5f5f5]">深色模式</div>
                    </div>
                </div>
            </div>
        </div>
    );
}
