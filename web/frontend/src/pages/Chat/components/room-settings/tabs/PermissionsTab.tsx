import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import { roomsApi } from '../../../../../services/api/rooms';
import type { RoomRolePermission } from '../../../../../services/api/rooms';
import { defaultPermissions } from '../types';
import type { BaseTabProps } from '../types';

export default function PermissionsTab({ roomId, canManagePermissions }: BaseTabProps) {
    const [permissions, setPermissions] = useState<RoomRolePermission[]>([]);
    const [loading, setLoading] = useState(true);
    const [isSaving, setIsSaving] = useState(false);
    const [saveSuccess, setSaveSuccess] = useState(false);

    useEffect(() => {
        const load = async () => {
            try {
                const res = await roomsApi.getPermissions(roomId);
                const ensurePermissions = (list: RoomRolePermission[]): RoomRolePermission[] => {
                    const map = new Map(list.map(item => [item.role, item]));
                    return defaultPermissions().map((item) => ({
                        ...item,
                        ...(map.get(item.role) || {}),
                        role: item.role,
                    }));
                };
                setPermissions(ensurePermissions(res || []));
            } catch (err) {
                toast.error('加载权限失败');
            } finally {
                setLoading(false);
            }
        };
        void load();
    }, [roomId]);

    const handleSave = async () => {
        if (!canManagePermissions) return toast.error('没有权限修改角色权限');
        setIsSaving(true);
        try {
            await roomsApi.updatePermissions(roomId, permissions);
            toast.success('权限已更新');
            setSaveSuccess(true);
            setTimeout(() => setSaveSuccess(false), 2000);
        } catch (err: any) {
            toast.error(err?.message || '更新权限失败');
        } finally {
            setIsSaving(false);
        }
    };

    const updatePermissionField = (role: RoomRolePermission['role'], key: keyof RoomRolePermission, value: boolean) => {
        setPermissions(prev => prev.map(item => item.role === role ? { ...item, [key]: value } as RoomRolePermission : item));
    };

    const roleLabel = (role: RoomRolePermission['role']) => {
        if (role === 'owner') return '群主';
        if (role === 'admin') return '管理员';
        return '成员';
    };

    if (loading) return <div className="text-sm text-[var(--text-muted)] p-5">加载中...</div>;

    return (
        <div className="space-y-6">
            <section className="bg-[var(--bg-main)] rounded-xl border border-[var(--bg-sidebar)] p-5">
                <div className="text-sm font-bold text-[var(--text-main)] mb-4">角色权限矩阵</div>
                <div className="overflow-x-auto">
                    <table className="min-w-full text-xs text-[var(--text-muted)]">
                        <thead className="text-[11px] uppercase text-[var(--text-muted)]">
                            <tr>
                                <th className="text-left py-2 pr-4">角色</th>
                                <th className="text-left py-2 pr-4">发言</th>
                                <th className="text-left py-2 pr-4">上传</th>
                                <th className="text-left py-2 pr-4">置顶</th>
                                <th className="text-left py-2 pr-4">管理成员</th>
                                <th className="text-left py-2 pr-4">管理设置</th>
                                <th className="text-left py-2 pr-4">管理消息</th>
                                <th className="text-left py-2">全体提醒</th>
                            </tr>
                        </thead>
                        <tbody>
                            {permissions.map((perm) => (
                                <tr key={perm.role} className="border-t border-[var(--bg-sidebar)]">
                                    <td className="py-3 pr-4 text-[var(--text-main)] font-semibold">{roleLabel(perm.role)}</td>
                                    {(['can_send', 'can_upload', 'can_pin', 'can_manage_members', 'can_manage_settings', 'can_manage_messages', 'can_mention_all'] as (keyof RoomRolePermission)[]).map((field) => (
                                        <td key={field} className="py-3 pr-4">
                                            <input
                                                type="checkbox"
                                                checked={!!perm[field]}
                                                disabled={!canManagePermissions || perm.role === 'owner'}
                                                onChange={(e) => updatePermissionField(perm.role, field, e.target.checked)}
                                                className="w-4 h-4 accent-[var(--accent)]"
                                            />
                                        </td>
                                    ))}
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
                <div className="mt-4 flex items-center justify-between">
                    <span className="text-xs text-[var(--text-muted)]">仅群主可编辑权限</span>
                    <button
                        type="button"
                        disabled={!canManagePermissions || isSaving || saveSuccess}
                        onClick={handleSave}
                        className={`px-4 py-2 rounded-md font-semibold text-sm transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)] disabled:opacity-60 disabled:active:scale-100 ${saveSuccess ? 'bg-green-600 hover:bg-green-500 text-white' : 'bg-[var(--accent)] hover:bg-[#5b4eb3] text-white'}`}
                    >
                        {isSaving ? '保存中...' : saveSuccess ? '已保存' : '保存权限'}
                    </button>
                </div>
            </section>
        </div>
    );
}
