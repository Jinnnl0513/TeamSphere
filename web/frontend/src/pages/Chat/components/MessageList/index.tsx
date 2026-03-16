import { useState, useEffect, useMemo, useRef } from 'react';
import type { MouseEvent } from 'react';
import { format, isToday, isYesterday } from 'date-fns';
import { useChatStore, type ChatMessage } from '../../../../stores/chatStore';
import { useAuthStore } from '../../../../stores/authStore';
import type { User } from '../../../../types/models';
import toast from 'react-hot-toast';
import UserProfileModal from '../UserProfileModal';
import MessageListContainer from './MessageListContainer';
import MessageItem from './MessageItem';
import MessageContextMenu, { type MessageContextMenuState } from './MessageContextMenu';
import ExternalLinkGuard from './ExternalLinkGuard';
import ImagePreviewModal from './ImagePreviewModal';
import ForwardModal from './ForwardModal';
import { useScrollBehavior } from './hooks/useScrollBehavior';
import { useVirtualizer } from '@tanstack/react-virtual';
import { useMessageEditing } from './hooks/useMessageEditing';
import { useExternalLinks } from './hooks/useExternalLinks';
import { resolveFileUrl } from '../../../../utils/urls';

interface MessageListProps {
    messages: ChatMessage[];
    currentUserId?: number;
    isDm: boolean;
    roomId: number | null;
    dmId: number | null;
    myRole: 'owner' | 'admin' | 'member';
    onReply: (msg: ChatMessage) => void;
    onViewThread: (msg: ChatMessage) => void;
    lastReadMsgId?: number;
    nowTs?: number;
    mentionFilter?: boolean;
    batchMode?: boolean;
    setBatchMode?: (val: boolean) => void;
}

