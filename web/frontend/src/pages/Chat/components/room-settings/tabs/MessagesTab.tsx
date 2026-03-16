import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import { roomsApi } from '../../../../../services/api/rooms';
import type { RoomSettings } from '../../../../../services/api/rooms';
import { listToText, parseList } from '../types';
import type { BaseTabProps } from '../types';
import { InputRow, SelectRow } from '../components/SharedUI';

export default function MessagesTab({ roomId, canManageSettings }: BaseTabProps) {
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
                toast.error('加载设置失败');
            } finally {
                setLoading(false);
            }
        };
        void load();
    }, [roomId]);

    const handleSave = async () => {
        if (!canManageSettings) return toast.error('没有权限保存设置');
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
                <div className="text-sm font-bold text-[var(--text-main)]">消息策略</div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <InputRow
                        label="消息保留天数 (0 表示永久)" type="number"
                        value={settings.message_retention_days ?? 0} min={0}
                        onChange={(v) => setSettings(s => ({ ...s, message_retention_days: v as number }))}
                        disabled={!canManageSettings}
                    />
                    <InputRow
                        label="置顶上限" type="number"
                        value={settings.pin_limit ?? 50} min={1}
                        onChange={(v) => setSettings(s => ({ ...s, pin_limit: v as number }))}
                        disabled={!canManageSettings}
                    />
                    <InputRow
                        label="慢速模式 (秒)" type="number"
                        value={settings.slow_mode_seconds ?? 0} min={0} max={300}
                        onChange={(v) => setSettings(s => ({ ...s, slow_mode_seconds: v as number }))}
                        disabled={!canManageSettings}
                    />
                </div>
            </section>

            <section className="bg-[var(--bg-main)] rounded-xl border border-[var(--bg-sidebar)] p-5 space-y-4">
                <div className="text-sm font-bold text-[var(--text-main)]">通知与免打扰</div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <SelectRow
                        label="通知模式"
                        value={settings.notify_mode ?? 'all'}
                        options={[
                            { value: 'all', label: '全部通知' },
                            { value: 'mentions', label: '仅@和关键词' },
                            { value: 'none', label: '静默' },
                        ]}
                        disabled={!canManageSettings}
                        onChange={(v) => setSettings(s => ({ ...s, notify_mode: v as RoomSettings['notify_mode'] }))}
                    />
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">通知关键词</label>
                        <textarea
                            value={listToText(settings.notify_keywords)}
                            onChange={(e) => setSettings(s => ({ ...s, notify_keywords: parseList(e.target.value) }))}
                            className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none resize-none h-20"
                            disabled={!canManageSettings}
                            placeholder="每行一个关键词"
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">免打扰开始</label>
                        <input
                            type="time"
                            value={settings.dnd_start || ''}
                            onChange={(e) => setSettings(s => ({ ...s, dnd_start: e.target.value || null }))}
                            className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none"
                            disabled={!canManageSettings}
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">免打扰结束</label>
                        <input
                            type="time"
                            value={settings.dnd_end || ''}
                            onChange={(e) => setSettings(s => ({ ...s, dnd_end: e.target.value || null }))}
                            className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none"
                            disabled={!canManageSettings}
                        />
                    </div>
                </div>
                <div className="flex items-center justify-end">
                    <button
                        type="button"
                        disabled={!canManageSettings || isSaving || saveSuccess}
                        onClick={handleSave}
                        className={`px-4 py-2 rounded-md font-semibold text-sm transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)] disabled:opacity-60 disabled:active:scale-100 ${saveSuccess ? 'bg-green-600 hover:bg-green-500 text-white' : 'bg-[var(--accent)] hover:bg-[#5b4eb3] text-white'}`}
                    >
                        {isSaving ? '保存中...' : saveSuccess ? '已保存' : '保存设置'}
                    </button>
                </div>
            </section>
        </div>
    );
}
