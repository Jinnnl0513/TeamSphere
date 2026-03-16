import React, { useState, useEffect, useMemo } from 'react';
import toast from 'react-hot-toast';
import { roomsApi } from '../../../../../services/api/rooms';
import type { InviteLink, RoomMember } from '../../../../../services/api/rooms';
import { useFriendStore } from '../../../../../stores/friendStore';
import type { BaseTabProps } from '../types';
import { InputRow } from '../components/SharedUI';
import { copyText } from '../../../../../utils/clipboard';

export default function InvitesTab({ roomId, canManageMembers }: BaseTabProps) {
    const [inviteLinks, setInviteLinks] = useState<InviteLink[]>([]);
    const [members, setMembers] = useState<RoomMember[]>([]);
    
    // UI state
    const [inviteUserId, setInviteUserId] = useState<number | ''>('');
    const [inviteMsg, setInviteMsg] = useState('');
    const [inviteError, setInviteError] = useState('');
    const [inviteMaxUses, setInviteMaxUses] = useState(0);
    const [inviteExpiresHours, setInviteExpiresHours] = useState(0);

    const [isProcessingInvite, setIsProcessingInvite] = useState(false);
    const [isProcessingLinks, setIsProcessingLinks] = useState(false);
    const [loading, setLoading] = useState(true);
    
    const { friends } = useFriendStore();

    useEffect(() => {
        const load = async () => {
            try {
                const membersRes = await roomsApi.listMembers(roomId);
                if (Array.isArray(membersRes)) setMembers(membersRes);

                if (canManageMembers) {
                    const linksRes = await roomsApi.listInviteLinks(roomId);
                    if (Array.isArray(linksRes)) setInviteLinks(linksRes);
                }
            } catch (err) {
                console.error("Failed to load invites tab details");
            } finally {
                setLoading(false);
            }
        };
        void load();
    }, [roomId, canManageMembers]);

    const inviteableFriends = useMemo(() => {
        if (!friends || friends.length === 0) return [];
        const memberIds = new Set(members.map(m => m.user_id));
        return friends.filter(f => !memberIds.has(f.user.id));
    }, [friends, members]);

    const handleSendInvite = async (e: React.FormEvent) => {
        e.preventDefault();
        setInviteMsg(''); setInviteError('');
        if (!inviteUserId) return;

        setIsProcessingInvite(true);
        try {
            await roomsApi.invite(roomId, Number(inviteUserId));
            setInviteMsg('邀请已发送');
            setInviteUserId('');
        } catch (err: any) {
            setInviteError(err?.message || '发送邀请失败');
        } finally {
            setIsProcessingInvite(false);
        }
    };

    const handleCreateInviteLink = async () => {
        if (!canManageMembers) return;
        setIsProcessingLinks(true);
        try {
            await roomsApi.createInviteLink(roomId, { max_uses: inviteMaxUses, expires_hours: inviteExpiresHours });
            const res = await roomsApi.listInviteLinks(roomId);
            setInviteLinks(Array.isArray(res) ? res : []);
            toast.success('邀请链接已生成');
        } catch (err: any) {
            toast.error(err?.message || '生成链接失败');
        } finally {
            setIsProcessingLinks(false);
        }
    };

    const handleDeleteInviteLink = async (linkId: number) => {
        setIsProcessingLinks(true);
        try {
            await roomsApi.deleteInviteLink(roomId, linkId);
            setInviteLinks(prev => prev.filter(l => l.id !== linkId));
            toast.success('已删除');
        } catch (err: any) {
            toast.error(err?.message || '删除失败');
        } finally {
            setIsProcessingLinks(false);
        }
    };

    if (loading) return <div className="text-sm text-[var(--text-muted)] p-5">加载中...</div>;

    return (
        <div className="space-y-6">
            <section className="bg-[var(--bg-main)] rounded-xl border border-[var(--bg-sidebar)] p-5">
                <div className="text-sm font-bold text-[var(--text-main)] mb-3">邀请好友加入</div>
                <form onSubmit={handleSendInvite} className="space-y-4">
                    {inviteMsg && <div className="p-3 bg-green-500/10 border border-green-500/20 text-green-500 rounded text-sm">{inviteMsg}</div>}
                    {inviteError && <div className="p-3 bg-red-500/10 border border-red-500/20 text-red-500 rounded text-sm">{inviteError}</div>}
                    <div className="flex flex-col md:flex-row gap-3">
                        <select
                            value={inviteUserId} onChange={e => setInviteUserId(e.target.value ? Number(e.target.value) : '')}
                            className="flex-1 bg-[var(--bg-input)] text-[var(--text-main)] p-2 rounded-lg border border-[var(--bg-sidebar)]"
                            disabled={inviteableFriends.length === 0}
                        >
                            <option value="">{inviteableFriends.length === 0 ? '暂无可邀请好友' : '选择一位好友'}</option>
                            {inviteableFriends.map(f => (
                                <option key={f.user.id} value={f.user.id}>{f.user.username}</option>
                            ))}
                        </select>
                        <button type="submit" disabled={!inviteUserId || inviteableFriends.length === 0 || isProcessingInvite} className="px-4 py-2 rounded-md bg-[var(--accent)] text-white text-sm font-semibold disabled:opacity-60 transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)] disabled:active:scale-100">
                            {isProcessingInvite ? '发送中...' : '发送邀请'}
                        </button>
                    </div>
                </form>
            </section>

            <section className="bg-[var(--bg-main)] rounded-xl border border-[var(--bg-sidebar)] p-5">
                <div className="text-sm font-bold text-[var(--text-main)] mb-3">频道邀请链接</div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-3 mb-4">
                    <InputRow label="最大使用次数 (0=不限)" type="number" value={inviteMaxUses} min={0} onChange={setInviteMaxUses} disabled={!canManageMembers} />
                    <InputRow label="有效期小时 (0=永久)" type="number" value={inviteExpiresHours} min={0} onChange={setInviteExpiresHours} disabled={!canManageMembers} />
                    <div className="flex items-end">
                        <button type="button" onClick={handleCreateInviteLink} disabled={!canManageMembers || isProcessingLinks} className="w-full px-4 py-2 rounded-md bg-[var(--accent)] text-white text-sm font-semibold disabled:opacity-60 transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)] disabled:active:scale-100">
                            {isProcessingLinks ? '生成中...' : '生成链接'}
                        </button>
                    </div>
                </div>
                {!canManageMembers ? (
                    <div className="text-sm text-[var(--text-muted)]">无管理权限。</div>
                ) : inviteLinks.length === 0 ? (
                    <div className="text-sm text-[var(--text-muted)]">暂无邀请链接。</div>
                ) : (
                    <div className="space-y-3">
                        {inviteLinks.map(link => (
                            <div key={link.id} className="flex flex-col md:flex-row md:items-center justify-between gap-3 bg-[var(--bg-secondary)] border border-[var(--bg-sidebar)] rounded-lg p-4">
                                <div>
                                    <div className="text-sm font-semibold text-[var(--text-main)]">代码: {link.code}</div>
                                    <div className="text-xs text-[var(--text-muted)]">使用次数: {link.uses}/{link.max_uses === 0 ? '不限' : link.max_uses}</div>
                                    <div className="text-xs text-[var(--text-muted)]">创建时间: {new Date(link.created_at).toLocaleString()}</div>
                                    {link.expires_at && <div className="text-xs text-[var(--text-muted)]">过期: {new Date(link.expires_at).toLocaleString()}</div>}
                                </div>
                                <div className="flex items-center gap-2">
                                    <button
                                        type="button"
                                        onClick={() => {
                                            const url = `${window.location.origin}/invite/${link.code}`;
                                            copyText(url).then(() => toast.success('已复制')).catch(() => toast.error('复制失败'));
                                        }}
                                        className="px-3 py-1.5 rounded bg-[var(--bg-hover)] text-[var(--text-main)] text-xs font-semibold transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[var(--text-muted)] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)]"
                                    >复制链接</button>
                                    <button type="button" onClick={() => handleDeleteInviteLink(link.id)} disabled={isProcessingLinks} className="px-3 py-1.5 rounded border border-[#6b2f2f] text-[#ff6b6b] hover:bg-[#6b2f2f]/30 text-xs font-semibold transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[#ff6b6b] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)] disabled:opacity-60 disabled:active:scale-100">删除</button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </section>
        </div>
    );
}

