import React, { useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { useNavigate } from 'react-router-dom';
import { setupApi } from '../services/api/setup';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Label } from '../components/ui/Label';
import { validateEmail } from '../utils/validators';

type Feedback = {
    tone: 'error' | 'success';
    text: string;
};

const getErrorMessage = (err: unknown, fallback: string) => {
    if (err instanceof AxiosError) {
        return err.response?.data?.message || err.message || fallback;
    }
    if (err instanceof Error) {
        return err.message || fallback;
    }
    return fallback;
};

export default function Setup() {
    const navigate = useNavigate();
    const [step, setStep] = useState(1);
    const [loading, setLoading] = useState(false);
    const [feedback, setFeedback] = useState<Feedback | null>(null);
    const [dbConfigured, setDbConfigured] = useState(false);

    const [dbHost, setDbHost] = useState('localhost');
    const [dbPort, setDbPort] = useState('5432');
    const [dbUser, setDbUser] = useState('team_sphere');
    const [dbPass, setDbPass] = useState('');
    const [dbName, setDbName] = useState('team_sphere');

    const [redisEnabled, setRedisEnabled] = useState(false);
    const [redisHost, setRedisHost] = useState('');
    const [redisPort, setRedisPort] = useState('6379');
    const [redisPass, setRedisPass] = useState('');
    const [redisDb, setRedisDb] = useState('0');

    const [smtpHost, setSmtpHost] = useState('');
    const [smtpPort, setSmtpPort] = useState('587');
    const [smtpUser, setSmtpUser] = useState('');
    const [smtpPass, setSmtpPass] = useState('');
    const [smtpFrom, setSmtpFrom] = useState('');
    const [testRecipient, setTestRecipient] = useState('');

    const [adminUser, setAdminUser] = useState('');
    const [adminEmail, setAdminEmail] = useState('');
    const [adminPass, setAdminPass] = useState('');

    useEffect(() => {
        setupApi.getStatus()
            .then((res) => {
                if (!res.needed) {
                    navigate('/login');
                    return;
                }

                const alreadyConfigured = !!res.db_configured;
                setDbConfigured(alreadyConfigured);
                if (alreadyConfigured) {
                    setStep(3);
                }
            })
            .catch((err) => {
                if (err instanceof AxiosError && err.response?.status === 404) {
                    navigate('/login');
                    return;
                }
                setFeedback({
                    tone: 'error',
                    text: `\u52a0\u8f7d\u521d\u59cb\u5316\u72b6\u6001\u5931\u8d25\uff1a${getErrorMessage(err, '\u672a\u77e5\u9519\u8bef')}`,
                });
            });
    }, [navigate]);

    const stepTitle = dbConfigured
        ? '\u521b\u5efa\u7ba1\u7406\u5458\u8d26\u53f7'
        : step === 1
            ? '\u6570\u636e\u5e93\u8fde\u63a5'
            : step === 2
                ? '\u90ae\u4ef6\u8bbe\u7f6e'
                : '\u521b\u5efa\u7ba1\u7406\u5458\u8d26\u53f7';

    const clearFeedback = () => setFeedback(null);
    const hasEmailConfig = [smtpHost, smtpUser, smtpPass, smtpFrom].some((value) => value.trim() !== '');

    const validateDBConfig = () => {
        if (!dbHost.trim()) {
            setFeedback({ tone: 'error', text: '\u6570\u636e\u5e93\u5730\u5740\u4e0d\u80fd\u4e3a\u7a7a\u3002' });
            return false;
        }
        if (!/^\d+$/.test(dbPort) || Number(dbPort) <= 0) {
            setFeedback({ tone: 'error', text: '\u6570\u636e\u5e93\u7aef\u53e3\u5fc5\u987b\u662f\u6709\u6548\u7684\u6b63\u6574\u6570\u3002' });
            return false;
        }
        if (!dbUser.trim()) {
            setFeedback({ tone: 'error', text: '\u6570\u636e\u5e93\u7528\u6237\u4e0d\u80fd\u4e3a\u7a7a\u3002' });
            return false;
        }
        if (!dbName.trim()) {
            setFeedback({ tone: 'error', text: '\u6570\u636e\u5e93\u540d\u4e0d\u80fd\u4e3a\u7a7a\u3002' });
            return false;
        }

        return true;
    };

    const validateRedisConfig = () => {
        if (!redisEnabled) {
            return true;
        }
        if (!redisHost.trim()) {
            setFeedback({ tone: 'error', text: '\u0052\u0065\u0064\u0069\u0073 \u5730\u5740\u4e0d\u80fd\u4e3a\u7a7a\u3002' });
            return false;
        }
        if (!/^\d+$/.test(redisPort) || Number(redisPort) <= 0) {
            setFeedback({ tone: 'error', text: '\u0052\u0065\u0064\u0069\u0073 \u7aef\u53e3\u5fc5\u987b\u662f\u6709\u6548\u7684\u6b63\u6574\u6570\u3002' });
            return false;
        }
        if (!/^\d+$/.test(redisDb) || Number(redisDb) < 0) {
            setFeedback({ tone: 'error', text: '\u0052\u0065\u0064\u0069\u0073 \u0044\u0042 \u5fc5\u987b\u662f\u975e\u8d1f\u6574\u6570\u3002' });
            return false;
        }
        return true;
    };

    const validateEmailConfig = (requireRecipient: boolean) => {
        const fields = [
            { value: smtpHost, message: '\u8bf7\u8f93\u5165 SMTP \u670d\u52a1\u5668\u3002' },
            { value: smtpPort, message: '\u8bf7\u8f93\u5165 SMTP \u7aef\u53e3\u3002' },
            { value: smtpUser, message: '\u8bf7\u8f93\u5165 SMTP \u7528\u6237\u540d\u3002' },
            { value: smtpPass, message: '\u8bf7\u8f93\u5165 SMTP \u5bc6\u7801\u3002' },
            { value: smtpFrom, message: '\u8bf7\u8f93\u5165\u53d1\u4ef6\u5730\u5740\u3002' },
        ];

        for (const field of fields) {
            if (!field.value.trim()) {
                setFeedback({ tone: 'error', text: field.message });
                return false;
            }
        }

        if (!/^\d+$/.test(smtpPort) || Number(smtpPort) <= 0) {
            setFeedback({ tone: 'error', text: 'SMTP \u7aef\u53e3\u5fc5\u987b\u662f\u6709\u6548\u7684\u6b63\u6574\u6570\u3002' });
            return false;
        }

        if (requireRecipient && !testRecipient.trim()) {
            setFeedback({ tone: 'error', text: '\u8bf7\u5148\u8f93\u5165\u6d4b\u8bd5\u6536\u4ef6\u90ae\u7bb1\u5730\u5740\u3002' });
            return false;
        }

        return true;
    };

    const handleTestDB = async () => {
        clearFeedback();
        if (!validateDBConfig()) {
            return;
        }
        if (!validateRedisConfig()) {
            return;
        }

        setLoading(true);

        try {
            await setupApi.testConnection({
                db: {
                    host: dbHost.trim(),
                    port: Number(dbPort),
                    user: dbUser.trim(),
                    password: dbPass,
                    dbname: dbName.trim(),
                },
                redis_enabled: redisEnabled,
                ...(redisEnabled
                    ? {
                        redis: {
                            host: redisHost.trim(),
                            port: Number(redisPort),
                            password: redisPass,
                            db: Number(redisDb),
                        },
                    }
                    : {}),
            });
            setFeedback({
                tone: 'success',
                text: redisEnabled ? '\u6570\u636e\u5e93\u4e0e Redis \u8fde\u63a5\u6210\u529f\u3002' : '\u6570\u636e\u5e93\u8fde\u63a5\u6210\u529f\u3002',
            });
        } catch (err) {
            setFeedback({ tone: 'error', text: getErrorMessage(err, '\u6570\u636e\u5e93\u8fde\u63a5\u5931\u8d25\u3002') });
        } finally {
            setLoading(false);
        }
    };

    const handleContinueFromDB = () => {
        clearFeedback();
        if (!validateDBConfig()) {
            return;
        }
        if (!validateRedisConfig()) {
            return;
        }

        setStep(2);
    };

    const handleTestEmail = async () => {
        clearFeedback();
        if (!validateEmailConfig(true)) {
            return;
        }

        setLoading(true);

        try {
            await setupApi.testEmail({
                smtp_host: smtpHost.trim(),
                smtp_port: Number(smtpPort),
                username: smtpUser.trim(),
                password: smtpPass,
                from_address: smtpFrom.trim(),
                from_name: 'TeamSphere',
                to: testRecipient.trim(),
            });
            setFeedback({ tone: 'success', text: '\u6d4b\u8bd5\u90ae\u4ef6\u53d1\u9001\u6210\u529f\u3002' });
        } catch (err) {
            setFeedback({ tone: 'error', text: getErrorMessage(err, '\u6d4b\u8bd5\u90ae\u4ef6\u53d1\u9001\u5931\u8d25\u3002') });
        } finally {
            setLoading(false);
        }
    };

    const handleComplete = async () => {
        if (!adminUser.trim()) {
            setFeedback({ tone: 'error', text: '\u7ba1\u7406\u5458\u7528\u6237\u540d\u4e0d\u80fd\u4e3a\u7a7a\u3002' });
            return;
        }
        const emailError = validateEmail(adminEmail.trim());
        if (emailError) {
            setFeedback({ tone: 'error', text: emailError });
            return;
        }
        if (!/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,128}$/.test(adminPass)) {
            setFeedback({ tone: 'error', text: '\u7ba1\u7406\u5458\u5bc6\u7801\u9700\u5305\u542b\u5927\u5c0f\u5199\u5b57\u6bcd\u548c\u6570\u5b57\uff0c\u957f\u5ea6 8-128 \u4f4d\u3002' });
            return;
        }
        if (!validateRedisConfig()) {
            return;
        }
        if (hasEmailConfig && !validateEmailConfig(false)) {
            return;
        }

        setLoading(true);
        clearFeedback();

        const payload = dbConfigured
            ? {
                admin_username: adminUser.trim(),
                admin_password: adminPass,
                admin_email: adminEmail.trim(),
            }
            : {
                db: {
                    host: dbHost.trim(),
                    port: Number(dbPort),
                    user: dbUser.trim(),
                    password: dbPass,
                    dbname: dbName.trim(),
                },
                redis_enabled: redisEnabled,
                ...(redisEnabled
                    ? {
                        redis: {
                            host: redisHost.trim(),
                            port: Number(redisPort),
                            password: redisPass,
                            db: Number(redisDb),
                        },
                    }
                    : {}),
                ...(hasEmailConfig
                    ? {
                        email_enabled: true,
                        email: {
                            smtp_host: smtpHost.trim(),
                            smtp_port: Number(smtpPort),
                            username: smtpUser.trim(),
                            password: smtpPass,
                            from_address: smtpFrom.trim(),
                            from_name: 'TeamSphere',
                        },
                    }
                    : {
                        email_enabled: false,
                    }),
                admin_username: adminUser.trim(),
                admin_password: adminPass,
                admin_email: adminEmail.trim(),
            };

        try {
            const res = await setupApi.setup(payload);
            localStorage.setItem('token', res.token);
            if (res.refresh_token) {
                localStorage.setItem('refresh_token', res.refresh_token);
            }

            await new Promise((resolve) => setTimeout(resolve, 1500));
            window.location.href = '/chat';
        } catch (err) {
            setFeedback({ tone: 'error', text: getErrorMessage(err, '\u521d\u59cb\u5316\u5931\u8d25\u3002') });
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-[radial-gradient(circle_at_top,_#3d4457_0%,_#17191f_45%,_#101116_100%)] px-4 py-10">
            <div className="absolute inset-0 bg-[linear-gradient(135deg,rgba(255,255,255,0.08),transparent_28%,transparent_72%,rgba(255,255,255,0.05))]" />
            <div className="custom-scrollbar relative w-full max-w-[560px] max-h-[88vh] overflow-y-auto rounded-[28px] border border-white/10 bg-[#171a20]/92 p-8 shadow-[0_32px_80px_rgba(0,0,0,0.45)] backdrop-blur-xl">
                <div className="mb-8 flex items-start justify-between gap-4">
                    <div>
                        <p className="text-xs font-semibold uppercase tracking-[0.28em] text-[#8f99ad]">{'TeamSphere \u521d\u59cb\u5316'}</p>
                        <h1 className="mt-3 text-3xl font-semibold text-white">{stepTitle}</h1>
                        <p className="mt-2 text-sm text-[#aab3c5]">
                            {dbConfigured
                                ? '\u6570\u636e\u5e93\u5df2\u5c31\u7eea\uff0c\u521b\u5efa\u9996\u4e2a\u7ba1\u7406\u5458\u8d26\u53f7\u5373\u53ef\u5b8c\u6210\u521d\u59cb\u5316\u3002'
                                : `\u7b2c ${step} / 3 \u6b65`}
                        </p>
                    </div>
                    <div className="rounded-full border border-white/10 bg-white/5 px-3 py-2 text-xs font-medium uppercase tracking-[0.2em] text-[#c8d0df]">
                        {'\u8fdc\u7a0b\u53ef\u7528'}
                    </div>
                </div>

                {feedback && (
                    <div className={`mb-6 rounded-2xl border px-4 py-3 text-sm ${feedback.tone === 'error'
                        ? 'border-[#f05b6d]/40 bg-[#f05b6d]/10 text-[#ffd5db]'
                        : 'border-[#57c08a]/40 bg-[#57c08a]/10 text-[#dbffea]'}`}>
                        {feedback.text}
                    </div>
                )}

                <div className="mb-6 rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-[#d7ddeb]">
                    {'\u521d\u59cb\u5316\u671f\u95f4\u5141\u8bb8\u8fdc\u7a0b\u8bbf\u95ee\uff0c\u5b8c\u6210\u540e setup \u63a5\u53e3\u4f1a\u81ea\u52a8\u5173\u95ed\u3002'}
                </div>

                {!dbConfigured && step === 1 && (
                    <div className="space-y-5">
                        <div>
                            <Label>{'\u6570\u636e\u5e93\u5730\u5740'}</Label>
                            <Input value={dbHost} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setDbHost(e.target.value)} />
                        </div>
                        <div className="grid gap-4 sm:grid-cols-[1fr_2fr]">
                            <div>
                                <Label>{'\u7aef\u53e3'}</Label>
                                <Input value={dbPort} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setDbPort(e.target.value)} />
                            </div>
                            <div>
                                <Label>{'\u6570\u636e\u5e93\u540d'}</Label>
                                <Input value={dbName} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setDbName(e.target.value)} />
                            </div>
                        </div>
                        <div>
                            <Label>{'\u6570\u636e\u5e93\u7528\u6237'}</Label>
                            <Input value={dbUser} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setDbUser(e.target.value)} />
                        </div>
                        <div>
                            <Label>{'\u6570\u636e\u5e93\u5bc6\u7801'}</Label>
                            <Input type="password" value={dbPass} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setDbPass(e.target.value)} />
                        </div>
                        <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-[#d7ddeb]">
                            {'\u0052\u0065\u0064\u0069\u0073 \u4e3a\u53ef\u9009\u9879\u3002\u5982\u679c\u4e0d\u542f\u7528\uff0c\u7cfb\u7edf\u5c06\u8df3\u8fc7 \u0052\u0065\u0064\u0069\u0073\u3002'}
                        </div>
                        <label htmlFor="redis-enabled" className="flex items-center justify-between rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
                            <span className="text-sm font-medium text-[#d7ddeb]">{'\u542f\u7528 \u0052\u0065\u0064\u0069\u0073'}</span>
                            <span className="relative inline-flex items-center">
                                <input
                                    id="redis-enabled"
                                    type="checkbox"
                                    checked={redisEnabled}
                                    onChange={(e) => setRedisEnabled(e.target.checked)}
                                    className="peer sr-only"
                                />
                                <span className="h-6 w-11 rounded-full border border-white/15 bg-white/10 transition-colors peer-checked:bg-[#6c5dd3]" />
                                <span className="absolute left-1 top-1 h-4 w-4 rounded-full bg-white shadow transition-transform peer-checked:translate-x-5" />
                            </span>
                        </label>
                        {redisEnabled && (
                            <div className="space-y-5">
                                <div>
                                    <Label>{'\u0052\u0065\u0064\u0069\u0073 \u5730\u5740'}</Label>
                                    <Input value={redisHost} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setRedisHost(e.target.value)} />
                                </div>
                                <div className="grid gap-4 sm:grid-cols-[1fr_1fr]">
                                    <div>
                                        <Label>{'\u0052\u0065\u0064\u0069\u0073 \u7aef\u53e3'}</Label>
                                        <Input value={redisPort} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setRedisPort(e.target.value)} />
                                    </div>
                                    <div>
                                        <Label>{'\u0052\u0065\u0064\u0069\u0073 \u0044\u0042'}</Label>
                                        <Input value={redisDb} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setRedisDb(e.target.value)} />
                                    </div>
                                </div>
                                <div>
                                    <Label>{'\u0052\u0065\u0064\u0069\u0073 \u5bc6\u7801'}</Label>
                                    <Input type="password" value={redisPass} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setRedisPass(e.target.value)} />
                                </div>
                            </div>
                        )}
                        <div className="flex flex-col gap-3 sm:flex-row">
                            <Button variant="secondary" className="h-12 text-base sm:flex-1" onClick={handleTestDB} disabled={loading}>
                                {loading ? '\u6b63\u5728\u6d4b\u8bd5\u8fde\u63a5...' : '\u6d4b\u8bd5\u8fde\u63a5'}
                            </Button>
                            <Button className="h-12 text-base sm:flex-1" onClick={handleContinueFromDB} disabled={loading}>
                                {'\u7ee7\u7eed'}
                            </Button>
                        </div>
                    </div>
                )}

                {!dbConfigured && step === 2 && (
                    <div className="space-y-5">
                        <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-[#d7ddeb]">
                            {'\u90ae\u4ef6\u914d\u7f6e\u4e3a\u53ef\u9009\u9879\u3002\u5982\u679c\u586b\u5199\u5e76\u4fdd\u5b58\uff0c\u7cfb\u7edf\u4f1a\u81ea\u52a8\u542f\u7528\u90ae\u4ef6\u53d1\u9001\uff0c\u5e76\u5f00\u542f\u6ce8\u518c\u5fc5\u987b\u90ae\u7bb1\u9a8c\u8bc1\u3002'}
                        </div>

                        <div className="space-y-5">
                            <div className="grid gap-4 sm:grid-cols-[2fr_1fr]">
                                <div>
                                    <Label>{'SMTP \u670d\u52a1\u5668'}</Label>
                                    <Input value={smtpHost} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSmtpHost(e.target.value)} />
                                </div>
                                <div>
                                    <Label>{'SMTP \u7aef\u53e3'}</Label>
                                    <Input value={smtpPort} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSmtpPort(e.target.value)} />
                                </div>
                            </div>
                            <div>
                                <Label>{'SMTP \u7528\u6237\u540d'}</Label>
                                <Input
                                    value={smtpUser}
                                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                                        setSmtpUser(e.target.value);
                                        if (!smtpFrom) {
                                            setSmtpFrom(e.target.value);
                                        }
                                    }}
                                />
                            </div>
                            <div>
                                <Label>{'SMTP \u5bc6\u7801'}</Label>
                                <Input type="password" value={smtpPass} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSmtpPass(e.target.value)} />
                            </div>
                            <div>
                                <Label>{'\u53d1\u4ef6\u5730\u5740'}</Label>
                                <Input value={smtpFrom} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSmtpFrom(e.target.value)} />
                            </div>
                            <div>
                                <Label>{'\u6d4b\u8bd5\u6536\u4ef6\u4eba'}</Label>
                                <Input value={testRecipient} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setTestRecipient(e.target.value)} />
                            </div>
                        </div>

                        <div className="flex flex-col gap-3 sm:flex-row">
                            <Button variant="ghost" className="sm:w-auto" onClick={() => { clearFeedback(); setStep(1); }} disabled={loading}>
                                {'\u4e0a\u4e00\u6b65'}
                            </Button>
                            {hasEmailConfig && (
                                <Button variant="secondary" className="sm:flex-1" onClick={handleTestEmail} disabled={loading}>
                                    {loading ? '\u53d1\u9001\u4e2d...' : '\u53d1\u9001\u6d4b\u8bd5\u90ae\u4ef6'}
                                </Button>
                            )}
                            <Button className="sm:flex-1" onClick={() => { clearFeedback(); setStep(3); }} disabled={loading}>
                                {'\u7ee7\u7eed'}
                            </Button>
                        </div>
                    </div>
                )}

                {step === 3 && (
                    <div className="space-y-5">
                        <div>
                            <Label>{'\u7ba1\u7406\u5458\u7528\u6237\u540d'}</Label>
                            <Input value={adminUser} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setAdminUser(e.target.value)} />
                        </div>
                        <div>
                            <Label>{'管理员邮箱'}</Label>
                            <Input type="email" value={adminEmail} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setAdminEmail(e.target.value)} />
                        </div>
                        <div>
                            <Label>{'\u7ba1\u7406\u5458\u5bc6\u7801'}</Label>
                            <Input type="password" value={adminPass} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setAdminPass(e.target.value)} />
                        </div>

                        <div className="flex flex-col gap-3 sm:flex-row">
                            {!dbConfigured && (
                                <Button variant="ghost" className="sm:w-auto" onClick={() => { clearFeedback(); setStep(2); }} disabled={loading}>
                                    {'\u4e0a\u4e00\u6b65'}
                                </Button>
                            )}
                            <Button className="sm:flex-1" onClick={handleComplete} disabled={loading}>
                                {loading ? '\u6b63\u5728\u5b8c\u6210\u521d\u59cb\u5316...' : '\u5b8c\u6210\u521d\u59cb\u5316'}
                            </Button>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
