import type { ReplyInfo } from '../../../../stores/chatStore';

export default function ReplyCard({ replyTo }: { replyTo: ReplyInfo }) {
    const displayName = replyTo.user?.username || '未知用户';
    let preview: string;
    if (replyTo.is_deleted) preview = '此消息已被撤回';
    else if (replyTo.msg_type === 'image') preview = '[图片]';
    else if (replyTo.msg_type === 'file') preview = '[文件]';
    else preview = (replyTo.content || '').slice(0, 120) + ((replyTo.content || '').length > 120 ? '...' : '');

    return (
        <div className="border-l-[3px] border-[#5865F2] pl-3 py-1 mb-1.5 mt-0.5 text-[13px] bg-black/10 rounded-r select-none flex items-start gap-1.5">
            <svg className="w-3.5 h-3.5 shrink-0 mt-0.5 text-[#5865F2]/60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
            </svg>
            <div className="flex flex-col min-w-0">
                <span className="font-semibold text-[#5865F2] text-[12px] truncate">{displayName}</span>
                <span className={`text-[var(--text-muted)] truncate ${replyTo.is_deleted ? 'italic' : ''}`}>{preview}</span>
            </div>
        </div>
    );
}