import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { inviteApi } from '../../services/api/invite';
import type { InviteRoomPreview, InviteLinkInfo } from '../../services/api/invite';
import { useAuthStore } from '../../stores/authStore';
import { useChatStore } from '../../stores/chatStore';

export default function Invite() {
    const { code } = useParams<{ code: string }>();
    const navigate = useNavigate();
    const { user } = useAuthStore();
    const { fetchRooms } = useChatStore();

    const [status, setStatus] = useState<'loading' | 'preview' | 'joining' | 'success' | 'error'>('loading');
    const [room, setRoom] = useState<InviteRoomPreview | null>(null);
    const [linkInfo, setLinkInfo] = useState<InviteLinkInfo | null>(null);
    const [errorMsg, setErrorMsg] = useState('');

    useEffect(() => {
        const fetchInfo = async () => {
            if (!user) {
                // Save current URL so auth can redirect back
                sessionStorage.setItem('invite_redirect', `/invite/${code}`);
                navigate('/login');
                return;
            }

            if (!code) {
                setStatus('error');
                setErrorMsg('邀请码无效');
                return;
            }

            try {
                const res = await inviteApi.getInvite(code);
                setRoom(res.room);
                setLinkInfo(res.link);
                setStatus('preview');
            } catch (err: any) {
                setStatus('error');
                setErrorMsg(err.message || '邀请链接无效或已过期');
            }
        };

        fetchInfo();
    }, [code, user, navigate]);

    const handleJoin = async () => {
        if (!code) return;
        setStatus('joining');
        try {
            await inviteApi.useInvite(code);
            await fetchRooms();
            setStatus('success');
            setTimeout(() => {
                navigate('/chat');
            }, 1500);
        } catch (err: any) {
            const msg = err.message || '加入频道失败';
            // 已是成员也算成功
            if (err.response?.status === 409) {
                await fetchRooms();
                setStatus('success');
                setTimeout(() => navigate('/chat'), 1500);
            } else {
                setStatus('error');
                setErrorMsg(msg);
            }
        }
    };

    return (
        <div className="flex w-full h-screen items-center justify-center bg-[var(--bg-main)]">
            <div className="bg-[var(--bg-secondary)] p-8 rounded-2xl shadow-2xl max-w-md w-full text-center border border-white/5">

                {/* Loading */}
                {status === 'loading' && (
                    <div className="flex flex-col items-center gap-4">
                        <div className="w-16 h-16 rounded-2xl bg-[var(--accent)]/10 flex items-center justify-center animate-pulse">
                            <span className="text-3xl">💬</span>
                        </div>
                        <p className="text-[var(--text-muted)] text-sm">正在获取邀请信息...</p>
                    </div>
                )}

                {/* Preview */}
                {status === 'preview' && room && (
                    <div className="flex flex-col items-center gap-5 animate-in zoom-in duration-300">
                        <div className="w-20 h-20 rounded-2xl bg-gradient-to-br from-[var(--accent)] to-purple-600 flex items-center justify-center text-4xl shadow-lg">
                            #
                        </div>
                        <div>
                            <p className="text-xs font-bold text-[var(--text-muted)] uppercase tracking-wider mb-1">您受邀加入</p>
                            <h2 className="text-2xl font-bold text-[var(--text-main)]">{room.name}</h2>
                            {room.description && (
                                <p className="text-sm text-[var(--text-muted)] mt-1 max-w-xs mx-auto">{room.description}</p>
                            )}
                        </div>

                        {/* Link meta */}
                        {linkInfo && (
                            <div className="flex gap-4 text-xs text-[var(--text-muted)]">
                                {linkInfo.max_uses > 0 && (
                                    <span className="bg-black/20 px-2.5 py-1 rounded-full">
                                        {linkInfo.uses}/{linkInfo.max_uses} 次使用
                                    </span>
                                )}
                                {linkInfo.expires_at && (
                                    <span className="bg-black/20 px-2.5 py-1 rounded-full">
                                        到期：{new Date(linkInfo.expires_at).toLocaleDateString('zh-CN')}
                                    </span>
                                )}
                            </div>
                        )}

                        <div className="flex gap-3 mt-2 w-full">
                            <button
                                onClick={() => navigate('/chat')}
                                className="flex-1 py-2.5 rounded-lg text-sm font-semibold text-[var(--text-muted)] border border-white/10 hover:bg-white/5 transition-colors"
                            >
                                取消
                            </button>
                            <button
                                onClick={handleJoin}
                                className="flex-1 py-2.5 rounded-lg text-sm font-semibold text-white bg-[var(--accent)] hover:bg-[#5b4eb3] transition-colors shadow-md"
                            >
                                加入频道
                            </button>
                        </div>
                    </div>
                )}

                {/* Joining */}
                {status === 'joining' && (
                    <div className="flex flex-col items-center gap-4 animate-in zoom-in duration-200">
                        <div className="w-16 h-16 rounded-2xl bg-[var(--accent)]/10 flex items-center justify-center">
                            <svg className="w-8 h-8 text-[var(--accent)] animate-spin" fill="none" viewBox="0 0 24 24">
                                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                            </svg>
                        </div>
                        <p className="text-[var(--text-muted)] text-sm">正在加入频道...</p>
                    </div>
                )}

                {/* Success */}
                {status === 'success' && (
                    <div className="flex flex-col items-center gap-4 animate-in zoom-in duration-300">
                        <div className="w-16 h-16 rounded-2xl bg-green-500/10 text-green-500 flex items-center justify-center">
                            <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M5 13l4 4L19 7" />
                            </svg>
                        </div>
                        <div>
                            <h2 className="text-xl font-bold text-[var(--text-main)]">加入成功！</h2>
                            <p className="text-sm text-[var(--text-muted)] mt-1">正在跳转到聊天界面...</p>
                        </div>
                    </div>
                )}

                {/* Error */}
                {status === 'error' && (
                    <div className="flex flex-col items-center gap-4 animate-in zoom-in duration-300">
                        <div className="w-16 h-16 rounded-2xl bg-red-500/10 text-red-400 flex items-center justify-center">
                            <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                            </svg>
                        </div>
                        <div>
                            <h2 className="text-xl font-bold text-[var(--text-main)]">邀请无效</h2>
                            <p className="text-sm text-[var(--text-muted)] mt-1">{errorMsg}</p>
                        </div>
                        <button
                            onClick={() => navigate('/chat')}
                            className="mt-2 bg-[var(--accent)] hover:bg-[#5b4eb3] text-white px-6 py-2 rounded-lg text-sm font-semibold transition-colors"
                        >
                            返回主页
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}
