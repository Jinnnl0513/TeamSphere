import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import { roomsApi } from '../../../../../services/api/rooms';
import type { RoomSettings } from '../../../../../services/api/rooms';
import { listToText, parseList } from '../types';
import type { BaseTabProps } from '../types';
import { InputRow, SelectRow, ToggleRow } from '../components/SharedUI';

export default function SafetyTab({ roomId, canManageSettings }: BaseTabProps) {
    const [settings, setSettings] = useState<Partial<RoomSettings>>({});
    const [loading, setLoading] = useState(true);
    const [isSaving, setIsSaving] = useState(false);
    const [saveSuccess, setSaveSuccess] = useState(false);

    useEffect(() => {
        const load = async () => {
            try {
                const res = await roomsApi.getSettings(roomId);
                if (res) setSettings(res);
            } catch (err) {
                toast.error('加载安全设置失败');
            } finally {
                setLoading(false);
            }
        };
        void load();
    }, [roomId]);

    const handleSave = async () => {
        if (!canManageSettings) return toast.error('没有权限');
        setIsSaving(true);
        try {
            await roomsApi.updateSettings(roomId, settings as RoomSettings);
            toast.success('设置已保存');
            setSaveSuccess(true);
            setTimeout(() => setSaveSuccess(false), 2000);
        } catch (err: any) {
            toast.error(err?.message || '保存设置失败');
        } finally {
            setIsSaving(false);
        }
    };

    if (loading) return <div className="text-sm text-[var(--text-muted)] p-5">加载中...</div>;

    return (
        <div className="space-y-6">
            <section className="bg-[var(--bg-main)] rounded-xl border border-[var(--bg-sidebar)] p-5 space-y-4">
                <div className="text-sm font-bold text-[var(--text-main)]">内容与链接过滤</div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <SelectRow
                        label="过滤模式"
                        value={settings.content_filter_mode ?? 'off'}
                        options={[
                            { value: 'off', label: '关闭' },
                            { value: 'block_log', label: '拦截并记录' },
                        ]}
                        disabled={!canManageSettings}
                        onChange={(v) => setSettings(s => ({ ...s, content_filter_mode: v as RoomSettings['content_filter_mode'] }))}
                    />
                    <InputRow
                        label="最大文件大小 (MB)" type="number"
                        value={settings.max_file_size_mb ?? 10} min={1}
                        onChange={(v) => setSettings(s => ({ ...s, max_file_size_mb: v as number }))}
                        disabled={!canManageSettings}
                    />
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">敏感词</label>
                        <textarea
                            value={listToText(settings.blocked_keywords)}
                            onChange={(e) => setSettings(s => ({ ...s, blocked_keywords: parseList(e.target.value) }))}
                            className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none resize-none h-24"
                            disabled={!canManageSettings}
                            placeholder="每行一个关键词"
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">允许的链接域名</label>
                        <textarea
                            value={listToText(settings.allowed_link_domains)}
                            onChange={(e) => setSettings(s => ({ ...s, allowed_link_domains: parseList(e.target.value) }))}
                            className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none resize-none h-24"
                            disabled={!canManageSettings}
                            placeholder="留空表示不限制"
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">禁止的链接域名</label>
                        <textarea
                            value={listToText(settings.blocked_link_domains)}
                            onChange={(e) => setSettings(s => ({ ...s, blocked_link_domains: parseList(e.target.value) }))}
                            className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none resize-none h-24"
                            disabled={!canManageSettings}
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">允许的文件类型</label>
                        <textarea
                            value={listToText(settings.allowed_file_types)}
                            onChange={(e) => setSettings(s => ({ ...s, allowed_file_types: parseList(e.target.value) }))}
                            className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none resize-none h-24"
                            disabled={!canManageSettings}
                            placeholder="例如: pdf, docx, png"
                        />
                    </div>
                </div>
            </section>

            <section className="bg-[var(--bg-main)] rounded-xl border border-[var(--bg-sidebar)] p-5 space-y-4">
                <div className="text-sm font-bold text-[var(--text-main)]">反刷与重复消息</div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <InputRow
                        label="刷屏阈值" type="number"
                        value={settings.anti_spam_rate ?? 8} min={0}
                        onChange={(v) => setSettings(s => ({ ...s, anti_spam_rate: v as number }))}
                        disabled={!canManageSettings}
                    />
                    <InputRow
                        label="时间窗口 (秒)" type="number"
                        value={settings.anti_spam_window_sec ?? 10} min={1}
                        onChange={(v) => setSettings(s => ({ ...s, anti_spam_window_sec: v as number }))}
                        disabled={!canManageSettings}
                    />
                    <ToggleRow
                        title="拦截重复消息" desc="连续发送相同内容会被拦截"
                        checked={!!settings.anti_repeat} disabled={!canManageSettings}
                        onChange={(v) => setSettings(s => ({ ...s, anti_repeat: v }))}
                    />
                </div>
                <div className="flex items-center justify-end">
                    <button
                        type="button" disabled={!canManageSettings || isSaving || saveSuccess} onClick={handleSave}
                        className={`px-4 py-2 rounded-md font-semibold text-sm transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)] disabled:opacity-60 disabled:active:scale-100 ${saveSuccess ? 'bg-green-600 hover:bg-green-500 text-white' : 'bg-[var(--accent)] hover:bg-[#5b4eb3] text-white'}`}
                    >
                        {isSaving ? '保存中...' : saveSuccess ? '已保存' : '保存设置'}
                    </button>
                </div>
            </section>
        </div>
    );
}
