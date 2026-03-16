import { useCallback, useEffect, useState } from 'react';
import { Save, Loader2, Megaphone } from 'lucide-react';
import { adminApi } from '../../../services/api/admin';

export default function AnnouncementView() {
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [announcement, setAnnouncement] = useState('');

    const fetchSettings = useCallback(async () => {
        setLoading(true);
        try {
            const data = await adminApi.getAnnouncement();
            setAnnouncement(data.content ?? '');
        } catch (err) {
            console.error('Failed to fetch announcement', err);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchSettings();
    }, [fetchSettings]);

    const handleSave = async (e: React.FormEvent) => {
        e.preventDefault();
        setSaving(true);
        try {
            await adminApi.setAnnouncement(announcement);
            alert('系统公告保存成功！');
            await fetchSettings(); // refresh
        } catch (err) {
            console.error('Failed to save announcement', err);
            alert('系统公告保存失败。');
        } finally {
            setSaving(false);
        }
    };

    if (loading) {
        return (
            <div className="flex justify-center items-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-[var(--accent)]" />
            </div>
        );
    }

    if (!loading && announcement === null) {
        return <div className="text-red-500">加载系统公告失败。</div>;
    }

    return (
        <div className="max-w-3xl space-y-6 animate-in fade-in duration-500">
            <div>
                <h2 className="text-2xl font-bold text-[var(--text-main)]">系统公告管理</h2>
                <p className="mt-1 text-sm text-[var(--text-secondary)]">
                    编辑顶部系统公告，会在所有用户的聊天页面顶部展示。
                </p>
            </div>

            <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 flex items-start">
                <Megaphone className="h-5 w-5 text-blue-500 mt-0.5 mr-3 flex-shrink-0" />
                <div className="text-sm text-blue-800 dark:text-blue-300">
                    留空则表示不显示系统公告。系统公告支持直接输入纯文本。
                </div>
            </div>

            <form onSubmit={handleSave} className="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-xl shadow-sm">
                <div className="p-6 space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-[var(--text-main)] mb-1">
                            系统公告内容
                        </label>
                        <p className="text-xs text-[var(--text-secondary)] mb-2">
                            留空以清除公告
                        </p>
                        <textarea
                            rows={6}
                            className="w-full bg-[var(--bg-secondary)] border border-[var(--border-color)] text-[var(--text-main)] text-sm rounded-lg focus:ring-[var(--accent)] focus:border-[var(--accent)] block p-2.5 transition-colors resize-y"
                            value={announcement}
                            onChange={(e) => setAnnouncement(e.target.value)}
                            placeholder="请输入希望告诉大家的系统公告内容..."
                        />
                    </div>
                </div>

                <div className="px-6 py-4 bg-[var(--bg-secondary)] border-t border-[var(--border-color)] rounded-b-xl flex justify-end">
                    <button
                        type="submit"
                        disabled={saving}
                        className="flex items-center text-white bg-[var(--accent)] hover:bg-opacity-90 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-6 py-2.5 disabled:opacity-50 transition-colors"
                    >
                        {saving ? <Loader2 className="w-5 h-5 mr-2 animate-spin" /> : <Save className="w-5 h-5 mr-2" />}
                        保存公告
                    </button>
                </div>
            </form>
        </div>
    );
}
