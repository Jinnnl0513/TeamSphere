import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import { roomsApi } from '../../../../../services/api/rooms';
import type { RoomJoinRequest } from '../../../../../services/api/rooms';
import type { BaseTabProps } from '../types';

export default function JoinRequestsTab({ roomId, canManageMembers }: BaseTabProps) {
    const [joinRequests, setJoinRequests] = useState<RoomJoinRequest[]>([]);
    const [loading, setLoading] = useState(true);
    const [isProcessingJoin, setIsProcessingJoin] = useState(false);

    useEffect(() => {
        if (!canManageMembers) {
            setLoading(false);
            return;
        }
        const load = async () => {
            try {
                const res = await roomsApi.listJoinRequests(roomId);
                if (Array.isArray(res)) setJoinRequests(res);
            } catch (err) {
                console.error('Failed to load join requests', err);
            } finally {
                setLoading(false);
            }
        };
        void load();
    }, [roomId, canManageMembers]);

    const handleApproveJoin = async (reqId: number) => {
        setIsProcessingJoin(true);
        try {
            await roomsApi.approveJoinRequest(roomId, reqId);
            setJoinRequests(prev => prev.filter(r => r.id !== reqId));
            toast.success('已通过申请');
        } catch (err: any) {
            toast.error(err?.message || '审批失败');
        } finally {
            setIsProcessingJoin(false);
        }
    };

    const handleRejectJoin = async (reqId: number) => {
        setIsProcessingJoin(true);
        try {
            await roomsApi.rejectJoinRequest(roomId, reqId);
            setJoinRequests(prev => prev.filter(r => r.id !== reqId));
            toast.success('已拒绝申请');
        } catch (err: any) {
            toast.error(err?.message || '操作失败');
        } finally {
            setIsProcessingJoin(false);
        }
    };

    if (loading) return <div className="text-sm text-[var(--text-muted)] p-5">加载中...</div>;

    return (
        <div className="space-y-6">
            <section className="bg-[var(--bg-main)] rounded-xl border border-[var(--bg-sidebar)] p-5">
                <div className="text-sm font-bold text-[var(--text-main)] mb-3">加入申请</div>
                {!canManageMembers && (
                    <div className="text-sm text-[var(--text-muted)]">仅群主或管理员可查看申请。</div>
                )}
                {canManageMembers && joinRequests.length === 0 && (
                    <div className="text-sm text-[var(--text-muted)]">暂无待审批申请。</div>
                )}
                {canManageMembers && joinRequests.length > 0 && (
                    <div className="space-y-3">
                        {joinRequests.map(req => (
                            <div key={req.id} className="flex flex-col md:flex-row md:items-center justify-between gap-3 bg-[var(--bg-secondary)] border border-[var(--bg-sidebar)] rounded-lg p-4">
                                <div>
                                    <div className="text-sm font-semibold text-[var(--text-main)]">用户ID: {req.user_id}</div>
                                    <div className="text-xs text-[var(--text-muted)]">申请时间: {new Date(req.created_at).toLocaleString()}</div>
                                    {req.reason && <div className="text-xs text-[var(--text-muted)]">原因: {req.reason}</div>}
                                </div>
                                <div className="flex items-center gap-2">
                                    <button
                                        onClick={() => handleApproveJoin(req.id)}
                                        disabled={isProcessingJoin}
                                        className="px-3 py-1.5 rounded bg-[var(--accent)] text-white text-xs font-semibold disabled:opacity-60 transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)] disabled:active:scale-100"
                                    >
                                        通过
                                    </button>
                                    <button
                                        onClick={() => handleRejectJoin(req.id)}
                                        disabled={isProcessingJoin}
                                        className="px-3 py-1.5 rounded bg-[var(--bg-hover)] text-[var(--text-main)] text-xs font-semibold disabled:opacity-60 transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[var(--text-muted)] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)] disabled:active:scale-100"
                                    >
                                        拒绝
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </section>
        </div>
    );
}
