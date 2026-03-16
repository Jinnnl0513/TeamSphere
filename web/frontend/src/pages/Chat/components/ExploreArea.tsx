
import { useState, useEffect, useCallback } from 'react';
import { useChatStore } from '../../../stores/chatStore';
import type { Room } from '../../../types/models';
import { roomsApi } from '../../../services/api/rooms';
import toast from 'react-hot-toast';

import { useNavigate } from 'react-router-dom';

export default function ExploreArea() {
    const navigate = useNavigate();
    const { fetchRooms } = useChatStore();
    const [allRooms, setAllRooms] = useState<Room[]>([]);
    const [loading, setLoading] = useState(false);

    const [showCreateModal, setShowCreateModal] = useState(false);
    const [newRoomName, setNewRoomName] = useState('');
    const [newRoomDesc, setNewRoomDesc] = useState('');
    const [createError, setCreateError] = useState('');
    const [isCreating, setIsCreating] = useState(false);

    const loadRooms = useCallback(async () => {
        setLoading(true);
        try {
            const res = await roomsApi.discover();
            setAllRooms(res || []);
            fetchRooms();
        } catch (err) {
            console.error('Fetch discoverable rooms failed:', err);
        } finally {
            setLoading(false);
        }
    }, [fetchRooms]);

    useEffect(() => {
        void loadRooms();
    }, [loadRooms]);

    const handleJoin = async (e: React.MouseEvent, roomId: number) => {
        e.stopPropagation();
        try {
            await roomsApi.join(roomId);
            await fetchRooms();
            navigate(`/chat/rooms/${roomId}`);
        } catch (err: any) {
            const code = err?.code;
            if (code === 20201) {
                toast.success('加入申请已提交，等待审批');
                return;
            }
            if (code === 40306) {
                toast.error('该频道为私有，需邀请加入');
                return;
            }
            if (code === 40904) {
                toast('已有待处理的申请');
                return;
            }
            toast.error(err?.message || '加入频道失败');
        }
    };

    const handleCreateRoom = async (e: React.FormEvent) => {
        e.preventDefault();
        setCreateError('');
        if (!newRoomName.trim()) return;

        setIsCreating(true);
        try {
            const res = await roomsApi.create({
                name: newRoomName.trim(),
                description: newRoomDesc.trim() || '欢迎来到我的新频道！',
            });
            await fetchRooms();
            setShowCreateModal(false);
            if (res && res.id) {
                navigate(`/chat/rooms/${res.id}`);
            }
        } catch (err: any) {
            setCreateError(err?.message || '无法创建频道，请稍后重试');
        } finally {
            setIsCreating(false);
        }
    };

    return (
        <div className="flex-1 bg-[var(--bg-main)] min-w-0 z-0 h-screen overflow-y-auto custom-scrollbar relative">
            <div className="px-12 py-10 max-w-6xl mx-auto">
                <div className="flex items-center justify-between mb-8">
                    <h2 className="text-xl font-bold text-[var(--text-main)] transition-colors">推荐的公开频道 ({allRooms.length})</h2>
                    <button
                        onClick={() => setShowCreateModal(true)}
                        className="bg-[var(--accent)] hover:bg-[#5b4eb3] text-white px-5 py-2 rounded-lg text-sm font-semibold transition-all shadow-md hover:shadow-lg hover:-translate-y-0.5 flex items-center"
                    >
                        <span className="text-lg mr-1 leading-none font-bold">+</span>
                        创建新频道
                    </button>
                </div>

                {loading ? (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                        {[...Array(8)].map((_, i) => (
                            <div key={i} className="h-44 rounded-2xl bg-[var(--bg-secondary)] border border-[var(--bg-sidebar)] animate-pulse"></div>
                        ))}
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                        <div
                            onClick={() => setShowCreateModal(true)}
                            className="h-44 rounded-2xl border-2 border-dashed border-[var(--bg-sidebar)] hover:border-[var(--accent)]/50 bg-transparent flex flex-col items-center justify-center cursor-pointer transition-colors group"
                        >
                            <div className="w-12 h-12 rounded-full bg-[var(--bg-secondary)] group-hover:bg-[var(--accent)]/10 text-[var(--accent)] flex items-center justify-center text-3xl font-light mb-2 transition-colors">
                                +
                            </div>
                            <span className="font-semibold text-[var(--text-main)] group-hover:text-[var(--accent)] transition-colors">创建属于你的空间</span>
                        </div>

                        {allRooms.map((room) => (
                            <div key={room.id} className="h-44 rounded-2xl bg-[var(--bg-secondary)] border border-[var(--bg-sidebar)] hover:border-[var(--accent)]/40 hover:shadow-xl hover:-translate-y-1 cursor-pointer transition-all duration-300 flex flex-col relative overflow-hidden group">
                                <div className="h-16 bg-gradient-to-r from-[var(--bg-sidebar)] to-[var(--bg-secondary)] relative">
                                    <div className="absolute top-2 right-2 px-2 py-0.5 rounded-full bg-black/30 backdrop-blur-sm text-[10px] font-bold text-white shadow-sm flex items-center">
                                        <div className="w-1.5 h-1.5 rounded-full bg-[#20c997] mr-1.5 animate-pulse"></div>
                                        活跃中
                                    </div>
                                </div>

                                <div className="absolute top-8 left-4 p-1 bg-[var(--bg-secondary)] rounded-[12px] group-hover:rotate-6 transition-transform">
                                    <div className="w-[52px] h-[52px] rounded-[10px] bg-gradient-to-br from-[#5b4eb3] to-[#7f71cd] flex items-center justify-center text-white text-2xl font-bold shadow-inner">
                                        {room.name[0]?.toUpperCase() || '#'}
                                    </div>
                                </div>

                                <div className="pt-10 px-4 pb-4 flex-1 flex flex-col">
                                    <div className="flex items-center justify-between mb-1">
                                        <h3 className="font-bold text-[var(--text-main)] text-[15px] truncate max-w-[150px]">{room.name}</h3>
                                    </div>
                                    <p className="text-xs text-[var(--text-muted)] line-clamp-2 leading-relaxed flex-1">
                                        {room.description || '一个还没有描述的隐秘角落。快来探索吧？'}
                                    </p>

                                    <div className="mt-3 flex items-center justify-between">
                                        <div className="text-[11px] font-bold text-[var(--text-muted)] bg-[var(--bg-main)] px-2 py-1 rounded border border-[var(--bg-sidebar)]">ID: {room.id}</div>
                                        <button
                                            onClick={(e) => handleJoin(e, room.id)}
                                            className="text-white text-xs font-bold bg-[var(--bg-hover)] hover:bg-[var(--accent)] px-4 py-1.5 rounded-full transition-colors"
                                        >
                                            进入
                                        </button>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            {showCreateModal && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
                    <div className="bg-[var(--bg-secondary)] w-full max-w-md rounded-2xl shadow-2xl overflow-hidden border border-[var(--bg-sidebar)] flex flex-col transform animate-in zoom-in-95 duration-200">
                        <div className="px-6 py-5 border-b border-[var(--bg-sidebar)] font-bold text-xl text-[var(--text-main)] flex justify-between items-center text-center">
                            <span className="w-full">创建一处新天地</span>
                            <button
                                onClick={() => setShowCreateModal(false)}
                                className="absolute right-4 text-[var(--text-muted)] hover:text-[#ff6b6b] transition-colors p-1"
                            >
                                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                            </button>
                        </div>

                        <div className="p-6">
                            <p className="text-sm text-[var(--text-muted)] mb-6 text-center">赋予你的专属频道一个与众不同的个性和主题。完成之后，所有人都可以找到这片乐土。</p>

                            <form onSubmit={handleCreateRoom} className="space-y-4">
                                <div>
                                    <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">频道名称 <span className="text-[#ff6b6b]">*</span></label>
                                    <input
                                        type="text"
                                        value={newRoomName}
                                        onChange={(e) => setNewRoomName(e.target.value)}
                                        className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none transition-shadow"
                                        placeholder="例如: 极客交流基地"
                                        maxLength={50}
                                        required
                                        autoFocus
                                    />
                                </div>
                                <div>
                                    <label className="block text-xs font-bold text-[var(--text-muted)] uppercase mb-2">描述主题 (可选)</label>
                                    <textarea
                                        value={newRoomDesc}
                                        onChange={(e) => setNewRoomDesc(e.target.value)}
                                        className="w-full bg-[var(--bg-input)] text-[var(--text-main)] p-3 rounded-lg border-none focus:ring-2 focus:ring-[var(--accent)] outline-none transition-shadow resize-none h-20 placeholder-[var(--text-muted)]/50"
                                        placeholder="一句话介绍一下这个频道的主题..."
                                        maxLength={100}
                                    />
                                </div>

                                {createError && (
                                    <div className="flex items-center text-[#ff6b6b] text-sm bg-[#ff6b6b]/10 p-2 rounded">
                                        <svg className="w-4 h-4 mr-1 shrink-0" fill="currentColor" viewBox="0 0 20 20"><path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd"></path></svg>
                                        {createError}
                                    </div>
                                )}

                                <div className="mt-8 bg-[var(--bg-sidebar)] -mx-6 -mb-6 p-4 flex justify-end space-x-3 rounded-b-2xl">
                                    <button
                                        type="button"
                                        onClick={() => setShowCreateModal(false)}
                                        className="px-5 py-2 hover:underline text-[var(--text-main)] transition-colors text-sm font-semibold"
                                    >
                                        取消
                                    </button>
                                    <button
                                        type="submit"
                                        disabled={isCreating || !newRoomName.trim()}
                                        className="bg-[var(--accent)] hover:bg-[#5b4eb3] text-white px-6 py-2 rounded-md font-bold text-sm transition-colors shadow-sm disabled:opacity-50 flex items-center"
                                    >
                                        {isCreating ? '创建中...' : '创建频道'}
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
