
import { useNotificationStore } from '../../../../stores/notificationStore';

export default function NotificationSettings() {
    const {
        desktopEnabled, soundEnabled, permissionState,
        setDesktopEnabled, setSoundEnabled, requestPermission
    } = useNotificationStore();

    return (
        <div className="space-y-6 animate-in slide-in-from-right-4 duration-300">
            <h2 className="text-xl font-bold text-[var(--text-main)] mb-6">通知设置</h2>
            <p className="text-[var(--text-muted)] text-sm mb-4">掌握您接收哪些应用内和跨平台的声音与弹窗消息。</p>

            <div className="bg-[var(--bg-secondary)] border border-[var(--text-muted)]/10 rounded-xl p-4 space-y-4">
                <div className="flex flex-col gap-2">
                    <div className="flex items-center justify-between">
                        <div>
                            <div className="text-[var(--text-main)] font-semibold text-sm flex items-center gap-2">
                                桌面消息推送
                                {permissionState === 'denied' && (
                                    <span className="text-[10px] bg-[var(--color-discord-red-400)]/10 text-[var(--color-discord-red-400)] px-2 py-0.5 rounded font-bold">已被系统禁用</span>
                                )}
                            </div>
                            <div className="text-[var(--text-muted)] text-xs mt-1">当您不在当前窗口但仍在线时是否弹窗提醒</div>
                        </div>
                        <div className="flex items-center gap-4">
                            {desktopEnabled && permissionState === 'granted' && (
                                <button
                                    onClick={() => {
                                        import('../../../../stores/notificationStore').then(({ sendDesktopNotification }) => {
                                            sendDesktopNotification('这是一条测试通知', '如果您看到了它，说明桌面通知工作正常！');
                                        });
                                    }}
                                    className="text-xs px-3 py-1 bg-[var(--accent)]/10 text-[var(--accent)] hover:bg-[var(--accent)] hover:text-white transition-colors rounded-md"
                                >
                                    测试通知
                                </button>
                            )}
                            <button
                                onClick={() => {
                                    if (permissionState === 'default') {
                                        requestPermission();
                                    } else if (permissionState === 'granted') {
                                        setDesktopEnabled(!desktopEnabled);
                                    } else {
                                        alert('您已在浏览器设置中禁用了通知，请前往浏览器设置打开权限。');
                                    }
                                }}
                                className={`w-10 h-5 rounded-full relative shadow-inner cursor-pointer transition-colors ${desktopEnabled && permissionState === 'granted' ? 'bg-[#20c997]' : 'bg-[var(--text-muted)]/30'}`}
                            >
                                <div className={`w-4 h-4 rounded-full bg-white absolute top-0.5 shadow-sm transition-all ${desktopEnabled && permissionState === 'granted' ? 'right-0.5' : 'left-0.5'}`}></div>
                            </button>
                        </div>
                    </div>
                    {desktopEnabled && permissionState === 'granted' && (
                        <div className="text-[11px] text-[var(--text-muted)]/80 mt-1 bg-[var(--bg-main)] p-2 rounded border border-[var(--text-muted)]/10">
                            <strong>收不到通知？</strong> 请确保没有处于全屏应用中，并检查 Windows 的“通知和操作”设置是否开启，以及是否禁用了“专注助手 / 勿扰模式”。
                        </div>
                    )}
                </div>

                <div className="w-full h-[1px] bg-[var(--text-muted)]/10"></div>

                <div className="flex items-center justify-between">
                    <div>
                        <div className="text-[var(--text-main)] font-semibold text-sm">收到新消息时声音提示</div>
                        <div className="text-[var(--text-muted)] text-xs mt-1">控制是否在收到私聊或频道的非群发消息时播放短促的提示音</div>
                    </div>
                    <button
                        onClick={() => setSoundEnabled(!soundEnabled)}
                        className={`w-10 h-5 rounded-full relative shadow-inner cursor-pointer transition-colors ${soundEnabled ? 'bg-[#20c997]' : 'bg-[var(--text-muted)]/30'}`}
                    >
                        <div className={`w-4 h-4 rounded-full bg-white absolute top-0.5 shadow-sm transition-all ${soundEnabled ? 'right-0.5' : 'left-0.5'}`}></div>
                    </button>
                </div>
            </div>
        </div>
    );
}
