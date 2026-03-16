import { format, isToday, isYesterday } from 'date-fns';
import type { MouseEvent, RefObject, ReactNode } from 'react';
import type { ChatMessage } from '../../../../stores/chatStore';
import type { User } from '../../../../types/models';
import { TRUSTED_DOMAINS, API_BASE_URL } from '../../../../config/app';
import MessageReactions from '../MessageReactions';
import ReplyCard from './ReplyCard';
import MessageEditBox from './MessageEditBox';
import { inferFileNameFromUrl } from './hooks/useExternalLinks';
import { resolveFileUrl, isUploadsPath } from '../../../../utils/urls';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';

export function isSafeUrl(url: string): boolean {
    if (!url) return false;
    if (isUploadsPath(url)) return true;
    try {
        const parsed = new URL(url, window.location.origin);
        if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') return false;
        const host = parsed.hostname.toLowerCase();
        const currentHost = window.location.hostname.toLowerCase();
        const apiHost = new URL(API_BASE_URL, window.location.origin).hostname.toLowerCase();
        if (host === currentHost) return true;
        if (host === apiHost) return true;
        if (TRUSTED_DOMAINS.length === 0) return true;
        return TRUSTED_DOMAINS.some((domain: string) => host === domain || host.endsWith(`.${domain}`));
    } catch {
        return false;
    }
}

function isHttpUrl(text: string): boolean {
    return /^https?:\/\/\S+$/i.test(text);
}

function formatFileSize(size: number): string {
    if (!Number.isFinite(size) || size <= 0) return '';
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let idx = 0;
    let value = size;
    while (value >= 1024 && idx < units.length - 1) {
        value /= 1024;
        idx++;
    }
    const precision = value >= 10 || idx === 0 ? 0 : 1;
    return `${value.toFixed(precision)} ${units[idx]}`;
}

export function renderTextWithMentionsAndLinks(
    text: string,
    currentUsername: string | undefined,
    onOpenLinkConfirm: (url: string) => void
) {
    const urlRegex = /(https?:\/\/\S+)/gi;
    const chunks = text.split(urlRegex);

    return chunks.map((chunk, i) => {
        if (isHttpUrl(chunk) && isSafeUrl(chunk)) {
            return (
                <button
                    key={`url-${i}`}
                    type="button"
                    onClick={() => onOpenLinkConfirm(chunk)}
                    className="underline underline-offset-2 text-[#8ea1ff] hover:text-[#b4beff] transition-colors break-all"
                    title={chunk}
                >
                    {chunk}
                </button>
            );
        }

        return chunk.split(/(@[A-Za-z0-9_]+)/g).map((part, j) => {
            if (part.startsWith('@')) {
                const isMe = part === `@${currentUsername}`;
                return (
                    <span
                        key={`mention-${i}-${j}`}
                        className={`font-medium px-1 rounded-sm cursor-pointer ${isMe ? 'bg-[var(--accent)]/30 text-[var(--accent)]' : 'bg-[#5865F2]/20 text-[#5865F2] hover:underline'}`}
                    >
                        {part}
                    </span>
                );
            }
            return <span key={`text-${i}-${j}`}>{part}</span>;
        });
    });
}

function getForwardPreview(meta: ChatMessage['forward_meta']) {
    if (!meta) return '';
    if (meta.is_deleted) return '原消息已被撤回';
    if (meta.msg_type === 'image') return '[图片]';
    if (meta.msg_type === 'file') return '[文件]';
    if (meta.msg_type === 'audio') return '[语音]';
    return meta.content || '';
}

