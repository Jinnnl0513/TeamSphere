import React, { useState, useRef, useEffect } from 'react';
import { useChatStore, type ChatMessage } from '../../../stores/chatStore';
import { useAuthStore } from '../../../stores/authStore';
import { createPopup } from '@picmo/popup-picker';
import toast from 'react-hot-toast';
import { filesApi } from '../../../services/api/files';
import type { User } from '../../../types/models';
import { MAX_MESSAGE_LENGTH } from '../../../constants';

export interface RoomMemberData {
    user_id: number;
    role: string;
    user: User;
    muted_until?: string | null;
}

interface MessageInputAreaProps {
    roomId: number | null;
    dmId: number | null;
    isDm: boolean;
    titleName: string;
    isConnected: boolean;
    isMuted: boolean;
    myMutedUntilDate: string | null;
    replyingTo: ChatMessage | null;
    setReplyingTo: React.Dispatch<React.SetStateAction<ChatMessage | null>>;
    roomMembers: RoomMemberData[];
}

export default function MessageInputArea({
    roomId,
    dmId,
    isDm,
    titleName,
    isConnected,
    isMuted,
    myMutedUntilDate,
    replyingTo,
    setReplyingTo,
    roomMembers
}: MessageInputAreaProps) {
    const { user: currentUser } = useAuthStore();
    const sendMessage = useChatStore(s => s.sendMessage);
    const sendDmMessage = useChatStore(s => s.sendDmMessage);
    const sendTyping = useChatStore(s => s.sendTyping);
    const sendDmTyping = useChatStore(s => s.sendDmTyping);
    const typingUsers = useChatStore(s => s.typingUsers);

    const [inputValue, setInputValue] = useState('');
    const [mentionState, setMentionState] = useState({ active: false, query: '', position: 0 });
    const [mentionIndex, setMentionIndex] = useState(0);

    const emojiButtonRef = useRef<HTMLButtonElement>(null);

    useEffect(() => {
        if (!emojiButtonRef.current) return;
        
        const picker = createPopup({
            animate: true,
            theme: 'dark'
        }, {
            referenceElement: emojiButtonRef.current,
            triggerElement: emojiButtonRef.current,
            position: 'top-end',
            className: 'custom-picmo-popup',
            showCloseButton: false
        });

        picker.addEventListener('emoji:select', (selection: any) => {
             setInputValue(prev => prev + selection.emoji);
             textareaRef.current?.focus();
        });

        const handleEmojiClick = () => picker.toggle();
        const btn = emojiButtonRef.current;
        btn.addEventListener('click', handleEmojiClick);

        return () => {
            btn.removeEventListener('click', handleEmojiClick);
            picker.destroy();
        };
    }, []);
    const lastTypingTime = useRef<number>(0);
    const textareaRef = useRef<HTMLTextAreaElement>(null);
    const fileInputRef = useRef<HTMLInputElement>(null);
    const mentionableMembers = (!isDm && mentionState.active)
        ? roomMembers
            .filter(m => m.user.id !== currentUser?.id)
            .filter(m => m.user.username.toLowerCase().includes(mentionState.query.toLowerCase()))
        : [];
    const trimmedInput = inputValue.trim();
    const inputLength = Array.from(trimmedInput).length;
    const isOverMessageLimit = inputLength > MAX_MESSAGE_LENGTH;

    useEffect(() => { setMentionIndex(0); }, [mentionState.query]);

    const handleSend = () => {
        const finalMessage = inputValue.trim();
        if (!finalMessage) return;
        if (Array.from(finalMessage).length > MAX_MESSAGE_LENGTH) {
            toast.error(`内容字数需在 ${MAX_MESSAGE_LENGTH} 字以内`);
            return;
        }

        const replyToId = (replyingTo && replyingTo.id > 0) ? replyingTo.id : undefined;

        if (isDm && dmId) {
            sendDmMessage(dmId, finalMessage, 'text', replyToId);
        } else if (!isDm && roomId) {
            sendMessage(roomId, finalMessage, 'text', replyToId);
        } else {
            return;
        }

        setInputValue('');
        setReplyingTo(null);
        if (textareaRef.current) {
            textareaRef.current.style.height = 'auto';
            textareaRef.current.focus();
        }
    };

    const insertMention = (username: string) => {
        const before = inputValue.substring(0, mentionState.position);
        const after = inputValue.substring(mentionState.position + mentionState.query.length + 1);
        setInputValue(`${before}@${username} ${after}`);
        setMentionState({ active: false, query: '', position: 0 });
        textareaRef.current?.focus();
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (mentionState.active && mentionableMembers.length > 0) {
            if (e.key === 'ArrowUp') {
                e.preventDefault();
                setMentionIndex(prev => (prev > 0 ? prev - 1 : mentionableMembers.length - 1));
                return;
            }
            if (e.key === 'ArrowDown') {
                e.preventDefault();
                setMentionIndex(prev => (prev < mentionableMembers.length - 1 ? prev + 1 : 0));
                return;
            }
            if (e.key === 'Enter' || e.key === 'Tab') {
                e.preventDefault();
                insertMention(mentionableMembers[mentionIndex].user.username);
                return;
            }
            if (e.key === 'Escape') {
                setMentionState({ active: false, query: '', position: 0 });
                return;
            }
        }
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            setMentionState({ active: false, query: '', position: 0 });
            handleSend();
        }
    };

    const handleDrop = async (e: React.DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        if (isMuted || !isConnected) return;
        const file = e.dataTransfer.files?.[0];
        if (file) {
            await uploadFile(file);
        }
    };

    const handlePaste = async (e: React.ClipboardEvent<HTMLTextAreaElement>) => {
        const items = e.clipboardData?.items;
        if (!items) return;
        for (const item of items) {
            if (item.type.startsWith('image/')) {
                const file = item.getAsFile();
                if (file) {
                    e.preventDefault();
                    await uploadFile(file);
                    return;
                }
            }
        }
    };


    const handleInput = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        const val = e.target.value;
        setInputValue(val);

        if (textareaRef.current) {
            textareaRef.current.style.height = 'auto';
            textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 400)}px`;
        }

        if (!isDm) {
            const cursor = e.target.selectionStart;
            const textBefore = val.substring(0, cursor);
            const lastAt = textBefore.lastIndexOf('@');
            if (lastAt !== -1 && (lastAt === 0 || textBefore[lastAt - 1] === ' ' || textBefore[lastAt - 1] === '\n')) {
                const query = textBefore.substring(lastAt + 1);
                if (!query.includes(' ') && !query.includes('\n')) {
                    setMentionState({ active: true, query, position: lastAt });
                    return;
                }
            }
            setMentionState({ active: false, query: '', position: 0 });
        }

        const now = Date.now();
        if (now - lastTypingTime.current > 2000 && val.length > 0) {
            lastTypingTime.current = now;
            if (isDm && dmId) sendDmTyping(dmId);
            else if (!isDm && roomId) sendTyping(roomId);
        }
    };

    const uploadFile = async (file: File, msgTypeOverride?: string) => {
        try {
            const res = await filesApi.upload(file);
            const fileUrl = res.url;
            if (!fileUrl) throw new Error('Missing URL in upload response');
            const msgType = msgTypeOverride || (file.type.startsWith('image/') ? 'image' : file.type.startsWith('audio/') ? 'audio' : 'file');
            const fileMeta = { file_size: res.size ?? file.size, mime_type: res.mime_type ?? file.type };
            if (isDm && dmId) sendDmMessage(dmId, fileUrl, msgType, undefined, fileMeta);
            else if (!isDm && roomId) sendMessage(roomId, fileUrl, msgType, undefined, fileMeta);
        } catch (err: unknown) {
            alert('上传失败: ' + ((err as Error).message || '未知错误'));
        }
    };

    const handleUploadFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;
        e.target.value = '';
        await uploadFile(file);
    };

    return (
        <div className="px-4 pb-6 pt-1 shrink-0 z-0 bg-[var(--bg-main)] relative" onDragOver={(e) => e.preventDefault()} onDrop={handleDrop}>
            {mentionState.active && mentionableMembers.length > 0 && (
                <div className="absolute bottom-full mb-2 left-4 w-64 max-h-60 overflow-y-auto custom-scrollbar bg-[var(--bg-sidebar)] border border-[var(--bg-secondary)] rounded-md shadow-2xl z-50 py-1">
                    <div className="px-3 py-1 text-xs font-semibold text-[var(--text-muted)] uppercase">提及成员</div>
                    {mentionableMembers.map((m, idx) => (
                        <div
                            key={m.user_id}
                            className={`px-3 py-1.5 flex items-center cursor-pointer transition-colors ${idx === mentionIndex ? 'bg-[#5865F2]/20 text-[#5865F2]' : 'hover:bg-[var(--bg-main)] text-[var(--text-main)]'}`}
                            onClick={() => insertMention(m.user.username)}
                        >
                            <div className="w-6 h-6 rounded-full bg-[var(--bg-secondary)] overflow-hidden shrink-0 mr-2">
                                <img src={m.user?.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${m.user?.username}`} className="w-full h-full object-cover" />
                            </div>
                            <div className="flex flex-col truncate">
                                <span className={`text-[14px] leading-tight truncate ${idx === mentionIndex ? 'font-bold' : 'font-medium'}`}>{m.user.username}</span>
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {(() => {
                const key = isDm ? `dm_${dmId}` : `room_${roomId}`;
                const activeTypers = (typingUsers[key] || []).filter((u: { username: string, expiresAt: number }) => u.expiresAt > Date.now());
                if (activeTypers.length === 0) return null;
                const text = activeTypers.length <= 3
                    ? activeTypers.map((u: { username: string }) => u.username).join(', ')
                    : `${activeTypers.length} 人`;
                return (
                    <div className="absolute -top-6 left-0 px-2 text-[13px] text-[var(--text-muted)] flex items-center animate-pulse">
                        <svg className="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                        {text} 正在输入...
                    </div>
                );
            })()}

            {replyingTo && (
                <div className="bg-[var(--bg-secondary)] rounded-t-lg px-4 py-2 flex items-center justify-between text-sm text-[var(--text-muted)] -mb-2 relative z-10 border-b border-black/10">
                    <div className="flex items-center space-x-2 truncate">
                        <svg className="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                        </svg>
                        <span className="truncate">
                            回复 <strong>@{replyingTo.user?.username}</strong>：
                            {replyingTo.msg_type === 'image' ? '[图片]' : replyingTo.msg_type === 'file' ? '[文件]' : replyingTo.content}
                        </span>
                    </div>
                    <button onClick={() => setReplyingTo(null)} className="hover:text-[var(--text-main)] transition-colors shrink-0 ml-2">
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
            )}


            <div className={`bg-[var(--bg-input)] px-4 py-2.5 flex items-start space-x-3 w-full shadow-sm max-h-[50vh] relative z-20 ${replyingTo ? 'rounded-b-lg' : 'rounded-lg'}`}>
                <input type="file" className="hidden" ref={fileInputRef} onChange={handleUploadFile} disabled={!isConnected || isMuted} />
                <button
                    onClick={() => fileInputRef.current?.click()}
                    disabled={!isConnected || isMuted}
                    className={`w-6 h-6 rounded-full bg-[var(--bg-secondary)] flex items-center justify-center text-[var(--text-main)] transition-colors shrink-0 mt-0.5 ${isMuted || !isConnected ? 'opacity-50 cursor-not-allowed' : 'hover:text-white hover:bg-[var(--accent)]'}`}
                    title="发送文件"
                >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                    </svg>
                </button>

                <div className="flex-1 overflow-y-auto min-w-0 custom-scrollbar flex items-center">
                    <textarea
                        ref={textareaRef}
                        disabled={!isConnected || isMuted}
                        value={inputValue}
                        onChange={handleInput}
                        onKeyDown={handleKeyDown}
                        onPaste={handlePaste}
                        placeholder={
                            isConnected
                                ? (isMuted && myMutedUntilDate
                                    ? `你已被禁言至 ${new Date(myMutedUntilDate).toLocaleString()}`
                                    : (isDm ? `发送消息给 @${titleName}` : `发送消息到 #${titleName}`))
                                : '正在连接...'
                        }
                        className="w-full bg-transparent border-none outline-none text-[var(--text-main)] placeholder-[var(--text-muted)] text-[15px] resize-none overflow-hidden block custom-scrollbar mt-0.5 break-words disabled:opacity-50"
                        rows={1}
                    />
                </div>

                <div className="flex space-x-3 shrink-0 h-6 mt-0.5 items-center relative">
                    <button
                        ref={emojiButtonRef}
                        disabled={!isConnected || isMuted}
                        className={`transition-all focus:outline-none ${isMuted || !isConnected ? 'text-[var(--text-muted)] opacity-50 cursor-not-allowed' : 'text-[var(--text-muted)] hover:scale-110 hover:text-[var(--text-main)]'}`}
                        title="表情"
                    >
                        🙂
                    </button>
                    <button
                        onClick={handleSend}
                        disabled={!trimmedInput || !isConnected || isMuted || isOverMessageLimit}
                        className={`transition-transform duration-200 flex items-center justify-center rounded ${!trimmedInput || !isConnected || isMuted || isOverMessageLimit ? 'text-[var(--text-muted)] opacity-50 cursor-not-allowed' : 'text-[var(--accent)] hover:scale-110'}`}
                        title="发送 (Enter)"
                    >
                        <svg className="w-5 h-5 -rotate-90" fill="currentColor" viewBox="0 0 20 20">
                            <path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z" />
                        </svg>
                    </button>

                </div>
            </div>
            {inputValue.length > 0 && (
                <div className={`mt-1 text-right text-[11px] ${isOverMessageLimit ? 'text-[#ff6b6b]' : 'text-[var(--text-muted)]'}`}>
                    {inputLength}/{MAX_MESSAGE_LENGTH}
                </div>
            )}
        </div>
    );
}
