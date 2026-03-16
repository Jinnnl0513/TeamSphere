import { NavLink, useNavigate } from 'react-router-dom';
import {
    LayoutDashboard,
    Users,
    MessageSquare,
    Settings,
    Mail,
    Megaphone,
    LogOut
} from 'lucide-react';

export default function AdminSidebar() {
    const navigate = useNavigate();

    const navItems = [
        { name: '数据总览', path: '/admin/stats', icon: LayoutDashboard },
        { name: '用户管理', path: '/admin/users', icon: Users },
        { name: '房间管理', path: '/admin/rooms', icon: MessageSquare },
        { name: '系统设置', path: '/admin/settings', icon: Settings },
        { name: '邮箱配置', path: '/admin/email', icon: Mail },
        { name: '系统公告', path: '/admin/announcement', icon: Megaphone },
    ];

    return (
        <div className="w-64 h-full bg-[var(--bg-secondary)] border-r border-[var(--border-color)] flex flex-col flex-shrink-0">
            <div className="h-16 flex items-center px-6 border-b border-[var(--border-color)] shrink-0">
                <h1 className="text-xl font-semibold text-[var(--text-main)]">管理后台</h1>
            </div>

            <nav className="flex-1 py-4 overflow-y-auto space-y-1 px-3">
                {navItems.map((item) => {
                    const Icon = item.icon;
                    return (
                        <NavLink
                            key={item.name}
                            to={item.path}
                            className={({ isActive }) =>
                                `flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors ${isActive
                                    ? 'bg-[var(--accent)] text-white'
                                    : 'text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-main)]'
                                }`
                            }
                        >
                            <Icon className="mr-3 h-5 w-5 flex-shrink-0" aria-hidden="true" />
                            {item.name}
                        </NavLink>
                    );
                })}
            </nav>

            <div className="p-4 border-t border-[var(--border-color)] shrink-0">
                <button
                    onClick={() => navigate('/chat')}
                    className="flex w-full items-center px-3 py-2 text-sm font-medium text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-main)] rounded-md transition-colors"
                >
                    <LogOut className="mr-3 h-5 w-5 flex-shrink-0" />
                    返回聊天
                </button>
            </div>
        </div>
    );
}
