import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ApiError } from '../api/client';
import { AnimatedCharacters } from '../components/AnimatedCharacters';
import { StarIcon } from '../components/StarIcon';
import { Eye, EyeOff } from 'lucide-react';
import { normalizeText, validateEmail, validatePassword } from '../utils/validators';
import { authApi } from '../services/api/auth';

export default function ForgotPassword() {
    const navigate = useNavigate();
    const [email, setEmail] = useState('');
    const [code, setCode] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [countdown, setCountdown] = useState(0);
    const [sendingCode, setSendingCode] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    const [isTyping, setIsTyping] = useState(false);
    const [showPassword, setShowPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);

    useEffect(() => {
        if (countdown <= 0) return;
        const timer = window.setInterval(() => {
            setCountdown((prev) => (prev > 0 ? prev - 1 : 0));
        }, 1000);
        return () => window.clearInterval(timer);
    }, [countdown]);

    const handleSendCode = async () => {
        const normalizedEmail = normalizeText(email);
        setError('');
        setSuccess('');
        const emailError = validateEmail(normalizedEmail);
        if (emailError) {
            setError(emailError);
            return;
        }
        if (countdown > 0) return;

        setSendingCode(true);
        try {
            await authApi.sendPasswordResetCode({ email: normalizedEmail });
            setSuccess('如果该邮箱已注册，验证码已发送，请检查收件箱。');
            setCountdown(60);
        } catch (err) {
            if (err instanceof ApiError) {
                setError(err.message || '发送验证码失败。');
            } else {
                setError('发送验证码失败。');
            }
        } finally {
            setSendingCode(false);
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setSuccess('');

        const normalizedEmail = normalizeText(email);
        const normalizedCode = normalizeText(code);
        const emailError = validateEmail(normalizedEmail);
        if (emailError || !normalizedCode || !newPassword || !confirmPassword) {
            setError(emailError || '请完整填写信息。');
            return;
        }
        const passwordError = validatePassword(newPassword, 8, 128);
        if (passwordError) {
            setError(passwordError);
            return;
        }
        if (newPassword !== confirmPassword) {
            setError('两次输入的新密码不一致。');
            return;
        }

        setSubmitting(true);
        try {
            await authApi.resetPassword({
                email: normalizedEmail,
                code: normalizedCode,
                new_password: newPassword,
            });
            setSuccess('密码重置成功，正在返回登录页...');
            window.setTimeout(() => navigate('/login'), 1200);
        } catch (err) {
            if (err instanceof ApiError) {
                setError(err.message || '重置密码失败。');
            } else {
                setError('重置密码失败。');
            }
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <div className="flex w-full h-screen font-sans bg-white dark:bg-[#1e1f22]">
            <div className="hidden lg:flex flex-1 bg-[#f5f3f0] dark:bg-[#141517] items-center justify-center relative overflow-hidden transition-colors duration-300">
                <AnimatedCharacters
                    isTyping={isTyping}
                    showPassword={showPassword || showConfirmPassword}
                    passwordLength={Math.max(newPassword.length, confirmPassword.length)}
                />
            </div>

            <div className="flex-1 flex flex-col justify-center items-center px-8 sm:px-16 md:px-24 transition-colors duration-300">
                <div className="w-full max-w-[380px]">
                    <div className="flex justify-center mb-6 text-black dark:text-white">
                        <StarIcon />
                    </div>

                    <h1 className="text-3xl font-bold text-center text-gray-900 dark:text-white mb-2">忘记密码</h1>
                    <p className="text-sm text-center text-gray-500 dark:text-gray-400 mb-8">
                        通过邮箱验证码重置账号密码
                    </p>

                    {error && (
                        <div className="mb-4 rounded-lg bg-red-50 dark:bg-red-500/10 p-3 text-sm text-red-600 dark:text-red-400 border border-red-200 dark:border-red-500/30 text-center">
                            {error}
                        </div>
                    )}
                    {success && (
                        <div className="mb-4 rounded-lg bg-green-50 dark:bg-green-500/10 p-3 text-sm text-green-700 dark:text-green-300 border border-green-200 dark:border-green-500/30 text-center">
                            {success}
                        </div>
                    )}

                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div className="space-y-2">
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">邮箱</label>
                            <input
                                required
                                type="email"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                onFocus={() => setIsTyping(true)}
                                onBlur={() => setIsTyping(false)}
                                disabled={sendingCode || submitting}
                                className="w-full border-b border-gray-300 dark:border-gray-600 bg-transparent py-2 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:border-black dark:focus:border-white transition-colors"
                            />
                        </div>

                        <div className="space-y-2">
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">验证码</label>
                            <div className="flex items-center gap-2">
                                <input
                                    required
                                    value={code}
                                    onChange={(e) => setCode(e.target.value)}
                                    onFocus={() => setIsTyping(true)}
                                    onBlur={() => setIsTyping(false)}
                                    disabled={submitting}
                                    className="flex-1 border-b border-gray-300 dark:border-gray-600 bg-transparent py-2 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:border-black dark:focus:border-white transition-colors"
                                />
                                <button
                                    type="button"
                                    onClick={() => void handleSendCode()}
                                    disabled={sendingCode || submitting || countdown > 0}
                                    className="px-3 py-2 rounded-full text-sm font-medium bg-[#1a1c23] dark:bg-white text-white dark:text-black hover:opacity-90 disabled:opacity-60 disabled:cursor-not-allowed"
                                >
                                    {sendingCode ? '发送中...' : countdown > 0 ? `${countdown}s` : '获取验证码'}
                                </button>
                            </div>
                        </div>

                        <div className="space-y-2">
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">新密码</label>
                            <div className="relative">
                                <input
                                    required
                                    type={showPassword ? 'text' : 'password'}
                                    value={newPassword}
                                    onChange={(e) => setNewPassword(e.target.value)}
                                    maxLength={128}
                                    disabled={submitting}
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

                        <div className="space-y-2">
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">确认新密码</label>
                            <div className="relative">
                                <input
                                    required
                                    type={showConfirmPassword ? 'text' : 'password'}
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
                                    maxLength={128}
                                    disabled={submitting}
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
                                disabled={submitting}
                                className="w-full bg-[#1a1c23] dark:bg-white hover:bg-black dark:hover:bg-gray-200 text-white dark:text-black rounded-full py-3 font-medium transition-colors shadow-sm disabled:opacity-70 disabled:cursor-not-allowed"
                            >
                                {submitting ? '提交中...' : '重置密码'}
                            </button>
                        </div>
                    </form>

                    <div className="mt-8 text-center text-sm text-gray-500 dark:text-gray-400">
                        <Link to="/login" className="text-black dark:text-white font-medium hover:underline transition-colors">
                            返回登录
                        </Link>
                    </div>
                </div>
            </div>
        </div>
    );
}
