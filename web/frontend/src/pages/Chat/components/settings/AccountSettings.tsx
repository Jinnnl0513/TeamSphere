import React, { useEffect, useRef, useState } from 'react';
import { useAuthStore } from '../../../../stores/authStore';
import { ApiError } from '../../../../api/client';
import { usersApi } from '../../../../services/api/users';
import { authApi } from '../../../../services/api/auth';
import { AxiosError } from 'axios';
import { useNavigate } from 'react-router-dom';

export default function AccountSettings() {
    const { user, loadUser, clearAuth } = useAuthStore();
    const navigate = useNavigate();
    const [editMode, setEditMode] = useState<'none' | 'password'>('none');
    const [isSaving, setIsSaving] = useState(false);

    // Password State
    const [oldPassword, setOldPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [pwdError, setPwdError] = useState('');
    const [pwdSuccess, setPwdSuccess] = useState('');

    // TOTP State
    const [totpEnabled, setTotpEnabled] = useState(false);
    const [totpRequired, setTotpRequired] = useState(false);
    const [totpPolicy, setTotpPolicy] = useState<string | undefined>(undefined);
    const [totpLoading, setTotpLoading] = useState(false);
    const [totpSetup, setTotpSetup] = useState<{ secret: string; otpauth_url: string } | null>(null);
    const [totpCode, setTotpCode] = useState('');
    const [totpError, setTotpError] = useState('');
    const [totpSuccess, setTotpSuccess] = useState('');
    const [recoveryRemaining, setRecoveryRemaining] = useState(0);
    const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null);
    const [sessions, setSessions] = useState<Array<any>>([]);
    const [sessionsLoading, setSessionsLoading] = useState(false);

    // Avatar State
    const fileInputRef = useRef<HTMLInputElement>(null);
    const [isUploading, setIsUploading] = useState(false);
    const [isDeletingAccount, setIsDeletingAccount] = useState(false);

    useEffect(() => {
        const load = async () => {
            try {
                const res = await authApi.getTOTPStatus();
                setTotpEnabled(!!res.enabled);
                setTotpRequired(!!res.required);
                setTotpPolicy(res.policy);
                const status = await authApi.recoveryCodesStatus();
                setRecoveryRemaining(status.remaining);
            } catch (err) {
                console.error('Failed to load TOTP status', err);
            }
        };
        load();
        loadSessions();
    }, []);

    const refreshTOTPStatus = async () => {
        const res = await authApi.getTOTPStatus();
        setTotpEnabled(!!res.enabled);
        setTotpRequired(!!res.required);
        setTotpPolicy(res.policy);
    };

    const loadSessions = async () => {
        setSessionsLoading(true);
        try {
            const refreshToken = localStorage.getItem('refresh_token') || undefined;
            const res = await authApi.listSessions(refreshToken);
            setSessions(res.sessions || []);
        } catch (err) {
            console.error('Failed to load sessions', err);
        } finally {
            setSessionsLoading(false);
        }
    };

    const handleStartTOTP = async () => {
        setTotpError('');
        setTotpSuccess('');
        setTotpLoading(true);
        try {
            const res = await authApi.setupTOTP();
            setTotpSetup(res);
        } catch (err) {
            setTotpError(getErrorMessage(err, 'Failed to start TOTP setup'));
        } finally {
            setTotpLoading(false);
        }
    };

    const handleEnableTOTP = async () => {
        if (!totpCode) {
            setTotpError('Please enter the 6-digit code.');
            return;
        }
        setTotpError('');
        setTotpSuccess('');
        setTotpLoading(true);
        try {
            const res = await authApi.enableTOTP({ code: totpCode });
            setTotpSuccess('2FA enabled.');
            setRecoveryCodes(res.recovery_codes || []);
            setTotpSetup(null);
            setTotpCode('');
            await refreshTOTPStatus();
            const status = await authApi.recoveryCodesStatus();
            setRecoveryRemaining(status.remaining);
        } catch (err) {
            setTotpError(getErrorMessage(err, 'Failed to enable 2FA'));
        } finally {
            setTotpLoading(false);
        }
    };

    const handleRegenRecoveryCodes = async () => {
        if (!totpCode) {
            setTotpError('Please enter the 6-digit code.');
            return;
        }
        setTotpError('');
        setTotpSuccess('');
        setTotpLoading(true);
        try {
            const res = await authApi.regenRecoveryCodes({ code: totpCode });
            setRecoveryCodes(res.recovery_codes || []);
            setTotpSuccess('Recovery codes regenerated.');
            const status = await authApi.recoveryCodesStatus();
            setRecoveryRemaining(status.remaining);
        } catch (err) {
            setTotpError(getErrorMessage(err, 'Failed to regenerate recovery codes'));
        } finally {
            setTotpLoading(false);
        }
    };

    const handleRevokeSession = async (id: number) => {
        try {
            await authApi.revokeSession({ session_id: id });
            await loadSessions();
        } catch (err) {
            console.error('Failed to revoke session', err);
        }
    };

    const handleRevokeOtherSessions = async () => {
        const refreshToken = localStorage.getItem('refresh_token');
        if (!refreshToken) return;
        try {
            await authApi.revokeOtherSessions({ refresh_token: refreshToken });
            await loadSessions();
        } catch (err) {
            console.error('Failed to revoke other sessions', err);
        }
    };

    const handleDisableTOTP = async () => {
        if (!totpCode) {
            setTotpError('Please enter the 6-digit code.');
            return;
        }
        setTotpError('');
        setTotpSuccess('');
        setTotpLoading(true);
        try {
            await authApi.disableTOTP({ code: totpCode });
            setTotpSuccess('2FA disabled.');
            setTotpCode('');
            setTotpSetup(null);
            setRecoveryCodes(null);
            setRecoveryRemaining(0);
            await refreshTOTPStatus();
        } catch (err) {
            setTotpError(getErrorMessage(err, 'Failed to disable 2FA'));
        } finally {
            setTotpLoading(false);
        }
    };

    const getErrorMessage = (err: unknown, fallback: string) => {
        if (err instanceof ApiError) return err.message || fallback;
        if (err instanceof AxiosError) {
            return err.response?.data?.message || err.message || fallback;
        }
        if (err instanceof Error) return err.message || fallback;
        return fallback;
    };

    const handleSavePassword = async (e: React.FormEvent) => {
        e.preventDefault();
        setPwdError('');
        setPwdSuccess('');
        if (!oldPassword || !newPassword) return;

        if (newPassword.length < 8 || newPassword.length > 128) {
            setPwdError('新密码长度必须为 8 到 128 位');
            return;
        }

        setIsSaving(true);
        try {
            await usersApi.updatePassword({
                old_password: oldPassword,
                new_password: newPassword
            });
            setPwdSuccess('密码更新成功！');
            setOldPassword('');
            setNewPassword('');
            setTimeout(() => {
                setEditMode('none');
                setPwdSuccess('');
            }, 2000);
        } catch (err) {
            setPwdError(getErrorMessage(err, '更新密码失败'));
        } finally {
            setIsSaving(false);
        }
    };

    const handleAvatarClick = () => {
        fileInputRef.current?.click();
    };

    const handleAvatarChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        setIsUploading(true);
        try {
            await usersApi.uploadAvatar(file);
            await loadUser();
        } catch (err) {
            console.error('Failed to upload avatar', err);
            alert(getErrorMessage(err, '头像上传失败'));
        } finally {
            setIsUploading(false);
            if (fileInputRef.current) fileInputRef.current.value = '';
        }
    };

    const handleDeleteAccount = async () => {
        if (!window.confirm('确定要删除账号吗？此操作不可恢复。')) return;
        setIsDeletingAccount(true);
        try {
            await usersApi.deleteMe();
            clearAuth();
            navigate('/login', { replace: true });
        } catch (err) {
            alert(getErrorMessage(err, '删除账号失败'));
        } finally {
            setIsDeletingAccount(false);
        }
    };

    return (
        <div className="space-y-10 animate-in slide-in-from-right-4 duration-300">
            <div>
                <h2 className="text-xl font-bold text-[var(--text-main)] mb-6">我的账号</h2>
                <div className="bg-[var(--bg-secondary)] rounded-xl overflow-hidden relative">
                    <div className="h-24 bg-[var(--accent)] opacity-80"></div>
                    <div className="px-5 pt-3 pb-6 flex items-start justify-between">
                        {/* Avatar container with hover state */}
                        <div
                            className="relative -mt-14 w-24 h-24 rounded-full border-[6px] border-[var(--bg-secondary)] bg-[var(--bg-sidebar)] flex items-center justify-center text-4xl text-white font-bold shrink-0 shadow-sm cursor-pointer group"
                            onClick={handleAvatarClick}
                        >
                            <img src={user?.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${user?.username}`} alt="avatar" className="w-full h-full rounded-full object-cover" />

                            {/* Hover Overlay */}
                            <div className="absolute inset-0 bg-black/50 rounded-full opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity flex-col">
                                <svg className="w-6 h-6 text-white mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
                                <span className="text-[10px] font-bold text-white uppercase tracking-wider">更换头像</span>
                            </div>

                            {/* Uploading Spinner */}
                            {isUploading && (
                                <div className="absolute inset-0 bg-black/60 rounded-full flex items-center justify-center">
                                    <div className="w-6 h-6 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                                </div>
                            )}

                            <div className="absolute bottom-0 right-0 w-6 h-6 bg-[var(--color-discord-green-500)] border-4 border-[var(--bg-secondary)] rounded-full"></div>
                        </div>

                        <input
                            type="file"
                            accept="image/*"
                            ref={fileInputRef}
                            className="hidden"
                            onChange={handleAvatarChange}
                        />

                        <div className="flex-1 ml-4 mt-2">
                            <h3 className="text-2xl font-bold text-white leading-tight">{user?.username}</h3>
                        </div>
                    </div>

                    <div className="mx-4 mb-4 bg-black/20 rounded-lg p-4 space-y-4">
                        <div className="flex justify-between items-center">
                            <div>
                                <div className="text-xs font-bold text-[var(--text-muted)] uppercase mb-1">用户名</div>
                                <div className="text-sm text-[var(--text-main)]">{user?.username}</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <hr className="border-white/5" />

            <div>
                <h2 className="text-xl font-bold text-white mb-6">密码与验证</h2>

                {editMode === 'password' ? (
                    <form onSubmit={handleSavePassword} className="bg-[var(--bg-secondary)] rounded-xl p-5 space-y-4 max-w-sm border border-white/5">
                        <h3 className="font-semibold text-white">更新您的密码</h3>

                        {pwdError && <div className="text-[#ff6b6b] text-sm font-medium bg-[#ff6b6b]/10 p-2 rounded">{pwdError}</div>}
                        {pwdSuccess && <div className="text-[#20c997] text-sm font-medium bg-[#20c997]/10 p-2 rounded">{pwdSuccess}</div>}

                        <div>
                            <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-1">当前密码</label>
                            <input
                                required
                                type="password"
                                value={oldPassword}
                                onChange={e => setOldPassword(e.target.value)}
                                maxLength={128}
                                className="w-full bg-[var(--bg-main)] text-white rounded-md px-3 py-2 text-sm outline-none focus:ring-1 ring-[var(--accent)]"
                            />
                        </div>
                        <div>
                            <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-1">新密码</label>
                            <input
                                required
                                type="password"
                                value={newPassword}
                                onChange={e => setNewPassword(e.target.value)}
                                maxLength={128}
                                className="w-full bg-[var(--bg-main)] text-white rounded-md px-3 py-2 text-sm outline-none focus:ring-1 ring-[var(--accent)]"
                            />
                        </div>
                        <div className="flex space-x-3 pt-2">
                            <button
                                type="button"
                                onClick={() => { setEditMode('none'); setPwdError(''); }}
                                className="text-[var(--text-muted)] hover:text-white px-3 py-2 text-sm font-medium transition-colors"
                            >
                                取消
                            </button>
                            <button
                                type="submit"
                                disabled={isSaving}
                                className="bg-[var(--accent)] hover:bg-[#5b4eb3] disabled:opacity-50 text-white px-4 py-2 rounded-md text-sm font-semibold transition-colors"
                            >
                                {isSaving ? '处理中...' : '提交更新'}
                            </button>
                        </div>
                    </form>
                ) : (
                    <button
                        onClick={() => setEditMode('password')}
                        className="bg-[var(--accent)] hover:bg-indigo-600 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors shadow-sm"
                    >
                        更改密码
                    </button>
                )}
            </div>

            <div>
                <h2 className="text-xl font-bold text-white mb-6">Two-Factor Authentication</h2>
                <div className="bg-[var(--bg-secondary)] rounded-xl p-5 space-y-4 border border-white/5 max-w-md">
                    <div className="flex items-center justify-between">
                        <div>
                            <div className="text-xs font-bold text-[var(--text-muted)] uppercase mb-1">Status</div>
                            <div className="text-sm text-[var(--text-main)]">{totpEnabled ? 'Enabled' : 'Disabled'}</div>
                        </div>
                        {!totpEnabled && (
                            <button
                                onClick={handleStartTOTP}
                                disabled={totpLoading}
                                className="bg-[var(--accent)] hover:bg-[#5b4eb3] disabled:opacity-50 text-white px-4 py-2 rounded-md text-sm font-semibold transition-colors"
                            >
                                {totpLoading ? 'Preparing...' : 'Enable 2FA'}
                            </button>
                        )}
                    </div>

                    {totpError && <div className="text-[#ff6b6b] text-sm font-medium bg-[#ff6b6b]/10 p-2 rounded">{totpError}</div>}
                    {totpSuccess && <div className="text-[#20c997] text-sm font-medium bg-[#20c997]/10 p-2 rounded">{totpSuccess}</div>}

                    {totpSetup && !totpEnabled && (
                        <div className="space-y-3">
                            <div className="text-xs text-[var(--text-muted)]">Add this secret to your authenticator app (manual entry).</div>
                            <div className="bg-[var(--bg-main)] text-white rounded-md px-3 py-2 text-sm break-all">{totpSetup.secret}</div>
                            <div className="text-xs text-[var(--text-muted)]">Or use the otpauth URL:</div>
                            <div className="bg-[var(--bg-main)] text-white rounded-md px-3 py-2 text-xs break-all">{totpSetup.otpauth_url}</div>
                        </div>
                    )}

                    {(totpSetup || totpEnabled) && (
                        <div>
                            <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-1">6-digit code</label>
                            <input
                                value={totpCode}
                                onChange={(e) => setTotpCode(e.target.value)}
                                maxLength={6}
                                className="w-full bg-[var(--bg-main)] text-white rounded-md px-3 py-2 text-sm outline-none focus:ring-1 ring-[var(--accent)]"
                            />
                        </div>
                    )}

                    {totpSetup && !totpEnabled && (
                        <button
                            onClick={handleEnableTOTP}
                            disabled={totpLoading}
                            className="bg-[var(--accent)] hover:bg-[#5b4eb3] disabled:opacity-50 text-white px-4 py-2 rounded-md text-sm font-semibold transition-colors"
                        >
                            {totpLoading ? 'Enabling...' : 'Confirm & Enable'}
                        </button>
                    )}

                    {totpEnabled && (
                        <div className="space-y-2">
                            <div className="text-xs text-[var(--text-muted)]">Recovery codes remaining: {recoveryRemaining}</div>
                            <button
                                onClick={handleRegenRecoveryCodes}
                                disabled={totpLoading}
                                className="bg-[var(--bg-main)] hover:bg-[var(--bg-main)]/80 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors"
                            >
                                {totpLoading ? 'Generating...' : 'Regenerate Recovery Codes'}
                            </button>
                            {recoveryCodes && recoveryCodes.length > 0 && (
                                <div className="bg-[var(--bg-main)] text-white rounded-md px-3 py-2 text-xs break-all">
                                    {recoveryCodes.join(' ')}
                                </div>
                            )}
                        </div>
                    )}

                    {totpEnabled && (
                        <button
                            onClick={handleDisableTOTP}
                            disabled={totpLoading || totpRequired}
                            className="bg-transparent border border-[var(--color-discord-red-400)] text-[var(--color-discord-red-400)] hover:bg-[var(--color-discord-red-400)] hover:text-white px-4 py-2 rounded-md text-sm font-medium transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
                        >
                            {totpLoading ? 'Disabling...' : 'Disable 2FA'}
                        </button>
                    )}
                    {totpEnabled && totpRequired && (
                        <div className="text-xs text-[var(--text-muted)]">
                            当前策略为 <span className="text-white">{totpPolicy || 'required'}</span>，2FA 已被强制，无法关闭。
                        </div>
                    )}
                </div>
            </div>

            <div>
                <h2 className="text-xl font-bold text-white mb-6">登录设备</h2>
                <div className="bg-[var(--bg-secondary)] rounded-xl p-5 space-y-4 border border-white/5">
                    <div className="flex items-center justify-between">
                        <div className="text-sm text-[var(--text-main)]">当前登录会话</div>
                        <button
                            onClick={handleRevokeOtherSessions}
                            className="text-xs text-[var(--text-muted)] hover:text-white transition-colors"
                        >
                            退出其他设备
                        </button>
                    </div>
                    {sessionsLoading ? (
                        <div className="text-sm text-[var(--text-muted)]">加载中...</div>
                    ) : (
                        <div className="space-y-3">
                            {sessions.map((s) => (
                                <div key={s.id} className="flex items-center justify-between bg-[var(--bg-main)] rounded-lg p-3">
                                    <div>
                                        <div className="text-sm text-white">{s.user_agent || 'Unknown Device'}</div>
                                        <div className="text-xs text-[var(--text-muted)]">IP: {s.ip_address || '-'} · Last: {s.last_used_at || s.created_at}</div>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        {s.is_current && (
                                            <span className="text-xs text-[var(--color-discord-green-500)]">当前</span>
                                        )}
                                        {!s.is_current && !s.revoked_at && (
                                            <button
                                                onClick={() => handleRevokeSession(s.id)}
                                                className="text-xs text-[var(--color-discord-red-400)] hover:text-white"
                                            >
                                                退出
                                            </button>
                                        )}
                                        {s.revoked_at && (
                                            <span className="text-xs text-[var(--text-muted)]">已退出</span>
                                        )}
                                    </div>
                                </div>
                            ))}
                            {sessions.length === 0 && (
                                <div className="text-sm text-[var(--text-muted)]">暂无会话记录</div>
                            )}
                        </div>
                    )}
                </div>
            </div>

            <hr className="border-white/5" />

            <div>
                <h2 className="text-xl font-bold text-white mb-6 text-[var(--color-discord-red-400)]">危险区域</h2>
                <button
                    onClick={handleDeleteAccount}
                    disabled={isDeletingAccount}
                    className="bg-transparent border border-[var(--color-discord-red-400)] text-[var(--color-discord-red-400)] hover:bg-[var(--color-discord-red-400)] hover:text-white px-4 py-2 rounded-md text-sm font-medium transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
                >
                    {isDeletingAccount ? '删除中...' : '删除账号'}
                </button>
            </div>
        </div>
    );
}
