import { useEffect, useState } from 'react';
import { Save, Loader2, MailCheck } from 'lucide-react';
import { adminApi, type EmailSettings } from '../../../services/api/admin';

const PASSWORD_MASK = '********';

const text = {
    saveSuccess: '\u90ae\u7bb1\u914d\u7f6e\u4fdd\u5b58\u6210\u529f\u3002',
    saveFailed: '\u90ae\u7bb1\u914d\u7f6e\u4fdd\u5b58\u5931\u8d25\uff1a',
    title: '\u90ae\u7bb1\u914d\u7f6e',
    description: '\u914d\u7f6e\u7cfb\u7edf\u90ae\u4ef6\u7684 SMTP \u53c2\u6570\uff08\u6ce8\u518c\u9a8c\u8bc1\u3001\u5bc6\u7801\u91cd\u7f6e\u7b49\uff09\u3002',
    autoStartHint: '\u5f53\u201c\u6ce8\u518c\u5fc5\u987b\u90ae\u7bb1\u9a8c\u8bc1\u201d\u5f00\u542f\u540e\uff0c\u7cfb\u7edf\u4f1a\u81ea\u52a8\u542f\u7528\u90ae\u4ef6\u53d1\u9001\u3002',
    smtpHost: 'SMTP \u4e3b\u673a',
    smtpPort: 'SMTP \u7aef\u53e3',
    username: '\u7528\u6237\u540d',
    password: '\u5bc6\u7801',
    fromName: '\u53d1\u4ef6\u4eba\u540d\u79f0',
    fromAddress: '\u53d1\u4ef6\u90ae\u7bb1',
    keepPassword: '\u7559\u7a7a\u5219\u4fdd\u6301\u5f53\u524d\u5bc6\u7801',
    effectiveHint: '\u4fee\u6539\u4f1a\u7acb\u5373\u751f\u6548\u3002',
    saveButton: '\u4fdd\u5b58\u914d\u7f6e',
};

