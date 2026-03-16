import { useState, useEffect } from 'react';
import { roomsApi } from '../../../../../services/api/rooms';
import type { RoomStatsSummary } from '../../../../../services/api/rooms';
import type { BaseTabProps } from '../types';
import { InputRow, StatCard } from '../components/SharedUI';

export default function StatsTab({ roomId, canManageMembers }: BaseTabProps) {
    const [stats, setStats] = useState<RoomStatsSummary | null>(null);
    const [statsDays, setStatsDays] = useState(7);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        if (!canManageMembers) {
            setLoading(false);
            return;
        }
        const load = async () => {
            setLoading(true);
            try {
                const res = await roomsApi.getStatsSummary(roomId, statsDays);
                if (res) setStats(res);
            } catch (err) {
                console.error('Failed to load stats summary', err);
            } finally {
                setLoading(false);
            }
        };
        void load();
    }, [roomId, canManageMembers, statsDays]);

    if (!canManageMembers) {
        return (
            <div className="p-5">
                <div className="text-sm text-[var(--text-muted)]">仅群主或管理员可查看统计数据。</div>
            </div>
        );
    }

    if (loading && !stats) return <div className="text-sm text-[var(--text-muted)] p-5">加载中...</div>;

    return (
        <div className="space-y-6">
            <section className="bg-[var(--bg-main)] rounded-xl border border-[var(--bg-sidebar)] p-5">
                <div className="text-sm font-bold text-[var(--text-main)] mb-3">数据概览</div>
                <div className="space-y-4">
                    <div className="flex flex-col sm:flex-row gap-3 items-center">
                        <InputRow
                            label="统计天数"
                            type="number"
                            value={statsDays}
                            min={1}
                            onChange={(v) => setStatsDays(v as number)}
                        />
                    </div>
                    {loading ? (
                        <div className="text-sm text-[var(--text-muted)]">加载中...</div>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <StatCard title="消息总量" value={stats?.total_messages ?? 0} />
                            <StatCard title="活跃用户" value={stats?.active_users ?? 0} />
                        </div>
                    )}
                </div>
            </section>
        </div>
    );
}
