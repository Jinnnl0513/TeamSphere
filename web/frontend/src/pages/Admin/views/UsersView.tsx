import { useEffect, useState } from 'react';
import { Shield, ShieldAlert, Trash2, Search, Loader2 } from 'lucide-react';
import { useAuthStore } from '../../../stores/authStore';
import { adminApi, type UsersResponse } from '../../../services/api/admin';

export default function UsersView() {
    const { user: currentUser } = useAuthStore();
    const [data, setData] = useState<UsersResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [page, setPage] = useState(1);
    const [search, setSearch] = useState('');

    const [actionLoading, setActionLoading] = useState<number | null>(null);

    const fetchUsers = async (pageNum: number) => {
        setLoading(true);
        try {
            // Include search param if applicable, though Backend might currently only use page & page_size
            // According to backend outline, it's just GET /admin/users?page=1&page_size=20
            const res = await adminApi.listUsers(pageNum, 20);
            setData(res);
        } catch (err) {
            console.error('Failed to fetch users', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchUsers(page);
    }, [page]);

    const handleRoleUpdate = async (userId: number, currentRole: string) => {
        if (!window.confirm('确定要修改该用户角色吗？')) return;

        const newRole = currentRole === 'user' ? 'admin' : 'user';
        setActionLoading(userId);
        try {
            await adminApi.updateUserRole(userId, newRole);
            await fetchUsers(page); // refresh
        } catch (err) {
            console.error('Failed to update role', err);
            alert('修改角色失败');
        } finally {
            setActionLoading(null);
        }
    };

    const handleDeleteUser = async (userId: number) => {
        if (!window.confirm('警告：此操作不可恢复。确定要删除该用户吗？')) return;

        setActionLoading(userId);
        try {
            await adminApi.deleteUser(userId);
            await fetchUsers(page); // refresh
        } catch (err) {
            console.error('Failed to delete user', err);
            alert('删除用户失败');
        } finally {
            setActionLoading(null);
        }
    };

    // Derived filtered users (if backend doesn't support search yet)
    const filteredUsers = data?.users.filter(u =>
        u.username.toLowerCase().includes(search.toLowerCase())
    ) || [];

    return (
        <div className="space-y-6 animate-in fade-in duration-500">
            <div className="flex justify-between items-center">
                <h2 className="text-2xl font-bold text-[var(--text-main)]">用户管理</h2>
                <div className="relative">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                        <Search className="h-5 w-5 text-[var(--text-secondary)]" />
                    </div>
                    <input
                        type="text"
                        placeholder="搜索用户..."
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
                                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-[var(--text-secondary)] uppercase tracking-wider">用户</th>
                                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-[var(--text-secondary)] uppercase tracking-wider">角色</th>
                                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-[var(--text-secondary)] uppercase tracking-wider">加入时间</th>
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
                            ) : filteredUsers.length === 0 ? (
                                <tr>
                                    <td colSpan={4} className="px-6 py-12 text-center text-[var(--text-secondary)]">
                                        未找到用户。
                                    </td>
                                </tr>
                            ) : (
                                filteredUsers.map((user) => {
                                    const isSystemAdmin = user.role === 'system_admin' || user.role === 'owner';
                                    const isMe = user.id === currentUser?.id;

                                    return (
                                        <tr key={user.id} className="hover:bg-[var(--bg-hover)] transition-colors">
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <div className="flex items-center">
                                                    <div className="flex-shrink-0 h-10 w-10">
                                                        <img className="h-10 w-10 rounded-full bg-[var(--bg-secondary)] object-cover" src={user.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${user.username}`} alt="" />
                                                    </div>
                                                    <div className="ml-4">
                                                        <div className="text-sm font-medium text-[var(--text-main)]">
                                                            {user.username} {isMe && <span className="text-xs text-[var(--accent)]">(你)</span>}
                                                        </div>
                                                    </div>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full 
                                                    ${isSystemAdmin ? 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300' :
                                                        user.role === 'admin' ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300' :
                                                            'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300'}`}>
                                                    {user.role}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap text-sm text-[var(--text-secondary)]">
                                                {new Date(user.created_at).toLocaleDateString()}
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                                <div className="flex justify-end space-x-2">
                                                    {/* Cannot modify owner/system_admin or oneself */}
                                                    {!isSystemAdmin && !isMe && (
                                                        <>
                                                            <button
                                                                onClick={() => handleRoleUpdate(user.id, user.role)}
                                                                disabled={actionLoading === user.id}
                                                                className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 disabled:opacity-50 transition-colors"
                                                                title={user.role === 'admin' ? '降级为普通用户' : '提升为管理员'}
                                                            >
                                                                {user.role === 'admin' ? <ShieldAlert className="h-5 w-5" /> : <Shield className="h-5 w-5" />}
                                                            </button>
                                                            <button
                                                                onClick={() => handleDeleteUser(user.id)}
                                                                disabled={actionLoading === user.id}
                                                                className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 disabled:opacity-50 transition-colors"
                                                                title="删除用户"
                                                            >
                                                                <Trash2 className="h-5 w-5" />
                                                            </button>
                                                        </>
                                                    )}
                                                </div>
                                            </td>
                                        </tr>
                                    );
                                })
                            )}
                        </tbody>
                    </table>
                </div>

                {/* Pagination Placeholder */}
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
