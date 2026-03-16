import { useCallback, useEffect, useState } from 'react';
import { Save, Loader2, Info } from 'lucide-react';
import { adminApi, type SettingsData } from '../../../services/api/admin';

const editableKeys = [
    'registration.allow_register',
    'registration.email_required',
    'oauth.enabled',
    'oauth.frontend_base_url',
    'oauth.github.enabled',
    'oauth.github.client_id',
    'oauth.github.client_secret',
    'oauth.github.redirect_url',
    'oauth.github.scopes',
    'oauth.google.enabled',
    'oauth.google.client_id',
    'oauth.google.client_secret',
    'oauth.google.redirect_url',
    'oauth.google.scopes',
    'oauth.google.hosted_domain',
    'oauth.oidc.enabled',
    'oauth.oidc.client_id',
    'oauth.oidc.client_secret',
    'oauth.oidc.redirect_url',
    'oauth.oidc.scopes',
    'oauth.oidc.issuer_url',
    'security.2fa_policy',
] as const;

const fieldMeta: Record<string, { label: string; help?: string; type: 'text' | 'boolean' | 'textarea' | 'password' }> = {
    'registration.allow_register': {
        label: '允许新用户注册',
        help: '关闭后，普通用户将无法自行注册新账号。',
        type: 'boolean',
    },
    'registration.email_required': {
        label: '注册必须邮箱验证',
        help: '开启后，注册时需要完成邮箱验证码验证。',
        type: 'boolean',
    },
    'oauth.enabled': {
        label: 'Enable OAuth/SSO',
        help: 'Master switch. When disabled, OAuth/SSO entry points are hidden.',
        type: 'boolean',
    },
    'oauth.frontend_base_url': {
        label: 'OAuth Frontend Base URL',
        help: 'Example: https://app.example.com or http://localhost:5173. Leave empty to use current host.',
        type: 'text',
    },
    'oauth.github.enabled': {
        label: 'Enable GitHub OAuth',
        type: 'boolean',
    },
    'oauth.github.client_id': {
        label: 'GitHub Client ID',
        type: 'text',
    },
    'oauth.github.client_secret': {
        label: 'GitHub Client Secret',
        type: 'password',
    },
    'oauth.github.redirect_url': {
        label: 'GitHub Redirect URL',
        help: 'Example: https://api.example.com/api/v1/auth/oauth/github/callback',
        type: 'text',
    },
    'oauth.github.scopes': {
        label: 'GitHub Scopes',
        help: 'Comma separated, e.g. read:user,user:email',
        type: 'text',
    },
    'oauth.google.enabled': {
        label: 'Enable Google OAuth',
        type: 'boolean',
    },
    'oauth.google.client_id': {
        label: 'Google Client ID',
        type: 'text',
    },
    'oauth.google.client_secret': {
        label: 'Google Client Secret',
        type: 'password',
    },
    'oauth.google.redirect_url': {
        label: 'Google Redirect URL',
        help: 'Example: https://api.example.com/api/v1/auth/oauth/google/callback',
        type: 'text',
    },
    'oauth.google.scopes': {
        label: 'Google Scopes',
        help: 'Comma separated, e.g. openid,email,profile',
        type: 'text',
    },
    'oauth.google.hosted_domain': {
        label: 'Google Hosted Domain',
        help: 'Optional. Restrict to a GSuite domain, e.g. example.com',
        type: 'text',
    },
    'oauth.oidc.enabled': {
        label: 'Enable OIDC/SSO',
        type: 'boolean',
    },
    'oauth.oidc.client_id': {
        label: 'OIDC Client ID',
        type: 'text',
    },
    'oauth.oidc.client_secret': {
        label: 'OIDC Client Secret',
        type: 'password',
    },
    'oauth.oidc.redirect_url': {
        label: 'OIDC Redirect URL',
        help: 'Example: https://api.example.com/api/v1/auth/oauth/oidc/callback',
        type: 'text',
    },
    'oauth.oidc.scopes': {
        label: 'OIDC Scopes',
        help: 'Comma separated, e.g. openid,email,profile',
        type: 'text',
    },
    'oauth.oidc.issuer_url': {
        label: 'OIDC Issuer URL',
        help: 'Example: https://sso.example.com',
        type: 'text',
    },
    'security.2fa_policy': {
        label: '2FA Policy',
        help: 'off | optional | admins | required',
        type: 'text',
    },
};

