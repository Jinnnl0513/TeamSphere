import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate, Link, useSearchParams } from 'react-router-dom';
import { ApiError } from '../api/client';
import { useAuthStore } from '../stores/authStore';
import { Eye, EyeOff } from 'lucide-react';
import { AnimatedCharacters } from '../components/AnimatedCharacters';
import { StarIcon } from '../components/StarIcon';
import { normalizeText, validatePassword, validateUsername } from '../utils/validators';
import { authApi, type OAuthProvider } from '../services/api/auth';
import { API_BASE_URL } from '../config/app';

export default function Login() {
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();
    const { loadUser, setTokens } = useAuthStore();
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    const [allowRegister, setAllowRegister] = useState(true);
    const [requireTotp, setRequireTotp] = useState(false);
    const [totpCode, setTotpCode] = useState('');
    const [recoveryCode, setRecoveryCode] = useState('');
    const [oauthProviders, setOauthProviders] = useState<OAuthProvider[]>([]);
    const [oauthLoading, setOauthLoading] = useState(false);

    // 动画所需的交互状态
    const [isTyping, setIsTyping] = useState(false);

    useEffect(() => {
        const oauthError = searchParams.get('oauth_error');
        if (oauthError) {
            setError(decodeURIComponent(oauthError));
        }
    }, [searchParams]);

    useEffect(() => {
        authApi.getEmailRequired()
            .then((res: any) => {
                setAllowRegister(res.allow_register !== false);
            })
            .catch((err) => {
                console.error('Failed to fetch register config', err);
            });
    }, []);

    useEffect(() => {
        let mounted = true;
        setOauthLoading(true);
        authApi.getOAuthProviders()
            .then((res) => {
                if (!mounted) return;
                setOauthProviders(res.providers || []);
            })
            .catch((err) => {
                console.error('Failed to fetch oauth providers', err);
            })
            .finally(() => {
                if (!mounted) return;
                setOauthLoading(false);
            });
        return () => { mounted = false; };
    }, []);

    const oauthRedirect = useMemo(() => {
        return sessionStorage.getItem('invite_redirect') || '/chat';
    }, []);

    const handleOAuthLogin = (provider: string) => {
        const url = `${API_BASE_URL}/auth/oauth/${provider}/start?redirect=${encodeURIComponent(oauthRedirect)}`;
        window.location.href = url;
    };

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        const normalizedUsername = normalizeText(username);
        const usernameError = validateUsername(normalizedUsername);
        if (usernameError) {
            setError(usernameError);
            return;
        }
        const passwordError = validatePassword(password);
        if (passwordError) {
            setError(passwordError);
            return;
        }
        setError('');
        setLoading(true);

        try {
            const res = await authApi.login({
                username: normalizedUsername,
                password,
                ...(requireTotp ? { totp_code: totpCode, recovery_code: recoveryCode } : {}),
            });

            setTokens(res.token, res.refresh_token);
            await loadUser();
            setRequireTotp(false);
            setTotpCode('');
            setRecoveryCode('');

            const redirectUrl = sessionStorage.getItem('invite_redirect');
            if (redirectUrl) {
                sessionStorage.removeItem('invite_redirect');
                navigate(redirectUrl);
            } else {
                navigate('/chat');
            }
        } catch (err) {
            if (err instanceof ApiError) {
                if (err.code === 40106 && (err.data as any)?.setup_token) {
                    sessionStorage.setItem('totp_setup_token', (err.data as any).setup_token);
                    const redirectUrl = sessionStorage.getItem('invite_redirect') || '/chat';
                    sessionStorage.setItem('totp_setup_redirect', redirectUrl);
                    navigate('/2fa/setup');
                    return;
                }
                if (err.code === 40103) {
                    setRequireTotp(true);
                    setError('需要二次验证码。');
                } else if (err.code === 40104) {
                    setRequireTotp(true);
                    setError('二次验证码错误。');
                } else {
                    setError(err.message || '登录失败。');
                }
            } else {
                setError('发生未知错误。');
            }
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex w-full h-screen font-sans bg-white dark:bg-[#1e1f22]">
            <div className="hidden lg:flex flex-1 bg-[#f5f3f0] dark:bg-[#141517] items-center justify-center relative overflow-hidden transition-colors duration-300">
                <AnimatedCharacters
                    isTyping={isTyping}
                    showPassword={showPassword}
                    passwordLength={password.length}
                />
            </div>

            <div className="flex-1 flex flex-col justify-center items-center px-8 sm:px-16 md:px-24 transition-colors duration-300">
                <div className="w-full max-w-[360px]">
                    <div className="flex justify-center mb-6 text-black dark:text-white">
                        <StarIcon />
                    </div>

                    <h1 className="text-3xl font-bold text-center text-gray-900 dark:text-white mb-2">欢迎回来</h1>
                    <p className="text-sm text-center text-gray-500 dark:text-gray-400 mb-8">
                        请填写您的登录信息
                    </p>

                    {error && (
                        <div className="mb-6 rounded-lg bg-red-50 dark:bg-red-500/10 p-3 text-sm text-red-600 dark:text-red-400 border border-red-200 dark:border-red-500/30 animate-in fade-in slide-in-from-top-2 text-center">
                            {error}
                        </div>
                    )}

                    <form onSubmit={handleLogin} className="space-y-6">
                        <div className="space-y-2">
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                用户名
                            </label>
                            <input
                                required
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                onFocus={() => setIsTyping(true)}
                                onBlur={() => setIsTyping(false)}
                                disabled={loading}
                                className="w-full border-b border-gray-300 dark:border-gray-600 bg-transparent py-2 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:border-black dark:focus:border-white transition-colors"
                            />
                        </div>

                        <div className="space-y-2 relative">
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                密码
                            </label>
                            <div className="relative">
                                <input
                                    required
                                    type={showPassword ? 'text' : 'password'}
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    disabled={loading}
                                    className="w-full border-b border-gray-300 dark:border-gray-600 bg-transparent py-2 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:border-black dark:focus:border-white transition-colors pr-10"
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowPassword(!showPassword)}
                                    className="absolute right-0 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 p-1 transition-colors"
                                >
                                    {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                                </button>
                            </div>
                        </div>

                        {requireTotp && (
                            <div className="space-y-4">
                                <div className="space-y-2">
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                        二次验证码（TOTP）
                                    </label>
                                    <input
                                        value={totpCode}
                                        onChange={(e) => setTotpCode(e.target.value)}
                                        disabled={loading}
                                        maxLength={6}
                                        className="w-full border-b border-gray-300 dark:border-gray-600 bg-transparent py-2 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:border-black dark:focus:border-white transition-colors"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                        备用恢复码（可选）
                                    </label>
                                    <input
                                        value={recoveryCode}
                                        onChange={(e) => setRecoveryCode(e.target.value)}
                                        disabled={loading}
                                        className="w-full border-b border-gray-300 dark:border-gray-600 bg-transparent py-2 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:border-black dark:focus:border-white transition-colors"
                                    />
                                </div>
                            </div>
                        )}

                        <div className="text-right -mt-2">
                            <Link to="/forgot-password" className="text-xs text-gray-500 dark:text-gray-400 hover:text-black dark:hover:text-white transition-colors">
                                忘记密码？
                            </Link>
                        </div>

                        <div className="pt-4">
                            <button
                                type="submit"
                                disabled={loading}
                                className="w-full bg-[#1a1c23] dark:bg-white hover:bg-black dark:hover:bg-gray-200 text-white dark:text-black rounded-full py-3 font-medium transition-colors shadow-sm disabled:opacity-70 disabled:cursor-not-allowed"
                            >
                                {loading ? '登录中...' : '登录'}
                            </button>
                        </div>
                    </form>

                    {oauthProviders.length > 0 && (
                        <div className="mt-6">
                            <div className="flex items-center gap-3 text-xs text-gray-400">
                                <div className="flex-1 h-px bg-gray-200 dark:bg-gray-700" />
                                <span>其他登录方式</span>
                                <div className="flex-1 h-px bg-gray-200 dark:bg-gray-700" />
                            </div>
                            <div className="mt-4 space-y-2">
                                {oauthProviders.map((provider) => (
                                    <button
                                        key={provider.name}
                                        type="button"
                                        disabled={loading || oauthLoading}
                                        onClick={() => handleOAuthLogin(provider.name)}
                                        className="w-full border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-200 rounded-full py-2.5 font-medium transition-colors hover:border-black dark:hover:border-white hover:text-black dark:hover:text-white disabled:opacity-70 disabled:cursor-not-allowed"
                                    >
                                        使用 {provider.label} 登录
                                    </button>
                                ))}
                            </div>
                        </div>
                    )}

                    <div className="mt-8 text-center text-sm text-gray-500 dark:text-gray-400">
                        {allowRegister ? (
                            <>
                                没有账号？{' '}
                                <Link to="/register" className="text-black dark:text-white font-medium hover:underline transition-colors">
                                    立即注册
                                </Link>
                            </>
                        ) : (
                            <>当前已关闭注册。</>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}
