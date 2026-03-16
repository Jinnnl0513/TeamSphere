import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { ApiError } from '../api/client';
import { useAuthStore } from '../stores/authStore';
import { AxiosError } from 'axios';
import { Eye, EyeOff } from 'lucide-react';
import { AnimatedCharacters } from '../components/AnimatedCharacters';
import { StarIcon } from '../components/StarIcon';
import { normalizeText, validateEmail, validatePassword, validateUsername } from '../utils/validators';
import { authApi } from '../services/api/auth';

export default function Register() {
    const navigate = useNavigate();
    const { loadUser, setTokens } = useAuthStore();
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);

    // 动画所需的交互状态
    const [isTyping, setIsTyping] = useState(false);

    const [emailConfig, setEmailConfig] = useState({
        required: false,
        enabled: false,
        allowRegister: true,
    });
    const [email, setEmail] = useState('');
    const [code, setCode] = useState('');
    const [countdown, setCountdown] = useState(0);
    const [sendingCode, setSendingCode] = useState(false);

    useEffect(() => {
        authApi.getEmailRequired()
            .then((res: any) => {
                const allowRegister = res.allow_register !== false;
                setEmailConfig({
                    required: res.email_required,
                    enabled: res.email_enabled,
                    allowRegister,
                });
                if (!allowRegister) {
                    navigate('/login', { replace: true });
                }
            })
            .catch((err) => {
                console.error('Failed to fetch email config', err);
            });
    }, [navigate]);

    useEffect(() => {
        if (countdown <= 0) return;
        const timer = setInterval(() => setCountdown((current) => current - 1), 1000);
        return () => clearInterval(timer);
    }, [countdown]);

    const handleSendCode = async () => {
        const normalizedEmail = normalizeText(email);
        const emailError = validateEmail(normalizedEmail);
        if (emailError) {
            setError(emailError);
            return;
        }
        setError('');
        setSendingCode(true);
        try {
            await authApi.sendCode({ email: normalizedEmail });
            setCountdown(60);
        } catch (err) {
            if (err instanceof ApiError) {
                setError(err.message || '发送验证码失败。');
            } else if (err instanceof AxiosError && err.response) {
                setError(err.response.data.message || '发送验证码失败。');
            } else {
                setError('发生未知错误。');
            }
        } finally {
            setSendingCode(false);
        }
    };

    const handleRegister = async (e: React.FormEvent) => {
        e.preventDefault();
        const normalizedUsername = normalizeText(username);
        if (!emailConfig.allowRegister) {
            setError('当前已关闭注册。');
            return;
        }
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

        if (password !== confirmPassword) {
            setError('两次输入的密码不一致。');
            return;
        }
        setError('');
        setLoading(true);

        try {
            let verificationToken = '';

            if (emailConfig.required) {
                const normalizedEmail = normalizeText(email);
                const emailError = validateEmail(normalizedEmail);
                if (emailError || !code) {
                    setError(emailError || '请填写邮箱和验证码。');
                    setLoading(false);
                    return;
                }
                const verifyRes = await authApi.verifyEmail({ email: normalizedEmail, code });
                verificationToken = verifyRes.verification_token;
            }

            await authApi.register({
                username: normalizedUsername,
                password,
                email: emailConfig.required ? normalizeText(email) : undefined,
                verification_token: verificationToken || undefined,
            });

            const res = await authApi.login({
                username: normalizedUsername,
                password,
            });

            setTokens(res.token, res.refresh_token);
            await loadUser();

            const redirectUrl = sessionStorage.getItem('invite_redirect');
            if (redirectUrl) {
                sessionStorage.removeItem('invite_redirect');
                navigate(redirectUrl);
            } else {
                navigate('/chat');
            }
        } catch (err) {
            if (err instanceof ApiError) {
                setError(err.message || '注册或验证失败。');
            } else if (err instanceof AxiosError && err.response) {
                setError(err.response.data.message || '注册或验证失败。');
            } else {
                setError('发生未知错误。');
            }
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex w-full min-h-screen font-sans bg-white dark:bg-[#1e1f22]">
            <div className="hidden lg:flex flex-1 bg-[#f5f3f0] dark:bg-[#141517] items-center justify-center relative overflow-hidden transition-colors duration-300">
                <AnimatedCharacters
                    isTyping={isTyping}
                    showPassword={showPassword || showConfirmPassword}
                    passwordLength={Math.max(password.length, confirmPassword.length)}
                />
            </div>

            <div className="flex-1 flex flex-col justify-center items-center px-8 sm:px-16 md:px-24 py-12 transition-colors duration-300 overflow-y-auto">
                <div className="w-full max-w-[380px]">
                    <div className="flex justify-center mb-6 text-black dark:text-white">
                        <StarIcon />
                    </div>

                    <h1 className="text-3xl font-bold text-center text-gray-900 dark:text-white mb-2">创建账号</h1>
                    <p className="text-sm text-center text-gray-500 dark:text-gray-400 mb-8">
                        立即加入社区
                    </p>

                    {error && (
                        <div className="mb-6 rounded-lg bg-red-50 dark:bg-red-500/10 p-3 text-sm text-red-600 dark:text-red-400 border border-red-200 dark:border-red-500/30 animate-in fade-in slide-in-from-top-2 text-center">
                            {error}
                        </div>
                    )}

                    <form onSubmit={handleRegister} className="space-y-6">
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

                        {emailConfig.required && (
                            <>
                                <div className="space-y-2">
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                        邮箱
                                    </label>
                                    <div className="flex gap-2 relative">
                                        <input
                                            required
                                            type="email"
                                            value={email}
                                            onChange={(e) => setEmail(e.target.value)}
                                            onFocus={() => setIsTyping(true)}
                                            onBlur={() => setIsTyping(false)}
                                            disabled={loading || sendingCode}
                                            className="w-full border-b border-gray-300 dark:border-gray-600 bg-transparent py-2 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:border-black dark:focus:border-white transition-colors"
                                        />
                                        <button
                                            type="button"
                                            onClick={handleSendCode}
                                            disabled={countdown > 0 || loading || sendingCode || !email}
                                            className="whitespace-nowrap rounded text-xs px-3 py-1 mt-1 bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50 transition-colors"
                                        >
                                            {sendingCode ? '发送中...' : countdown > 0 ? `${countdown}s` : '获取验证码'}
                                        </button>
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                        验证码
                                    </label>
                                    <input
                                        required
                                        value={code}
                                        onChange={(e) => setCode(e.target.value)}
                                        onFocus={() => setIsTyping(true)}
                                        onBlur={() => setIsTyping(false)}
                                        disabled={loading}
                                        maxLength={6}
                                        className="w-full border-b border-gray-300 dark:border-gray-600 bg-transparent py-2 tracking-widest text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:border-black dark:focus:border-white transition-colors"
                                    />
                                </div>
                            </>
                        )}

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

                        <div className="space-y-2 relative">
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                确认密码
                            </label>
                            <div className="relative">
                                <input
                                    required
                                    type={showConfirmPassword ? 'text' : 'password'}
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
                                    disabled={loading}
                                    className="w-full border-b border-gray-300 dark:border-gray-600 bg-transparent py-2 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:border-black dark:focus:border-white transition-colors pr-10"
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                    className="absolute right-0 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 p-1 transition-colors"
                                >
                                    {showConfirmPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                                </button>
                            </div>
                        </div>

                        <div className="pt-4">
                            <button
                                type="submit"
                                disabled={loading}
                                className="w-full bg-[#1a1c23] dark:bg-white hover:bg-black dark:hover:bg-gray-200 text-white dark:text-black rounded-full py-3 font-medium transition-colors shadow-sm disabled:opacity-70 disabled:cursor-not-allowed"
                            >
                                {loading ? '创建中...' : '继续'}
                            </button>
                        </div>
                    </form>

                    <div className="mt-8 text-center text-sm text-gray-500 dark:text-gray-400">
                        已有账号？{' '}
                        <Link to="/login" className="text-black dark:text-white font-medium hover:underline transition-colors">
                            直接登录
                        </Link>
                    </div>
                </div>
            </div>
        </div>
    );
}
