import type { ChatMessage } from '../../../../stores/chatStore';

export interface MessageContextMenuState {
    x: number;
    y: number;
    msg: ChatMessage;
}

export default function MessageContextMenu({
    menu,
    currentUserId,
    isDm,
    myRole,
    currentUserRole,
    nowTs,
    onReply,
    onEdit,
    onRetract,
    onViewThread,
    onForward,
    isPinned,
    onPin,
    onUnpin,
    isBlocked,
    onBlock,
    onUnblock,
}: {
    menu: MessageContextMenuState;
    currentUserId?: number;
    isDm: boolean;
    myRole: string;
    currentUserRole?: string;
    nowTs: number;
    onReply: (msg: ChatMessage) => void;
    onEdit: (msg: ChatMessage) => void;
    onRetract: (msg: ChatMessage) => void;
    onViewThread: (msg: ChatMessage) => void;
    onForward: (msg: ChatMessage) => void;
    isPinned: boolean;
    onPin: (msg: ChatMessage) => void;
    onUnpin: (msg: ChatMessage) => void;
    isBlocked: boolean;
    onBlock: (msg: ChatMessage) => void;
    onUnblock: (msg: ChatMessage) => void;
}) {
    const msg = menu.msg;
    const withinTimeLimit = new Date(msg.created_at).getTime() > nowTs - 2 * 60 * 1000;
    const isSelf = msg.user?.id === currentUserId;
    const isRoomAdmin = !isDm && (myRole === 'owner' || myRole === 'admin');
    const isSysAdmin = currentUserRole === 'admin' || currentUserRole === 'owner' || currentUserRole === 'system_admin';
    const canRetract = msg.id > 0 && ((isSelf && withinTimeLimit) || isRoomAdmin || isSysAdmin);
    const canEdit = msg.id > 0 && isSelf && !msg.deleted_at && msg.msg_type === 'text';
    const canPin = !isDm && msg.id > 0 && (isRoomAdmin || isSysAdmin);
    const canThread = !isDm && msg.id > 0 && !msg.deleted_at;
    const canBlock = !!msg.user?.id && !isSelf;
    const canForward = msg.id > 0 && !msg.deleted_at;

    return (
        <div
            className="fixed z-[999] bg-[var(--bg-main)] border border-[var(--bg-secondary)] rounded-md shadow-xl py-1 w-40 overflow-hidden animate-in fade-in duration-100"
            style={{ top: menu.y, left: menu.x }}
            onContextMenu={(e) => e.preventDefault()}
        >
            <button
                className="w-full text-left px-4 py-2 text-sm text-[var(--text-main)] hover:bg-[#5865F2] hover:text-white transition-colors flex items-center"
                onClick={() => onReply(msg)}
            >
                <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                </svg>
                回复
            </button>

            {canThread && (
                <button
                    className="w-full text-left px-4 py-2 text-sm text-[var(--text-main)] hover:bg-[#4e5d94] hover:text-white transition-colors flex items-center"
                    onClick={() => onViewThread(msg)}
                >
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 8h10M7 12h6m-8 8l2-2h10a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                    查看线程
                </button>
            )}

            {canForward && (
                <button
                    className="w-full text-left px-4 py-2 text-sm text-[var(--text-main)] hover:bg-[#3b414a] hover:text-white transition-colors flex items-center"
                    onClick={() => onForward(msg)}
                >
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h10a2 2 0 012 2v10m-4-4L7 7m0 0v6m0-6h6" />
                    </svg>
                    转发到...
                </button>
            )}

            {canEdit && (
                <button
                    className="w-full text-left px-4 py-2 text-sm text-[var(--text-main)] hover:bg-[#f0b232] hover:text-white transition-colors flex items-center"
                    onClick={() => onEdit(msg)}
                >
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                    编辑消息
                </button>
            )}

            {canPin && (
                <button
                    className="w-full text-left px-4 py-2 text-sm text-[var(--text-main)] hover:bg-[#2f3136] hover:text-white transition-colors flex items-center"
                    onClick={() => (isPinned ? onUnpin(msg) : onPin(msg))}
                >
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 3l12 12m-6 6l6-6m-6 6V9m0 12l-4-4" />
                    </svg>
                    {isPinned ? '取消置顶' : '置顶消息'}
                </button>
            )}

            {canRetract && (
                <button
                    className="w-full text-left px-4 py-2 text-sm text-[#ff6b6b] hover:bg-[#ff6b6b] hover:text-white transition-colors flex items-center"
                    onClick={() => onRetract(msg)}
                >
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                    撤回消息
                </button>
            )}

            {canBlock && (
                <button
                    className="w-full text-left px-4 py-2 text-sm text-[var(--text-main)] hover:bg-[#3b414a] hover:text-white transition-colors flex items-center"
                    onClick={() => (isBlocked ? onUnblock(msg) : onBlock(msg))}
                >
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 5.636l-12.728 12.728M6.343 6.343a8 8 0 1111.314 11.314 8 8 0 01-11.314-11.314z" />
                    </svg>
                    {isBlocked ? '取消屏蔽' : '屏蔽用户'}
                </button>
            )}
        </div>
    );
}