export default function MessageItem({
    msg,
    showDateDivider,
    msgGroupDate,
    showUnreadDivider,
    unreadDividerRef,
    showHeader,
    dateRaw,
    timeStr,
    isEdited,
    editedAtText,
    currentUsername,
    currentUserId,
    isDm,
    roomId,
    dmId,
    editingMsgId,
    editingValue,
    isSavingEdit,
    editingBoxRef,
    editingTextareaRef,
    setEditingValue,
    cancelEdit,
    saveEdit,
    onOpenFileAction,
    onOpenLinkConfirm,
    onPreviewImage,
    onSelectUser,
    onOpenContextMenu,
    onReactionChange,
    showSelection,
    isSelected,
    onToggleSelect,
}: {
    msg: ChatMessage;
    showDateDivider: boolean;
    msgGroupDate: string;
    showUnreadDivider: boolean;
    unreadDividerRef?: RefObject<HTMLDivElement | null>;
    showHeader: boolean;
    dateRaw: Date;
    timeStr: string;
    isEdited: boolean;
    editedAtText: string;
    currentUsername?: string;
    currentUserId?: number;
    isDm: boolean;
    roomId: number | null;
    dmId: number | null;
    editingMsgId: number | null;
    editingValue: string;
    isSavingEdit: boolean;
    editingBoxRef: RefObject<HTMLDivElement | null>;
    editingTextareaRef: RefObject<HTMLTextAreaElement | null>;
    setEditingValue: (val: string) => void;
    cancelEdit: () => void;
    saveEdit: () => Promise<void>;
    onOpenFileAction: (url: string) => void;
    onOpenLinkConfirm: (url: string) => void;
    onPreviewImage: (src: string) => void;
    onSelectUser: (user: User | null) => void;
    onOpenContextMenu: (msg: ChatMessage, e: MouseEvent) => void;
    onReactionChange: (msgId: number, reactions: ChatMessage['reactions']) => void;
    showSelection?: boolean;
    isSelected?: boolean;
    onToggleSelect?: (msgId: number) => void;
}) {
    const isMentionedMe = !!currentUsername && (msg.content || '').includes(`@${currentUsername}`);
    const baseBg = isMentionedMe
        ? 'bg-[#f0b232]/10 hover:bg-[#f0b232]/20 border-l-[3px] border-[#f0b232]'
        : 'hover:bg-[var(--bg-secondary)]/30 border-l-[3px] border-transparent';

    const selectableId = msg.id > 0 ? `msg-${msg.id}` : undefined;
    const forwardPreview = getForwardPreview(msg.forward_meta);

    return (
        <div className="relative w-full" id={selectableId} data-msg-id={msg.id}>
            {showDateDivider && (
                <div className="relative flex items-center w-full my-4 pointer-events-none">
                    <div className="flex-grow border-t border-[var(--bg-secondary)]" />
                    <span className="mx-2 text-xs font-semibold text-[var(--text-muted)] bg-[var(--bg-main)] px-2">
                        {msgGroupDate}
                    </span>
                    <div className="flex-grow border-t border-[var(--bg-secondary)]" />
                </div>
            )}

            {showUnreadDivider && (
                <div
                    ref={unreadDividerRef}
                    className="relative flex items-center w-full my-2 pointer-events-none select-none"
                    aria-label="以下为未读消息"
                >
                    <div className="flex-grow border-t border-[#f04747]/60" />
                    <span className="mx-2 text-[11px] font-semibold text-[#f04747] bg-[var(--bg-main)] px-2 whitespace-nowrap">
                        —— 新消息 ——
                    </span>
                    <div className="flex-grow border-t border-[#f04747]/60" />
                </div>
            )}

            <div
                className={`flex w-full px-2 py-0.5 mt-[1px] group relative transition-colors ${baseBg} ${showHeader && !showDateDivider ? 'mt-4' : ''} ${msg.id < 0 ? 'opacity-70' : ''}`}
                onContextMenu={(e) => onOpenContextMenu(msg, e)}
            >
                {showSelection && msg.id > 0 && (
                    <div className="w-6 flex items-center justify-center pt-1">
                        <input
                            type="checkbox"
                            className="accent-[var(--accent)]"
                            checked={!!isSelected}
                            onChange={() => onToggleSelect?.(msg.id)}
                            onClick={(e) => e.stopPropagation()}
                        />
                    </div>
                )}

                {showHeader ? (
                    <div
                        className="w-10 h-10 rounded-full bg-gradient-to-br from-[var(--bg-sidebar)] to-[var(--bg-secondary)] overflow-hidden shadow-inner flex-shrink-0 mt-1 cursor-pointer"
                        onClick={() => onSelectUser(msg.user || null)}
                    >
                        <img
                            src={msg.user?.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${msg.user?.username}`}
                            alt={msg.user?.username}
                            className="w-full h-full object-cover"
                        />
                    </div>
                ) : (
                    <div className="w-10 flex-shrink-0 flex items-center justify-end pr-1 invisible group-hover:visible pt-1 select-none">
                        <span className="text-[10px] text-[var(--text-muted)]">{timeStr}</span>
                    </div>
                )}

                <div className="flex-1 ml-4 pr-12 min-w-0">
                    {showHeader && (
                        <div className="flex items-baseline space-x-2">
                            <span
                                className="font-semibold text-[15px] hover:underline cursor-pointer truncate flex items-baseline"
                                onClick={() => onSelectUser(msg.user || null)}
                            >
                                <span>{msg.user?.username || 'System'}</span>
                            </span>
                            <span className="text-xs text-[var(--text-muted)] flex-shrink-0 leading-none">
                                {isToday(dateRaw)
                                    ? `今天 ${timeStr}`
                                    : isYesterday(dateRaw)
                                        ? `昨天 ${timeStr}`
                                        : format(dateRaw, 'yyyy/MM/dd HH:mm')}
                            </span>
                            {isEdited && !msg.deleted_at && (
                                <span
                                    className="text-[11px] text-[var(--text-muted)] opacity-70 cursor-help"
                                    title={`编辑于 ${editedAtText}`}
                                >
                                    (已编辑)
                                </span>
                            )}
                            {isDm && msg.user?.id === currentUserId && msg.read_at && (
                                <span className="text-[11px] text-[var(--text-muted)] opacity-70 ml-1">
                                    已读
                                </span>
                            )}
                        </div>
                    )}

                    {msg.deleted_at ? (
                        <div className="text-[var(--text-muted)] text-[14px] italic mt-0.5 flex items-center">
                            <svg className="w-3.5 h-3.5 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                            此消息已被撤回
                        </div>
                    ) : msg.msg_type === 'system' ? (
                        <div className="text-[var(--text-muted)] text-[14px] italic mt-0.5 flex items-center">
                            <span className="font-semibold text-yellow-500 mr-2">系统提示</span>
                            {msg.content}
                        </div>
                    ) : msg.forward_meta ? (
                        <div className="mt-1">
                            {msg.reply_to && <ReplyCard replyTo={msg.reply_to} />}
                            {msg.content && (
                                <div className="text-[var(--text-main)] text-[15px] leading-relaxed break-words whitespace-pre-wrap mb-2">
                                    <ReactMarkdown
                                        remarkPlugins={[remarkGfm]}
                                        rehypePlugins={[rehypeHighlight]}
                                    >
                                        {msg.content}
                                    </ReactMarkdown>
                                </div>
                            )}
                            <div className="border border-[var(--bg-secondary)] rounded-md bg-[var(--bg-secondary)]/40 px-3 py-2 text-sm">
                                <div className="text-[11px] text-[var(--text-muted)] mb-1">转发自 @{msg.forward_meta.user?.username || '未知用户'}</div>
                                <div className="text-[var(--text-main)] break-words whitespace-pre-wrap text-[14px]">
                                    {forwardPreview}
                                </div>
                            </div>
                        </div>
                    ) : msg.msg_type === 'image' ? (
                        <div className="mt-1">
                            {msg.reply_to && <ReplyCard replyTo={msg.reply_to} />}
                            {isSafeUrl(msg.content) ? (
                                <img
                                    src={resolveFileUrl(msg.content)}
                                    alt="Image"
                                    className="max-h-80 max-w-sm rounded cursor-pointer object-cover hover:opacity-90 transition-opacity"
                                    onClick={() => onPreviewImage(resolveFileUrl(msg.content))}
                                    onError={(e: any) => {
                                        (e.target as HTMLImageElement).style.display = 'none';
                                    }}
                                />
                            ) : (
                                <span className="text-[var(--text-muted)] italic text-sm">[无效的图片链接]</span>
                            )}
                        </div>
                    ) : msg.msg_type === 'audio' ? (
                        <div className="mt-1">
                            {msg.reply_to && <ReplyCard replyTo={msg.reply_to} />}
                            {isSafeUrl(msg.content) ? (
                                <audio controls src={resolveFileUrl(msg.content)} className="w-64"></audio>
                            ) : (
                                <span className="text-[var(--text-muted)] italic text-sm">[\u65e0\u6548\u7684\u97f3\u9891\u94fe\u63a5]</span>
                            )}
                        </div>
                    ) : msg.msg_type === 'file' ? (
                        <div className="mt-1">
                            {msg.reply_to && <ReplyCard replyTo={msg.reply_to} />}
                            {isSafeUrl(msg.content) ? (
                                <button
                                    type="button"
                                    onClick={() => void onOpenFileAction(resolveFileUrl(msg.content))}
                                    className="inline-flex items-center gap-2 rounded-md border border-[var(--bg-sidebar)] bg-[var(--bg-secondary)] px-3 py-2 text-sm text-[var(--text-main)] hover:bg-[var(--bg-main)] transition-colors max-w-full"
                                >
                                    <span className="shrink-0">📎</span>
                                    <span className="flex flex-col min-w-0">
                                        <span className="truncate">{inferFileNameFromUrl(msg.content)}</span>
                                        {(msg.file_size || msg.mime_type) && (
                                            <span className="text-[11px] text-[var(--text-muted)]">
                                                {msg.file_size ? formatFileSize(msg.file_size) : ''}
                                                {msg.file_size && msg.mime_type ? ' · ' : ''}
                                                {msg.mime_type || ''}
                                            </span>
                                        )}
                                    </span>
                                </button>
                            ) : (
                                <span className="text-[var(--text-muted)] italic text-sm">[无效的文件链接]</span>
                            )}
                        </div>
                    ) : (
                        <div className="text-[var(--text-main)] text-[15px] leading-relaxed break-words whitespace-pre-wrap">
                            {msg.reply_to && <ReplyCard replyTo={msg.reply_to} />}
                            {editingMsgId === msg.id ? (
                                <MessageEditBox
                                    editingValue={editingValue}
                                    isSavingEdit={isSavingEdit}
                                    editingBoxRef={editingBoxRef}
                                    editingTextareaRef={editingTextareaRef}
                                    onChange={setEditingValue}
                                    onCancel={cancelEdit}
                                    onSave={saveEdit}
                                />
                            ) : (
                                <ReactMarkdown
                                    remarkPlugins={[remarkGfm]}
                                    rehypePlugins={[rehypeHighlight]}
                                    components={{
                                        a: ({ href, children }) => {
                                            const safe = href && isSafeUrl(href);
                                            if (!href || !safe) {
                                                return <span className="underline underline-offset-2 text-[#8ea1ff]">{children}</span>;
                                            }
                                            return (
                                                <button
                                                    type="button"
                                                    onClick={() => onOpenLinkConfirm(href)}
                                                    className="underline underline-offset-2 text-[#8ea1ff] hover:text-[#b4beff] transition-colors break-all"
                                                >
                                                    {children}
                                                </button>
                                            );
                                        },
                                        code: ({ inline, className, children }: { inline?: boolean; className?: string; children?: ReactNode }) => {
                                            return inline ? (
                                                <code className="px-1 py-0.5 rounded bg-[var(--bg-secondary)] text-[13px]">{children}</code>
                                            ) : (
                                                <code className={className}>{children}</code>
                                            );
                                        },
                                    }}
                                >
                                    {msg.content || ''}
                                </ReactMarkdown>
                            )}
                        </div>
                    )}
                </div>

                {!msg.deleted_at && msg.id > 0 && msg.msg_type !== 'system' && (
                    <MessageReactions
                        msgId={msg.id}
                        reactions={msg.reactions}
                        isDm={isDm}
                        roomId={roomId}
                        dmUserId={dmId}
                        onReactionChange={(mid, reactions) => onReactionChange(mid, reactions)}
                    />
                )}
            </div>
        </div>
    );
}