export default function EmailSettingsView() {
    const [config, setConfig] = useState<EmailSettings>({
        enabled: false,
        smtp_host: '',
        smtp_port: 465,
        username: '',
        password: '',
        from_address: '',
        from_name: '',
    });

    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        const fetchConfig = async () => {
            try {
                const data = await adminApi.getEmailSettings();
                setConfig(data);
            } catch (err) {
                console.error('Failed to fetch email config', err);
            } finally {
                setLoading(false);
            }
        };
        fetchConfig();
    }, []);

    const handleSave = async (e: React.FormEvent) => {
        e.preventDefault();
        setSaving(true);
        try {
            await adminApi.updateEmailSettings(config);
            alert(text.saveSuccess);
        } catch (err: any) {
            console.error('Failed to save email settings', err);
            alert(text.saveFailed + (err.response?.data?.error || err.message));
        } finally {
            setSaving(false);
        }
    };

    if (loading) {
        return (
            <div className="flex justify-center items-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-[var(--accent)]" />
            </div>
        );
    }

    return (
        <div className="max-w-2xl space-y-6 animate-in fade-in duration-500">
            <div>
                <h2 className="text-2xl font-bold text-[var(--text-main)]">{text.title}</h2>
                <p className="mt-1 text-sm text-[var(--text-secondary)]">{text.description}</p>
            </div>

            <div className="rounded-xl border border-[var(--border-color)] bg-[var(--bg-main)] px-5 py-4 text-sm text-[var(--text-secondary)]">
                {text.autoStartHint}
            </div>

            <form onSubmit={handleSave} className="bg-[var(--bg-main)] border border-[var(--border-color)] rounded-xl shadow-sm">
                <div className="p-6 space-y-6">
                    <div className="grid grid-cols-2 gap-6">
                        <div className="col-span-2 sm:col-span-1">
                            <label className="block text-sm font-medium text-[var(--text-main)] mb-1">{text.smtpHost}</label>
                            <input
                                type="text"
                                required
                                value={config.smtp_host}
                                onChange={(e) => setConfig({ ...config, smtp_host: e.target.value })}
                                className="w-full bg-[var(--bg-secondary)] border border-[var(--border-color)] text-[var(--text-main)] text-sm rounded-lg focus:ring-[var(--accent)] focus:border-[var(--accent)] block p-2.5 transition-colors"
                                placeholder="smtp.example.com"
                            />
                        </div>

                        <div className="col-span-2 sm:col-span-1">
                            <label className="block text-sm font-medium text-[var(--text-main)] mb-1">{text.smtpPort}</label>
                            <input
                                type="number"
                                required
                                value={config.smtp_port}
                                onChange={(e) => setConfig({ ...config, smtp_port: parseInt(e.target.value, 10) || 465 })}
                                className="w-full bg-[var(--bg-secondary)] border border-[var(--border-color)] text-[var(--text-main)] text-sm rounded-lg focus:ring-[var(--accent)] focus:border-[var(--accent)] block p-2.5 transition-colors"
                            />
                        </div>

                        <div className="col-span-2 sm:col-span-1">
                            <label className="block text-sm font-medium text-[var(--text-main)] mb-1">{text.username}</label>
                            <input
                                type="text"
                                required
                                value={config.username}
                                onChange={(e) => setConfig({ ...config, username: e.target.value })}
                                className="w-full bg-[var(--bg-secondary)] border border-[var(--border-color)] text-[var(--text-main)] text-sm rounded-lg focus:ring-[var(--accent)] focus:border-[var(--accent)] block p-2.5 transition-colors"
                            />
                        </div>

                        <div className="col-span-2 sm:col-span-1">
                            <label className="block text-sm font-medium text-[var(--text-main)] mb-1">{text.password}</label>
                            <input
                                type="password"
                                value={config.password === PASSWORD_MASK ? '' : config.password}
                                onChange={(e) => setConfig({ ...config, password: e.target.value })}
                                placeholder={config.password === PASSWORD_MASK ? text.keepPassword : ''}
                                className="w-full bg-[var(--bg-secondary)] border border-[var(--border-color)] text-[var(--text-main)] text-sm rounded-lg focus:ring-[var(--accent)] focus:border-[var(--accent)] block p-2.5 transition-colors"
                            />
                        </div>

                        <div className="col-span-2 sm:col-span-1">
                            <label className="block text-sm font-medium text-[var(--text-main)] mb-1">{text.fromName}</label>
                            <input
                                type="text"
                                required
                                value={config.from_name}
                                onChange={(e) => setConfig({ ...config, from_name: e.target.value })}
                                className="w-full bg-[var(--bg-secondary)] border border-[var(--border-color)] text-[var(--text-main)] text-sm rounded-lg focus:ring-[var(--accent)] focus:border-[var(--accent)] block p-2.5 transition-colors"
                                placeholder="TeamSphere Bot"
                            />
                        </div>

                        <div className="col-span-2 sm:col-span-1">
                            <label className="block text-sm font-medium text-[var(--text-main)] mb-1">{text.fromAddress}</label>
                            <input
                                type="email"
                                required
                                value={config.from_address}
                                onChange={(e) => setConfig({ ...config, from_address: e.target.value })}
                                className="w-full bg-[var(--bg-secondary)] border border-[var(--border-color)] text-[var(--text-main)] text-sm rounded-lg focus:ring-[var(--accent)] focus:border-[var(--accent)] block p-2.5 transition-colors"
                                placeholder="noreply@example.com"
                            />
                        </div>
                    </div>
                </div>

                <div className="px-6 py-4 bg-[var(--bg-secondary)] border-t border-[var(--border-color)] rounded-b-xl flex items-center justify-between">
                    <div className="flex items-center text-sm text-[var(--text-secondary)]">
                        <MailCheck className="w-4 h-4 mr-2" />
                        {text.effectiveHint}
                    </div>
                    <button
                        type="submit"
                        disabled={saving}
                        className="flex items-center text-white bg-[var(--accent)] hover:bg-opacity-90 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-6 py-2.5 disabled:opacity-50 transition-colors"
                    >
                        {saving ? <Loader2 className="w-5 h-5 mr-2 animate-spin" /> : <Save className="w-5 h-5 mr-2" />}
                        {text.saveButton}
                    </button>
                </div>
            </form>
        </div>
    );
}
