export default function FullScreenLoader() {
  return (
    <div className="fixed inset-0 flex flex-col items-center justify-center bg-[var(--bg-main)] z-50">
      <div className="relative flex items-center justify-center">
        <div className="absolute w-24 h-24 rounded-full border-[3px] border-[var(--accent)] border-t-transparent opacity-60 animate-spin"></div>
        <div className="absolute inset-0 bg-[var(--accent)] opacity-20 blur-xl rounded-full animate-pulse"></div>
        <div className="relative flex items-center justify-center w-16 h-16 bg-[var(--bg-card)] border border-[var(--border-light)] rounded-2xl shadow-xl z-10">
          <span className="text-3xl drop-shadow-md animate-bounce">聊</span>
        </div>
      </div>
      <div className="mt-8 flex flex-col items-center gap-2">
        <div className="flex items-center gap-1.5">
          <div className="w-2.5 h-2.5 rounded-full bg-[var(--accent)] animate-bounce" style={{ animationDelay: '0ms' }}></div>
          <div className="w-2.5 h-2.5 rounded-full bg-[var(--accent)] animate-bounce" style={{ animationDelay: '150ms' }}></div>
          <div className="w-2.5 h-2.5 rounded-full bg-[var(--accent)] animate-bounce" style={{ animationDelay: '300ms' }}></div>
        </div>
        <p className="text-[var(--text-muted)] text-sm font-medium tracking-widest uppercase mt-2 opacity-80 animate-pulse">
          加载中
        </p>
      </div>
    </div>
  );
}
