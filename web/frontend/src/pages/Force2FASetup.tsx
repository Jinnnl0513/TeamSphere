import { useEffect, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { authApi } from '../services/api/auth';
import { useAuthStore } from '../stores/authStore';
import FullScreenLoader from '../components/FullScreenLoader';

const sanitizeRedirect = (raw: string | null) => {
    if (!raw) return '/chat';
    if (raw.startsWith('/') && !raw.startsWith('//')) return raw;
    return '/chat';
};

export default function Force2FASetup() {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const { setTokens, loadUser } = useAuthStore();
    const [setupToken, setSetupToken] = useState<string | null>(null);
    const [setup, setSetup] = useState<{ secret: string; otpauth_url: string } | null>(null);
    const [code, setCode] = useState('');
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(true);
    const [enabling, setEnabling] = useState(false);
    const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null);
    const [redirectPath, setRedirectPath] = useState('/chat');

    useEffect(() => {
        const tokenFromQuery = searchParams.get('setup_token');
        const token = tokenFromQuery || sessionStorage.getItem('totp_setup_token');
        const redirect = sanitizeRedirect(searchParams.get('redirect') || sessionStorage.getItem('totp_setup_redirect'));
        setRedirectPath(redirect);
        if (!token) {
            setError('缺少 2FA 初始化凭证，请重新登录。');
            setLoading(false);
            return;
        }
        setSetupToken(token);
        setLoading(true);
        authApi.setupTOTPRequired({ setup_token: token })
            .then((res) => {
                setSetup(res);
            })
            .catch(() => {
                setError('初始化 2FA 失败，请重新登录。');
            })
            .finally(() => {
                setLoading(false);
            });
    }, [searchParams]);

    const handleEnable = async () => {
        if (!setupToken) return;
        if (!code) {
            setError('请输入 6 位验证码。');
            return;
        }
        setError(null);
        setEnabling(true);
        try {
            const res: any = await authApi.enableTOTPRequired({ setup_token: setupToken, code });
            if (res.token && res.refresh_token) {
                setTokens(res.token, res.refresh_token);
                await loadUser();
            }
            setRecoveryCodes(res.recovery_codes || []);
            sessionStorage.removeItem('totp_setup_token');
        } catch {
            setError('启用 2FA 失败，请重试。');
        } finally {
            setEnabling(false);
        }
    };

    if (loading) return <FullScreenLoader />;

    if (error) {
        return (
            <div className="min-h-screen flex flex-col items-center justify-center bg-white dark:bg-[#1e1f22] px-6">
                <div className="max-w-sm text-center space-y-4">
                    <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">需要开启 2FA</h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400">{error}</p>
                    <Link
                        to="/login"
                        className="inline-flex items-center justify-center rounded-full bg-black text-white dark:bg-white dark:text-black px-6 py-2 text-sm font-medium hover:opacity-90 transition-opacity"
                    >
                        返回登录
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen flex flex-col items-center justify-center bg-white dark:bg-[#1e1f22] px-6">
            <div className="max-w-md w-full space-y-4">
                <h1 className="text-2xl font-semibold text-gray-900 dark:text-white text-center">请先启用 2FA</h1>
                <p className="text-sm text-gray-500 dark:text-gray-400 text-center">
                    当前策略要求启用 2FA。请将密钥添加到验证器，并输入 6 位验证码完成启用。
                </p>

                {setup && (
                    <div className="space-y-3">
                        <div className="text-xs text-gray-500 dark:text-gray-400">密钥（手动输入）：</div>
                        <div className="bg-gray-50 dark:bg-[#2b2d31] text-gray-900 dark:text-white rounded-md px-3 py-2 text-sm break-all">
                            {setup.secret}
                        </div>
                        <div className="text-xs text-gray-500 dark:text-gray-400">otpauth URL：</div>
                        <div className="bg-gray-50 dark:bg-[#2b2d31] text-gray-900 dark:text-white rounded-md px-3 py-2 text-xs break-all">
                            {setup.otpauth_url}
                        </div>
                    </div>
                )}

                <div className="space-y-2">
                    <input
                        value={code}
                        onChange={(e) => setCode(e.target.value)}
                        placeholder="6 位验证码"
                        className="w-full border border-gray-300 dark:border-gray-600 bg-transparent py-2 px-3 text-gray-900 dark:text-white rounded-md focus:outline-none focus:border-black dark:focus:border-white transition-colors"
                    />
                    <button
                        onClick={handleEnable}
                        disabled={enabling}
                        className="w-full rounded-full bg-black text-white dark:bg-white dark:text-black px-6 py-2 text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-60"
                    >
                        {enabling ? '启用中...' : '确认启用'}
                    </button>
                </div>

                {recoveryCodes && recoveryCodes.length > 0 && (
                    <div className="space-y-2">
                        <div className="text-xs text-gray-500 dark:text-gray-400">恢复码（请妥善保存）：</div>
                        <div className="bg-gray-50 dark:bg-[#2b2d31] text-gray-900 dark:text-white rounded-md px-3 py-2 text-xs break-all">
                            {recoveryCodes.join(' ')}
                        </div>
                        <button
                            onClick={() => navigate(redirectPath, { replace: true })}
                            className="w-full rounded-full bg-black text-white dark:bg-white dark:text-black px-6 py-2 text-sm font-medium hover:opacity-90 transition-opacity"
                        >
                            继续进入
                        </button>
                    </div>
                )}

                {!recoveryCodes && (
                    <div className="text-center">
                        <Link to="/login" className="text-xs text-gray-500 hover:underline">返回登录</Link>
                    </div>
                )}
            </div>
        </div>
    );
}
