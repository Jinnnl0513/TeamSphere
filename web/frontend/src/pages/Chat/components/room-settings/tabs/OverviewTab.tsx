import React, { useState, useEffect, useRef } from 'react';
import toast from 'react-hot-toast';
import { roomsApi } from '../../../../../services/api/rooms';
import type { RoomSettings, RoomMember } from '../../../../../services/api/rooms';
import { filesApi } from '../../../../../services/api/files';
import { useAuthStore } from '../../../../../stores/authStore';
import { useChatStore } from '../../../../../stores/chatStore';
import { useNavigate } from 'react-router-dom';
import type { BaseTabProps } from '../types';

export default function OverviewTab({ roomId, canManageSettings, isOwner, onClose }: BaseTabProps & { onClose: () => void }) {
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [settings, setSettings] = useState<Partial<RoomSettings>>({});
    const [members, setMembers] = useState<RoomMember[]>([]);
    
    const [loading, setLoading] = useState(true);
    const [isSavingRoom, setIsSavingRoom] = useState(false);
    const [isSavingSettings, setIsSavingSettings] = useState(false);
    const [transferOwnerId, setTransferOwnerId] = useState<number | ''>('');
    const [saveRoomSuccess, setSaveRoomSuccess] = useState(false);
    const [saveSettingsSuccess, setSaveSettingsSuccess] = useState(false);
    
    const fileInputRef = useRef<HTMLInputElement>(null);
    const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);

    const currentUser = useAuthStore(s => s.user);
    const { fetchRooms, leaveRoom, rooms } = useChatStore();
    const navigate = useNavigate();

    useEffect(() => {
        const room = rooms.find(r => r.id === roomId);
        if (room) {
            setName(room.name || '');
            setDescription(room.description || '');
        }

        const load = async () => {
            try {
                const [settingsRes, membersRes] = await Promise.all([
                    roomsApi.getSettings(roomId),
                    roomsApi.listMembers(roomId),
                ]);
                if (settingsRes) setSettings(settingsRes);
                if (Array.isArray(membersRes)) setMembers(membersRes);
            } catch (err: unknown) {
                toast.error('加载频道基础信息失败');
            } finally {
                setLoading(false);
            }
        };
        void load();
    }, [roomId, rooms]);

    const handleUpdateRoom = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!canManageSettings) return toast.error('没有权限修改');
        if (!name.trim()) return toast.error('名称不能为空');
        
        setIsSavingRoom(true);
        try {
            await roomsApi.update(roomId, { name: name.trim(), description: description.trim() });
            await fetchRooms();
            toast.success('频道信息已更新');
            setSaveRoomSuccess(true);
            setTimeout(() => setSaveRoomSuccess(false), 2000);
        } catch (err: any) {
            toast.error(err?.message || '更新失败');
        } finally {
            setIsSavingRoom(false);
        }
    };

    const handleSaveSettings = async () => {
        if (!canManageSettings) return toast.error('没有权限');
        setIsSavingSettings(true);
        try {
            await roomsApi.updateSettings(roomId, settings as RoomSettings);
            toast.success('设置已保存');
            setSaveSettingsSuccess(true);
            setTimeout(() => setSaveSettingsSuccess(false), 2000);
        } catch (err: any) {
            toast.error(err?.message || '保存设置失败');
        } finally {
            setIsSavingSettings(false);
        }
    };

    const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        setIsUploadingAvatar(true);
        try {
            const res = await filesApi.upload(file);
            setSettings(s => ({ ...s, avatar_url: res.url }));
            toast.success('头像上传成功，请点击"保存设置"以应用');
        } catch (err: any) {
            toast.error(err?.message || '头像上传失败');
        } finally {
            setIsUploadingAvatar(false);
            if (fileInputRef.current) fileInputRef.current.value = '';
        }
    };

    const handleTransferOwner = async () => {
        if (!isOwner || !transferOwnerId) return;
        if (!window.confirm('确认转让群主身份？转让后你将变为管理员。')) return;
        try {
            await roomsApi.transferOwner(roomId, Number(transferOwnerId));
            await fetchRooms();
            toast.success('群主已转让');
            onClose(); // 转让后直接关闭，防止引发无权限错误
        } catch (err: any) {
            toast.error(err?.message || '转让失败');
        }
    };

    const handleLeaveRoom = async () => {
        if (isOwner) return toast.error('群主无法直接退群');
        if (!window.confirm('确认退出频道？')) return;
        try {
            await roomsApi.leave(roomId);
            leaveRoom(roomId);
            navigate('/chat/home');
            await fetchRooms();
            onClose();
        } catch (err: any) {
            toast.error(err?.message || '退出失败');
        }
    };

    const handleDeleteRoom = async () => {
        if (!isOwner) return;
        if (!window.confirm('确认删除频道？删除后不可恢复。')) return;
        try {
            await roomsApi.remove(roomId);
            navigate('/chat/home');
            await fetchRooms();
            onClose();
        } catch (err: any) {
            toast.error(err?.message || '删除失败');
        }
    };

    if (loading) return <div className="text-sm text-[var(--text-muted)] p-5">加载中...</div>;

    return (
        <div className="space-y-6">
            <section className="bg-[var(--bg-main)] rounded-xl border border-[var(--bg-sidebar)] p-5">
                <div className="text-sm font-bold text-[var(--text-main)] mb-3">频道基础信息</div>
                <form onSubmit={handleUpdateRoom} className="space-y-4">
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">频道名称</label>
                        <input
                            type="text" value={name} onChange={(e) => setName(e.target.value)}
                            className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none"
                            maxLength={50} disabled={!canManageSettings}
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">频道描述</label>
                        <textarea
                            value={description} onChange={(e) => setDescription(e.target.value)}
                            className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none resize-none h-20"
                            maxLength={120} disabled={!canManageSettings}
                        />
                    </div>
                    <div className="flex items-center justify-between">
                        <span className="text-xs text-[var(--text-muted)]">仅群主或管理员可修改</span>
                        <button type="submit" disabled={!canManageSettings || isSavingRoom || saveRoomSuccess} className={`px-4 py-2 rounded-md font-semibold text-sm transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)] disabled:opacity-60 disabled:active:scale-100 ${saveRoomSuccess ? 'bg-green-600 hover:bg-green-500 text-white' : 'bg-[var(--accent)] hover:bg-[#5b4eb3] text-white'}`}>
                            {isSavingRoom ? '保存中...' : saveRoomSuccess ? '已保存' : '保存信息'}
                        </button>
                    </div>
                </form>
            </section>

            <section className="bg-[var(--bg-main)] rounded-xl border border-[var(--bg-sidebar)] p-5">
                <div className="text-sm font-bold text-[var(--text-main)] mb-3">频道主题与头像</div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">频道主题</label>
                        <input
                            type="text" value={settings.topic || ''}
                            onChange={(e) => setSettings(s => ({ ...s, topic: e.target.value }))}
                            className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none"
                            disabled={!canManageSettings}
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">频道头像</label>
                        <div className="flex items-center gap-4">
                            <div 
                                className={`relative w-16 h-16 rounded-xl overflow-hidden shrink-0 border-2 border-[var(--bg-sidebar)] bg-[var(--bg-secondary)] flex items-center justify-center text-2xl text-white font-bold ${canManageSettings ? 'cursor-pointer group' : ''}`}
                                onClick={() => canManageSettings && fileInputRef.current?.click()}
                                title="点击上传新头像"
                            >
                                {settings.avatar_url ? (
                                    <img src={settings.avatar_url} alt="avatar" className="w-full h-full object-cover" />
                                ) : (
                                    <span>{name ? name.charAt(0).toUpperCase() : '#'}</span>
                                )}
                                
                                {canManageSettings && (
                                    <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity flex-col">
                                        <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
                                    </div>
                                )}
                                
                                {isUploadingAvatar && (
                                    <div className="absolute inset-0 bg-black/60 flex items-center justify-center">
                                        <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                                    </div>
                                )}
                            </div>
                            
                            <div className="flex-1 flex flex-col justify-center">
                                {canManageSettings && (
                                    <>
                                        <div className="text-sm text-[var(--text-muted)] mb-2">点击左侧头像上传新图片</div>
                                        {settings.avatar_url && (
                                            <button 
                                                type="button" 
                                                onClick={() => setSettings(s => ({ ...s, avatar_url: '' }))}
                                                className="self-start text-xs text-red-500 hover:text-red-400 font-semibold"
                                            >
                                                恢复默认头像
                                            </button>
                                        )}
                                    </>
                                )}
                                <input
                                    type="file"
                                    accept="image/*"
                                    ref={fileInputRef}
                                    className="hidden"
                                    onChange={handleAvatarUpload}
                                />
                            </div>
                        </div>
                    </div>
                </div>
                <div className="mt-4 flex items-center justify-end">
                    <button type="button" disabled={!canManageSettings || isSavingSettings || saveSettingsSuccess} onClick={handleSaveSettings} className={`px-4 py-2 rounded-md font-semibold text-sm transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)] disabled:opacity-60 disabled:active:scale-100 ${saveSettingsSuccess ? 'bg-green-600 hover:bg-green-500 text-white' : 'bg-[var(--accent)] hover:bg-[#5b4eb3] text-white'}`}>
                        {isSavingSettings ? '保存中...' : saveSettingsSuccess ? '已保存' : '保存设置'}
                    </button>
                </div>
            </section>

            {isOwner && (
                <section className="bg-[var(--bg-main)] rounded-xl border border-[var(--bg-sidebar)] p-5">
                    <div className="text-sm font-bold text-[var(--text-main)] mb-3">转让群主</div>
                    <div className="text-xs text-[var(--text-muted)] mb-3">转让后你将变为管理员，请谨慎操作。将会关闭设置窗口。</div>
                    <div className="flex flex-col sm:flex-row gap-3">
                        <select
                            value={transferOwnerId} onChange={(e) => setTransferOwnerId(e.target.value ? Number(e.target.value) : '')}
                            className="flex-1 bg-[var(--bg-input)] text-[var(--text-main)] p-2 rounded-lg border border-[var(--bg-sidebar)]"
                        >
                            <option value="">选择新的群主</option>
                            {members.filter(m => m.user_id !== currentUser?.id).map(member => (
                                <option key={member.user_id} value={member.user_id}>
                                    {member.user.username} ({member.role === 'admin' ? '管理员' : '成员'})
                                </option>
                            ))}
                        </select>
                        <button type="button" onClick={handleTransferOwner} disabled={!transferOwnerId} className="px-4 py-2 rounded-md bg-[var(--bg-hover)] text-[var(--text-main)] text-sm font-semibold hover:bg-[var(--accent)] hover:text-white transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)] disabled:opacity-60 disabled:active:scale-100">
                            确认转让
                        </button>
                    </div>
                </section>
            )}

            <section className="bg-[#3a1b1b]/40 rounded-xl border border-[#6b2f2f] p-5">
                <div className="text-sm font-bold text-[#ff6b6b] mb-3">危险操作</div>
                <div className="flex flex-col sm:flex-row gap-3">
                    <button type="button" onClick={handleLeaveRoom} className="px-4 py-2 rounded-md border border-[#6b2f2f] text-[#ff6b6b] text-sm font-semibold hover:bg-[#6b2f2f]/30 transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[#ff6b6b] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)]">
                        退出频道
                    </button>
                    {isOwner && (
                        <button type="button" onClick={handleDeleteRoom} className="px-4 py-2 rounded-md border border-[#6b2f2f] text-[#ff6b6b] text-sm font-semibold hover:bg-[#6b2f2f]/30 transition-all active:scale-95 focus:outline-none focus:ring-2 focus:ring-[#ff6b6b] focus:ring-offset-2 focus:ring-offset-[var(--bg-main)]">
                            删除频道
                        </button>
                    )}
                </div>
            </section>
        </div>
    );
}
