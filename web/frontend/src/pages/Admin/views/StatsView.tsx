import { useEffect, useState } from 'react';
import {
    Users,
    MessageSquare,
    MessageCircle
} from 'lucide-react';
import { adminApi, type Stats } from '../../../services/api/admin';

export default function StatsView() {
    const [stats, setStats] = useState<Stats | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        adminApi.getStats()
            .then(data => setStats(data))
            .catch(err => console.error('Failed to fetch stats:', err))
            .finally(() => setLoading(false));
    }, []);

    if (loading) {
        return (
            <div className="animate-pulse space-y-6">
                <div className="h-8 bg-[var(--bg-secondary)] rounded w-64"></div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    {[1, 2, 3].map(i => (
                        <div key={i} className="bg-[var(--bg-secondary)] border border-[var(--border-color)] rounded-xl h-28"></div>
                    ))}
                </div>
            </div>
        );
    }

    if (!stats) {
        return <div className="text-red-500">加载统计数据失败</div>;
    }

    const cards = [
        { name: '用户总数', value: stats.total_users, icon: Users, color: 'text-blue-500', bg: 'bg-blue-500/10' },
        { name: '房间总数', value: stats.total_rooms, icon: MessageCircle, color: 'text-green-500', bg: 'bg-green-500/10' },
        { name: '消息总数', value: stats.total_messages, icon: MessageSquare, color: 'text-purple-500', bg: 'bg-purple-500/10' },
    ];

    return (
        <div className="space-y-6 animate-in fade-in duration-500">
            <h2 className="text-2xl font-bold text-[var(--text-main)]">数据总览</h2>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {cards.map((card) => {
                    const Icon = card.icon;
                    return (
                        <div key={card.name} className="bg-[var(--bg-main)] border border-[var(--border-color)] rounded-xl p-6 flex items-center shadow-sm hover:shadow-md transition-shadow">
                            <div className={`p-4 rounded-full ${card.bg} mr-4`}>
                                <Icon className={`h-8 w-8 ${card.color}`} />
                            </div>
                            <div>
                                <p className="text-sm font-medium text-[var(--text-secondary)]">{card.name}</p>
                                <p className="text-3xl font-bold text-[var(--text-main)] mt-1">{(card.value ?? 0).toLocaleString()}</p>
                            </div>
                        </div>
                    );
                })}
            </div>
            {/* Can add charts or recent activity here later */}
        </div>
    );
}