export default function MessageList({
    messages,
    currentUserId,
    isDm,
    roomId,
    dmId,
    myRole,
    onReply,
    onViewThread,
    lastReadMsgId = 0,
    nowTs = Date.now(),
    mentionFilter = false,
    batchMode = false,
    setBatchMode,
}: MessageListProps) {
    const retractMessage = useChatStore(s => s.retractMessage);
    const editMessage = useChatStore(s => s.editMessage);
    const setMessageReactions = useChatStore(s => s.setMessageReactions);
    const pinnedMessagesByRoom = useChatStore(s => s.pinnedMessages);
    const pinMessage = useChatStore(s => s.pinMessage);
    const unpinMessage = useChatStore(s => s.unpinMessage);
    const fetchOlderHistory = useChatStore(s => s.fetchOlderHistory);
    const fetchOlderDmHistory = useChatStore(s => s.fetchOlderDmHistory);
    const hasMoreByRoom = useChatStore(s => s.hasMoreByRoom);
    const hasMoreByDm = useChatStore(s => s.hasMoreByDm);
    const isLoadingOlderByRoom = useChatStore(s => s.isLoadingOlderByRoom);
    const isLoadingOlderByDm = useChatStore(s => s.isLoadingOlderByDm);
    const blockedUserIds = useChatStore(s => s.blockedUserIds);
    const blockUser = useChatStore(s => s.blockUser);
    const unblockUser = useChatStore(s => s.unblockUser);
    const batchDeleteMessages = useChatStore(s => s.batchDeleteMessages);
    const { user: currentUser } = useAuthStore();

    const [previewImage, setPreviewImage] = useState<string | null>(null);
    const [selectedUser, setSelectedUser] = useState<User | null>(null);
    const [contextMenu, setContextMenu] = useState<MessageContextMenuState | null>(null);
    const [selectedIds, setSelectedIds] = useState<number[]>([]);
    const [forwardingMsg, setForwardingMsg] = useState<ChatMessage | null>(null);
    const unreadDividerRef = useRef<HTMLDivElement | null>(null);
    const lastUnreadScrollKey = useRef<string | null>(null);

    const blockedSet = useMemo(() => new Set(blockedUserIds), [blockedUserIds]);

    useEffect(() => {
        if (!batchMode) setSelectedIds([]);
    }, [batchMode]);

    const hasMore = isDm ? hasMoreByDm[dmId!] : hasMoreByRoom[roomId!];
    const shouldScrollToUnread = lastReadMsgId > 0 && messages.some(msg => msg.id > lastReadMsgId);
    const isLoadingOlder = isDm ? isLoadingOlderByDm[dmId!] : isLoadingOlderByRoom[roomId!];

    const filteredMessages = useMemo(() => {
        let list = messages.filter(msg => {
            const uid = msg.user?.id ?? 0;
            return uid <= 0 || !blockedSet.has(uid);
        });
        if (!isDm && mentionFilter && currentUser?.username) {
            const token = `@${currentUser.username}`;
            list = list.filter(msg => (msg.content || '').includes(token));
        }
        return list;
    }, [messages, mentionFilter, currentUser?.username, isDm, blockedSet]);

    const { listRef, showScrollToBottom, handleScroll, scrollToBottom } = useScrollBehavior({
        messages: filteredMessages,
        isDm,
        roomId,
        dmId,
        hasMore,
        isLoadingOlder,
        skipInitialScroll: shouldScrollToUnread,
        fetchOlderHistory,
        fetchOlderDmHistory,
    });

    const {
        editingMsgId,
        editingValue,
        isSavingEdit,
        editingBoxRef,
        editingTextareaRef,
        setEditingValue,
        startEdit,
        cancelEdit,
        saveEdit,
    } = useMessageEditing({
        messages,
        isDm,
        roomId,
        dmId,
        editMessage,
    });

    const {
        externalFileAction,
        externalLinkToVisit,
        openFileAction,
        closeFileAction,
        visitFile,
        downloadFile,
        openLinkConfirm,
        closeLinkConfirm,
        visitLink,
    } = useExternalLinks();

    useEffect(() => {
        const handle = () => setContextMenu(null);
        window.addEventListener('click', handle);
        return () => window.removeEventListener('click', handle);
    }, []);

    const formatDateGroup = (dateStr: string) => {
        if (!dateStr) return '未知日期';
        const d = new Date(dateStr);
        if (isToday(d)) return '今天';
        if (isYesterday(d)) return '昨天';
        return format(d, 'yyyy年M月d日');
    };

    const handleRetract = (msg: ChatMessage) => {
        retractMessage(msg.id, roomId, dmId).catch(e => toast.error(e.message || '撤回失败'));
        setContextMenu(null);
    };

    const handleEdit = (msg: ChatMessage) => {
        setContextMenu(null);
        startEdit(msg);
    };

    const handleOpenContextMenu = (msg: ChatMessage, e: MouseEvent) => {
        if (msg.deleted_at) return;
        e.preventDefault();
        e.stopPropagation();
        const menuWidth = 160;
        let x = e.clientX;
        if (window.innerWidth - x < menuWidth) x = window.innerWidth - menuWidth;
        setContextMenu({ x, y: e.clientY, msg });
    };

    const pinnedIds = useMemo(() => {
        if (!roomId) return new Set<number>();
        return new Set((pinnedMessagesByRoom[roomId] || []).map(m => m.id));
    }, [pinnedMessagesByRoom, roomId]);

    const imageUrls = useMemo(() => {
        return filteredMessages
            .filter((msg) => msg.msg_type === 'image' && !!msg.content)
            .map((msg) => resolveFileUrl(msg.content as string));
    }, [filteredMessages]);

    const renderedMessages = useMemo(() => {
        return filteredMessages.map((msg, index) => {
            const prevMsg = index > 0 ? filteredMessages[index - 1] : null;
            const msgGroupDate = formatDateGroup(msg.created_at);
            const prevGroupDate = prevMsg ? formatDateGroup(prevMsg.created_at) : null;
            const showDateDivider = msgGroupDate !== prevGroupDate;
            const diffTime = prevMsg
                ? new Date(msg.created_at).getTime() - new Date(prevMsg.created_at).getTime()
                : 0;
            const isTimeFar = diffTime > 5 * 60 * 1000;
            const showHeader = showDateDivider || isTimeFar || !prevMsg || prevMsg.user?.id !== msg.user?.id;
            const dateRaw = msg.created_at ? new Date(msg.created_at) : new Date();
            const timeStr = format(dateRaw, 'HH:mm');
            const isEdited = !!msg.updated_at && msg.updated_at !== msg.created_at;
            const editedAtText = msg.updated_at ? format(new Date(msg.updated_at), 'yyyy/MM/dd HH:mm:ss') : '';

            const showUnreadDivider =
                lastReadMsgId > 0 &&
                msg.id > 0 &&
                msg.id > lastReadMsgId &&
                (prevMsg === null || prevMsg.id <= lastReadMsgId);

            return {
                key: msg.id || msg.client_msg_id || index,
                msg,
                showDateDivider,
                msgGroupDate,
                showUnreadDivider,
                showHeader,
                dateRaw,
                timeStr,
                isEdited,
                editedAtText,
            };
        });
    }, [filteredMessages, lastReadMsgId, nowTs]);

    const shouldVirtualize = filteredMessages.length >= 500;
    const rowVirtualizer = useVirtualizer({
        count: renderedMessages.length,
        getScrollElement: () => listRef.current,
        estimateSize: () => 72,
        overscan: 8,
    });

    useEffect(() => {
        if (!shouldScrollToUnread) return;
        const key = `${isDm ? 'dm' : 'room'}:${roomId ?? 'na'}:${dmId ?? 'na'}`;
        if (lastUnreadScrollKey.current === key) return;
        const el = unreadDividerRef.current;
        if (!el) return;
        el.scrollIntoView({ block: 'center' });
        lastUnreadScrollKey.current = key;
    }, [isDm, roomId, dmId, shouldScrollToUnread, messages.length]);

    const toggleSelected = (msgId: number) => {
        if (msgId <= 0) return;
        setSelectedIds(prev => (prev.includes(msgId) ? prev.filter(id => id !== msgId) : [...prev, msgId]));
    };

    const handleBatchDelete = async () => {
        if (!roomId || selectedIds.length === 0) return;
        try {
            await batchDeleteMessages(roomId, selectedIds);
            toast.success(`已删除 ${selectedIds.length} 条消息`);
            setSelectedIds([]);
            setBatchMode?.(false);
        } catch (e: any) {
            toast.error(e?.message || '批量删除失败');
        }
    };

    return (
        <>
            <div className="flex-1 min-h-0 relative">
                {batchMode && !isDm && (
                    <div className="absolute top-2 right-4 z-20 bg-[var(--bg-secondary)] text-[var(--text-main)] text-sm px-3 py-2 rounded-md flex items-center gap-3 shadow">
                        <span>已选 {selectedIds.length} 条</span>
                        <button
                            className="text-[#ff6b6b] hover:underline"
                            onClick={handleBatchDelete}
                            disabled={selectedIds.length === 0}
                        >
                            删除所选
                        </button>
                        <button
                            className="text-[var(--text-muted)] hover:underline"
                            onClick={() => setBatchMode?.(false)}
                        >
                            退出
                        </button>
                    </div>
                )}

                <MessageListContainer
                    listRef={listRef}
                    onScroll={handleScroll}
                    isLoadingOlder={isLoadingOlder}
                    hasMore={!!hasMore}
                    messagesLength={filteredMessages.length}
                    isVirtualized={shouldVirtualize}
                >
                    {shouldVirtualize ? (
                        <div
                            className="relative w-full"
                            style={{ height: rowVirtualizer.getTotalSize() }}
                        >
                            {rowVirtualizer.getVirtualItems().map((virtualRow) => {
                                const item = renderedMessages[virtualRow.index];
                                return (
                                    <div
                                        key={item.key}
                                        data-index={virtualRow.index}
                                        ref={rowVirtualizer.measureElement}
                                        className="absolute left-0 top-0 w-full"
                                        style={{ transform: `translateY(${virtualRow.start}px)` }}
                                    >
                                        <MessageItem
                                            msg={item.msg}
                                            showDateDivider={item.showDateDivider}
                                            msgGroupDate={item.msgGroupDate}
                                            showUnreadDivider={item.showUnreadDivider}
                                            unreadDividerRef={item.showUnreadDivider ? unreadDividerRef : undefined}
                                            showHeader={item.showHeader}
                                            dateRaw={item.dateRaw}
                                            timeStr={item.timeStr}
                                            isEdited={item.isEdited}
                                            editedAtText={item.editedAtText}
                                            currentUsername={currentUser?.username}
                                            currentUserId={currentUserId}
                                            isDm={isDm}
                                            roomId={roomId}
                                            dmId={dmId}
                                            editingMsgId={editingMsgId}
                                            editingValue={editingValue}
                                            isSavingEdit={isSavingEdit}
                                            editingBoxRef={editingBoxRef}
                                            editingTextareaRef={editingTextareaRef}
                                            setEditingValue={setEditingValue}
                                            cancelEdit={cancelEdit}
                                            saveEdit={saveEdit}
                                            onOpenFileAction={openFileAction}
                                            onOpenLinkConfirm={openLinkConfirm}
                                            onPreviewImage={setPreviewImage}
                                            onSelectUser={setSelectedUser}
                                            onOpenContextMenu={handleOpenContextMenu}
                                            onReactionChange={(mid, reactions) =>
                                                setMessageReactions(mid, reactions ?? [], isDm, dmId)
                                            }
                                            showSelection={batchMode}
                                            isSelected={selectedIds.includes(item.msg.id)}
                                            onToggleSelect={toggleSelected}
                                        />
                                    </div>
                                );
                            })}
                        </div>
                    ) : (
                        renderedMessages.map((item) => (
                            <MessageItem
                                key={item.key}
                                msg={item.msg}
                                showDateDivider={item.showDateDivider}
                                msgGroupDate={item.msgGroupDate}
                                showUnreadDivider={item.showUnreadDivider}
                                unreadDividerRef={item.showUnreadDivider ? unreadDividerRef : undefined}
                                showHeader={item.showHeader}
                                dateRaw={item.dateRaw}
                                timeStr={item.timeStr}
                                isEdited={item.isEdited}
                                editedAtText={item.editedAtText}
                                currentUsername={currentUser?.username}
                                currentUserId={currentUserId}
                                isDm={isDm}
                                roomId={roomId}
                                dmId={dmId}
                                editingMsgId={editingMsgId}
                                editingValue={editingValue}
                                isSavingEdit={isSavingEdit}
                                editingBoxRef={editingBoxRef}
                                editingTextareaRef={editingTextareaRef}
                                setEditingValue={setEditingValue}
                                cancelEdit={cancelEdit}
                                saveEdit={saveEdit}
                                onOpenFileAction={openFileAction}
                                onOpenLinkConfirm={openLinkConfirm}
                                onPreviewImage={setPreviewImage}
                                onSelectUser={setSelectedUser}
                                onOpenContextMenu={handleOpenContextMenu}
                                onReactionChange={(mid, reactions) =>
                                    setMessageReactions(mid, reactions ?? [], isDm, dmId)
                                }
                                showSelection={batchMode}
                                isSelected={selectedIds.includes(item.msg.id)}
                                onToggleSelect={toggleSelected}
                            />
                        ))
                    )}
                </MessageListContainer>

                {showScrollToBottom && (
                    <button
                        onClick={() => scrollToBottom()}
                        className="absolute right-5 bottom-4 z-20 h-12 w-12 rounded-full border-2 border-[var(--accent)] text-[var(--accent)] hover:border-[var(--text-main)] hover:text-[var(--text-main)] transition-colors flex items-center justify-center"
                        title="回到底部"
                    >
                        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
                        </svg>
                    </button>
                )}
            </div>

            {previewImage && (
                <ImagePreviewModal
                    src={previewImage}
                    images={imageUrls}
                    onChange={setPreviewImage}
                    onClose={() => setPreviewImage(null)}
                />
            )}

            {contextMenu && (
                <MessageContextMenu
                    menu={contextMenu}
                    currentUserId={currentUserId}
                    isDm={isDm}
                    myRole={myRole}
                    currentUserRole={currentUser?.role}
                    nowTs={nowTs}
                    onReply={(msg) => {
                        onReply(msg);
                        setContextMenu(null);
                    }}
                    onViewThread={(msg) => {
                        onViewThread(msg);
                        setContextMenu(null);
                    }}
                    onForward={(msg) => {
                        setForwardingMsg(msg);
                        setContextMenu(null);
                    }}
                    isPinned={!!(roomId && pinnedIds.has(contextMenu.msg.id))}
                    onPin={(msg) => {
                        if (!roomId) return;
                        pinMessage(roomId, msg.id)
                            .then(() => toast.success('已置顶'))
                            .catch((e) => toast.error(e?.message || '置顶失败'));
                        setContextMenu(null);
                    }}
                    onUnpin={(msg) => {
                        if (!roomId) return;
                        unpinMessage(roomId, msg.id)
                            .then(() => toast.success('已取消置顶'))
                            .catch((e) => toast.error(e?.message || '取消置顶失败'));
                        setContextMenu(null);
                    }}
                    isBlocked={!!contextMenu.msg.user?.id && blockedSet.has(contextMenu.msg.user.id)}
                    onBlock={(msg) => {
                        if (msg.user?.id) blockUser(msg.user.id);
                        toast.success('已屏蔽该用户');
                        setContextMenu(null);
                    }}
                    onUnblock={(msg) => {
                        if (msg.user?.id) unblockUser(msg.user.id);
                        toast.success('已取消屏蔽');
                        setContextMenu(null);
                    }}
                    onEdit={handleEdit}
                    onRetract={handleRetract}
                />
            )}

            <ExternalLinkGuard
                externalFileAction={externalFileAction}
                externalLinkToVisit={externalLinkToVisit}
                onCancelFile={closeFileAction}
                onVisitFile={visitFile}
                onDownloadFile={() => void downloadFile()}
                onCancelLink={closeLinkConfirm}
                onVisitLink={visitLink}
            />

            <UserProfileModal
                user={selectedUser}
                isOpen={!!selectedUser}
                onClose={() => setSelectedUser(null)}
            />

            <ForwardModal
                isOpen={!!forwardingMsg}
                sourceType={isDm ? 'dm' : 'room'}
                message={forwardingMsg}
                onClose={() => setForwardingMsg(null)}
            />
        </>
    );
}
