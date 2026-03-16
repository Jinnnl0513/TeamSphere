import { useState } from 'react';
import { useAuthStore } from '../../../../stores/authStore';
import { usersApi } from '../../../../services/api/users';

export default function ProfileSettings() {
    const { user, loadUser } = useAuthStore();
    const [bio, setBio] = useState(user?.bio || '');
    const [profileColor, setProfileColor] = useState(user?.profile_color || '#6c5dd3');
    const [isSaving, setIsSaving] = useState(false);
    const [saveStatus, setSaveStatus] = useState<'idle' | 'success' | 'error'>('idle');

    const handleSave = async () => {
        setIsSaving(true);
        setSaveStatus('idle');
        try {
            await usersApi.updateProfile({
                bio: bio,
                profile_color: profileColor,
            });
            await loadUser();
            setSaveStatus('success');
            setTimeout(() => setSaveStatus('idle'), 2000);
        } catch (error) {
            console.error('Failed to update profile:', error);
            setSaveStatus('error');
        } finally {
            setIsSaving(false);
        }
    };

    return (
        <div className="animate-in slide-in-from-right-4 duration-300">
            <h2 className="text-xl font-bold text-[var(--text-main)] mb-6">用户资料主题</h2>
            <div className="flex flex-col md:flex-row gap-8">
                {/* Form Left */}
                <div className="flex-1 space-y-6">
                    {/* About Me */}
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">关于我</label>
                        <textarea
                            value={bio}
                            onChange={e => setBio(e.target.value)}
                            className="w-full bg-[var(--bg-secondary)] text-[var(--text-main)] p-3 rounded-md border-none focus:ring-1 focus:ring-[var(--accent)] outline-none resize-none h-24 text-sm"
                            placeholder="你能在这里写一段关于你自己的简短介绍..."
                        />
                        <p className="text-[11px] text-[var(--text-muted)] mt-1">你的群友将会在你的个人名片上看到它。</p>
                    </div>
                    {/* Custom Accent Color */}
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">名片主题色</label>
                        <div className="flex space-x-3">
                            {['#6c5dd3', '#ff6b6b', '#20c997', '#f39c12', '#3498db', '#9b59b6'].map(color => (
                                <div
                                    key={color}
                                    onClick={() => setProfileColor(color)}
                                    className={`w-8 h-8 rounded-full cursor-pointer transition-transform hover:scale-110 ${profileColor === color ? 'ring-2 ring-offset-2 ring-offset-[var(--bg-main)] ring-[var(--text-main)]' : 'ring-1 ring-black/10'}`}
                                    style={{ backgroundColor: color }}
                                />
                            ))}
                        </div>
                    </div>
                    <div className="pt-4 border-t border-[var(--text-muted)]/10 flex items-center gap-4">
                        <button
                            onClick={handleSave}
                            disabled={isSaving}
                            className={`bg-[var(--accent)] hover:bg-[#5b4eb3] disabled:opacity-50 disabled:cursor-not-allowed text-white px-5 py-2 rounded-md font-semibold text-sm transition-colors shadow-sm w-full md:w-auto ${saveStatus === 'success' ? 'bg-green-600 hover:bg-green-700' : ''}`}
                        >
                            {isSaving ? '保存中...' : saveStatus === 'success' ? '已保存！' : '保存设置'}
                        </button>
                        {saveStatus === 'error' && (
                            <span className="text-red-500 text-sm">保存失败，请稍后重试</span>
                        )}
                    </div>
                </div>

                {/* Preview Card Right */}
                <div className="w-72 shrink-0">
                    <div className="text-xs font-bold text-[var(--text-muted)] uppercase mb-2">实时效果预览</div>
                    <div className="w-full rounded-2xl overflow-hidden shadow-xl bg-[var(--bg-secondary)] border border-[var(--text-muted)]/10 relative">
                        <div className="h-20" style={{ backgroundColor: profileColor }}></div>
                        <div className="absolute top-12 left-4 w-16 h-16 rounded-full border-[6px] border-[var(--bg-secondary)] overflow-hidden bg-[var(--bg-main)] flex items-center justify-center text-white text-2xl font-bold z-10">
                            <img src={user?.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${user?.username}`} className="w-full h-full object-cover" />
                        </div>
                        <div className="absolute top-[85px] left-[55px] w-[14px] h-[14px] bg-[var(--color-discord-green-500)] border-[3px] border-[var(--bg-secondary)] rounded-full z-20"></div>

                        <div className="pt-10 px-4 pb-4 bg-[var(--bg-secondary)]">
                            <h3 className="font-bold text-[var(--text-main)] text-lg leading-tight">{user?.username}</h3>
                            <p className="text-xs text-[var(--text-muted)] mb-3">{user?.username}</p>

                            <div className="w-full h-[1px] bg-[var(--text-muted)]/10 mb-3"></div>

                            <h4 className="text-[10px] font-bold text-[var(--text-muted)] uppercase mb-2">关于我</h4>
                            <p className="text-[13px] text-[var(--text-main)] italic relative pl-2.5 min-h-[40px]">
                                <span className="absolute left-0 top-0 bottom-0 w-[3px] rounded" style={{ backgroundColor: profileColor }}></span>
                                {bio || "我还没有构思好华丽的自白。"}
                            </p>

                            <button
                                className="w-full mt-5 bg-[var(--text-muted)]/10 hover:bg-[var(--text-muted)]/20 text-xs font-bold py-2 rounded-md transition-colors"
                                style={{ color: profileColor }}
                            >
                                发消息
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
