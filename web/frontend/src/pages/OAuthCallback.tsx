import { useEffect, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import { authApi } from '../services/api/auth';
import FullScreenLoader from '../components/FullScreenLoader';

const sanitizeRedirect = (raw: string | null) => {
    if (!raw) return '/chat';
    if (raw.startsWith('/') && !raw.startsWith('//')) return raw;
    return '/chat';
};

export default function OAuthCallback() {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const { setTokens, loadUser, clearAuth } = useAuthStore();
    const [error, setError] = useState<string | null>(null);
    const [challenge, setChallenge] = useState<string | null>(null);
    const [code, setCode] = useState('');
    const [recoveryCode, setRecoveryCode] = useState('');
    const [verifying, setVerifying] = useState(false);

    useEffect(() => {
        const oauthError = searchParams.get('error');
        if (oauthError) {
            setError(decodeURIComponent(oauthError));
            return;
        }

        const token = searchParams.get('token');
        const refreshToken = searchParams.get('refresh_token');
        const redirect = sanitizeRedirect(searchParams.get('redirect'));
        const challengeParam = searchParams.get('challenge');
        const setupTokenParam = searchParams.get('setup_token');

        if (setupTokenParam) {
            sessionStorage.setItem('totp_setup_token', setupTokenParam);
            sessionStorage.setItem('totp_setup_redirect', redirect);
            navigate('/2fa/setup', { replace: true });
            return;
        }

        if (challengeParam) {
            setChallenge(challengeParam);
            return;
        }

        if (!token || !refreshToken) {
            setError('OAuth 登录失败，请重试。');
            return;
        }

        setTokens(token, refreshToken);
        loadUser()
            .then(() => {
                navigate(redirect, { replace: true });
            })
            .catch(() => {
                clearAuth();
                setError('登录状态初始化失败，请重新登录。');
            });
    }, [clearAuth, loadUser, navigate, searchParams, setTokens]);

    if (challenge) {
        return (
            <div className="min-h-screen flex flex-col items-center justify-center bg-white dark:bg-[#1e1f22] px-6">
                <div className="max-w-sm w-full space-y-4">
                    <h1 className="text-2xl font-semibold text-gray-900 dark:text-white text-center">二次验证</h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400 text-center">请输入 6 位动态验证码或备用恢复码。</p>
                    {error && (
                        <div className="rounded-lg bg-red-50 dark:bg-red-500/10 p-3 text-sm text-red-600 dark:text-red-400 border border-red-200 dark:border-red-500/30">
                            {error}
                        </div>
                    )}
                    <div className="space-y-3">
                        <input
                            value={code}
                            onChange={(e) => setCode(e.target.value)}
                            placeholder="TOTP 验证码"
                            className="w-full border border-gray-300 dark:border-gray-600 bg-transparent py-2 px-3 text-gray-900 dark:text-white rounded-md focus:outline-none focus:border-black dark:focus:border-white transition-colors"
                        />
                        <input
                            value={recoveryCode}
                            onChange={(e) => setRecoveryCode(e.target.value)}
                            placeholder="备用恢复码（可选）"
                            className="w-full border border-gray-300 dark:border-gray-600 bg-transparent py-2 px-3 text-gray-900 dark:text-white rounded-md focus:outline-none focus:border-black dark:focus:border-white transition-colors"
                        />
                    </div>
                    <button
                        onClick={async () => {
                            setError(null);
                            setVerifying(true);
                            try {
                                const res = await authApi.verifyTwoFALogin({
                                    challenge,
                                    code: code || undefined,
                                    recovery_code: recoveryCode || undefined,
                                });
                                const redirect = sanitizeRedirect(searchParams.get('redirect'));
                                setTokens(res.token, res.refresh_token);
                                await loadUser();
                                navigate(redirect, { replace: true });
                            } catch (err) {
                                setError('二次验证失败，请重试。');
                            } finally {
                                setVerifying(false);
                            }
                        }}
                        disabled={verifying}
                        className="w-full rounded-full bg-black text-white dark:bg-white dark:text-black px-6 py-2 text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-60"
                    >
                        {verifying ? '验证中...' : '确认登录'}
                    </button>
                    <div className="text-center">
                        <Link to="/login" className="text-xs text-gray-500 hover:underline">返回登录</Link>
                    </div>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="min-h-screen flex flex-col items-center justify-center bg-white dark:bg-[#1e1f22] px-6">
                <div className="max-w-sm text-center space-y-4">
                    <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">OAuth 登录失败</h1>
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

    return <FullScreenLoader />;
}