export default function SettingsView() {
    const [settings, setSettings] = useState<SettingsData | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [formValues, setFormValues] = useState<Record<string, string>>({});
    const fetchSettings = useCallback(async () => {
        setLoading(true);
        try {
            const data = await adminApi.getSettings();
            setSettings(data);
            const picked: Record<string, string> = {};
            for (const key of editableKeys) {
                picked[key] = data[key] ?? '';
            }
            setFormValues(picked);
        } catch (err) {
            console.error('Failed to fetch settings', err);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchSettings();
    }, [fetchSettings]);

    const handleSave = async (e: React.FormEvent) => {
        e.preventDefault();
        setSaving(true);
        try {
            const payload: Record<string, string> = {};
            for (const key of editableKeys) {
                payload[key] = formValues[key] ?? '';
            }
            await adminApi.updateSettings(payload);
            alert('设置保存成功！');
            await fetchSettings(); // refresh
        } catch (err) {
            console.error('Failed to save settings', err);
            alert('设置保存失败。');
        } finally {
            setSaving(false);
        }
    };

    const handleChange = (key: string, value: string) => {
        setFormValues(prev => ({ ...prev, [key]: value }));
    };

    if (loading) {
        return (
            <div className="flex justify-center items-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-[var(--accent)]" />
            </div>
        );
    }

    if (!settings) {
        return <div className="text-red-500">加载设置失败。</div>;
    }

    return (
        <div className="max-w-3xl space-y-6 animate-in fade-in duration-500">
            <div>
                <h2 className="text-2xl font-bold text-[var(--text-main)]">系统设置</h2>
                <p className="mt-1 text-sm text-[var(--text-secondary)]">
                    管理注册与全局公告相关设置（仅显示可编辑项目）。
                </p>
            </div>

            <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 flex items-start">
                <Info className="h-5 w-5 text-blue-500 mt-0.5 mr-3 flex-shrink-0" />
                <div className="text-sm text-blue-800 dark:text-blue-300">
                    邮箱相关参数请前往“邮箱配置”菜单修改；此页面不会显示和提交 `email.*` 等受限键。
                </div>
            </div>

            <form onSubmit={handleSave} className="bg-[var(--bg-main)] border border-[var(--border-color)] rounded-xl shadow-sm">
                <div className="p-6 space-y-6">
                    {editableKeys.map((key) => {
                        const meta = fieldMeta[key];
                        const value = formValues[key] ?? '';
                        return (
                            <div key={key}>
                                <label className="block text-sm font-medium text-[var(--text-main)] mb-1">
                                    {meta.label}
                                </label>
                                {meta.help && (
                                    <p className="text-xs text-[var(--text-secondary)] mb-2">{meta.help}</p>
                                )}
                                {meta.type === 'boolean' ? (
                                    <select
                                        className="w-full bg-[var(--bg-secondary)] border border-[var(--border-color)] text-[var(--text-main)] text-sm rounded-lg focus:ring-[var(--accent)] focus:border-[var(--accent)] block p-2.5 transition-colors"
                                        value={value === 'true' ? 'true' : 'false'}
                                        onChange={(e) => handleChange(key, e.target.value)}
                                    >
                                        <option value="true">开启（true）</option>
                                        <option value="false">关闭（false）</option>
                                    </select>
                                ) : meta.type === 'textarea' ? (
                                    <textarea
                                        rows={4}
                                        className="w-full bg-[var(--bg-secondary)] border border-[var(--border-color)] text-[var(--text-main)] text-sm rounded-lg focus:ring-[var(--accent)] focus:border-[var(--accent)] block p-2.5 transition-colors resize-y"
                                        value={value}
                                        onChange={(e) => handleChange(key, e.target.value)}
                                        placeholder="请输入系统公告内容"
                                    />
                                ) : meta.type === 'password' ? (
                                    <input
                                        type="password"
                                        className="w-full bg-[var(--bg-secondary)] border border-[var(--border-color)] text-[var(--text-main)] text-sm rounded-lg focus:ring-[var(--accent)] focus:border-[var(--accent)] block p-2.5 transition-colors"
                                        value={value}
                                        onChange={(e) => handleChange(key, e.target.value)}
                                    />
                                ) : (
                                    <input
                                        type="text"
                                        className="w-full bg-[var(--bg-secondary)] border border-[var(--border-color)] text-[var(--text-main)] text-sm rounded-lg focus:ring-[var(--accent)] focus:border-[var(--accent)] block p-2.5 transition-colors"
                                        value={value}
                                        onChange={(e) => handleChange(key, e.target.value)}
                                    />
                                )}
                            </div>
                        );
                    })}
                </div>

                <div className="px-6 py-4 bg-[var(--bg-secondary)] border-t border-[var(--border-color)] rounded-b-xl flex justify-end">
                    <button
                        type="submit"
                        disabled={saving}
                        className="flex items-center text-white bg-[var(--accent)] hover:bg-opacity-90 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-6 py-2.5 disabled:opacity-50 transition-colors"
                    >
                        {saving ? <Loader2 className="w-5 h-5 mr-2 animate-spin" /> : <Save className="w-5 h-5 mr-2" />}
                        保存设置
                    </button>
                </div>
            </form>
        </div>
    );
}
