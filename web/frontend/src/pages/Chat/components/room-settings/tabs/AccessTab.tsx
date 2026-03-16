import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import { roomsApi } from '../../../../../services/api/rooms';
import type { RoomSettings } from '../../../../../services/api/rooms';
import type { BaseTabProps } from '../types';
import { ToggleRow } from '../components/SharedUI';

export default function AccessTab({ roomId, canManageSettings }: BaseTabProps) {
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
                toast.error('获取设置失败');
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
                <div className="text-sm font-bold text-[var(--text-main)]">加入与可见性</div>
                <div className="space-y-4">
                    <ToggleRow
                        title="公开频道" desc="公开频道可被发现，私有频道仅能被邀请"
                        checked={!!settings.is_public} disabled={!canManageSettings}
                        onChange={(v) => setSettings(s => ({ ...s, is_public: v }))}
                    />
                    <ToggleRow
                        title="加入需审批" desc="开启后，加入频道需要管理员审批"
                        checked={!!settings.require_approval} disabled={!canManageSettings}
                        onChange={(v) => setSettings(s => ({ ...s, require_approval: v }))}
                    />
                    <ToggleRow
                        title="只读模式" desc="开启后仅允许阅读，禁止发言"
                        checked={!!settings.read_only} disabled={!canManageSettings}
                        onChange={(v) => setSettings(s => ({ ...s, read_only: v }))}
                    />
                    <ToggleRow
                        title="归档频道" desc="归档后频道仍可查看历史，但不允许发送"
                        checked={!!settings.archived} disabled={!canManageSettings}
                        onChange={(v) => setSettings(s => ({ ...s, archived: v }))}
                    />
                </div>
                <div className="flex items-center justify-end">
                    <button type="button" disabled={!canManageSettings || isSaving || saveSuccess} onClick={handleSave} className={`px-4 py-2 rounded-md font-semibold text-sm transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)] disabled:opacity-60 disabled:active:scale-100 ${saveSuccess ? 'bg-green-600 hover:bg-green-500 text-white' : 'bg-[var(--accent)] hover:bg-[#5b4eb3] text-white'}`}>
                        {isSaving ? '保存中...' : saveSuccess ? '已保存' : '保存设置'}
                    </button>
                </div>
            </section>
        </div>
    );
}

