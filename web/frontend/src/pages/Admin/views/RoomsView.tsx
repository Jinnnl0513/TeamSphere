import { useEffect, useState } from 'react';
import { Trash2, Search, Loader2 } from 'lucide-react';
import { adminApi, type RoomsResponse } from '../../../services/api/admin';

export default function RoomsView() {
    const [data, setData] = useState<RoomsResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [page, setPage] = useState(1);
    const [search, setSearch] = useState('');

    const [actionLoading, setActionLoading] = useState<number | null>(null);

    const fetchRooms = async (pageNum: number) => {
        setLoading(true);
        try {
            const res = await adminApi.listRooms(pageNum, 20);
            setData(res);
        } catch (err) {
            console.error('Failed to fetch rooms', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchRooms(page);
    }, [page]);

    const handleDeleteRoom = async (roomId: number, roomName: string) => {
        if (!window.confirm(`警告：此操作不可恢复。确定要彻底删除房间“${roomName}”吗？`)) return;

        setActionLoading(roomId);
        try {
            await adminApi.deleteRoom(roomId);
            await fetchRooms(page); // refresh
        } catch (err) {
            console.error('Failed to delete room', err);
            alert('删除房间失败');
        } finally {
            setActionLoading(null);
        }
    };

    const filteredRooms = data?.rooms.filter(r =>
        r.name.toLowerCase().includes(search.toLowerCase())
    ) || [];

    return (
        <div className="space-y-6 animate-in fade-in duration-500">
            <div className="flex justify-between items-center">
                <h2 className="text-2xl font-bold text-[var(--text-main)]">房间管理</h2>
                <div className="relative">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                        <Search className="h-5 w-5 text-[var(--text-secondary)]" />
                    </div>
                    <input
                        type="text"
                        placeholder="搜索房间..."
                        className="block w-64 pl-10 pr-3 py-2 border border-[var(--border-color)] rounded-md leading-5 bg-[var(--bg-main)] text-[var(--text-main)] placeholder-[var(--text-secondary)] focus:outline-none focus:ring-1 focus:ring-[var(--accent)] sm:text-sm transition-colors"
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                    />
                </div>
            </div>

            <div className="bg-[var(--bg-main)] border border-[var(--border-color)] rounded-xl overflow-hidden shadow-sm">
                <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-[var(--border-color)]">
                        <thead className="bg-[var(--bg-secondary)]">
                            <tr>
                                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-[var(--text-secondary)] uppercase tracking-wider">房间名</th>
                                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-[var(--text-secondary)] uppercase tracking-wider">成员数</th>
                                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-[var(--text-secondary)] uppercase tracking-wider">创建时间</th>
                                <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-[var(--text-secondary)] uppercase tracking-wider">操作</th>
                            </tr>
                        </thead>
                        <tbody className="bg-[var(--bg-main)] divide-y divide-[var(--border-color)]">
                            {loading && !data ? (
                                <tr>
                                    <td colSpan={4} className="px-6 py-12 text-center text-[var(--text-secondary)]">
                                        <Loader2 className="h-8 w-8 animate-spin mx-auto text-[var(--accent)]" />
                                    </td>
                                </tr>
                            ) : filteredRooms.length === 0 ? (
                                <tr>
                                    <td colSpan={4} className="px-6 py-12 text-center text-[var(--text-secondary)]">
                                        未找到房间。
                                    </td>
                                </tr>
                            ) : (
                                filteredRooms.map((room) => (
                                    <tr key={room.id} className="hover:bg-[var(--bg-hover)] transition-colors">
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="flex items-center">
                                                <div className="flex-shrink-0 h-10 w-10 flex items-center justify-center bg-[var(--bg-secondary)] text-[var(--text-secondary)] rounded-xl font-bold">
                                                    {room.name.substring(0, 2).toUpperCase()}
                                                </div>
                                                <div className="ml-4">
                                                    <div className="text-sm font-medium text-[var(--text-main)] group-hover:text-[var(--accent)] transition-colors">
                                                        {room.name}
                                                    </div>
                                                    <div className="text-sm text-[var(--text-secondary)] truncate w-48">
                                                        {room.description || '暂无描述'}
                                                    </div>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-[var(--bg-secondary)] text-[var(--text-main)]">
                                                {room.member_count} 人
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-[var(--text-secondary)]">
                                            {new Date(room.created_at).toLocaleDateString()}
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                            <button
                                                onClick={() => handleDeleteRoom(room.id, room.name)}
                                                disabled={actionLoading === room.id}
                                                className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 disabled:opacity-50 transition-colors"
                                                title="删除房间"
                                            >
                                                <Trash2 className="h-5 w-5 inline-block" />
                                            </button>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>

                {data && data.total_pages > 1 && (
                    <div className="px-6 py-3 border-t border-[var(--border-color)] flex items-center justify-between">
                        <button
                            disabled={page <= 1}
                            onClick={() => setPage(p => p - 1)}
                            className="text-sm text-[var(--text-secondary)] hover:text-[var(--text-main)] disabled:opacity-50"
                        >
                            上一页
                        </button>
                        <span className="text-sm text-[var(--text-secondary)]">第 {page} 页 / 共 {data.total_pages} 页</span>
                        <button
                            disabled={page >= data.total_pages}
                            onClick={() => setPage(p => p + 1)}
                            className="text-sm text-[var(--text-secondary)] hover:text-[var(--text-main)] disabled:opacity-50"
                        >
                            下一页
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}
