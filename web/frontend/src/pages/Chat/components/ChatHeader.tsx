import React from 'react';
import type { ChatMessage } from '../../../stores/chat/chatStore.types';
import type { User, Room } from '../../../types/models';

interface ChatHeaderProps {
    isDm: boolean;
    titleName: string;
    isPeerOnline: boolean;
    dmPeerInfo: User | undefined | null;
    activeRoom: Room | undefined | null;
    isConnected: boolean;
    showMembers: boolean;
    setShowMembers: React.Dispatch<React.SetStateAction<boolean>>;
    setIsSettingsOpen: React.Dispatch<React.SetStateAction<boolean>>;
    pinnedMessage?: ChatMessage | null;
    onOpenSearch?: () => void;
    batchMode?: boolean;
    setBatchMode?: (val: boolean) => void;
    myRole?: 'owner' | 'admin' | 'member';
    currentUserRole?: string;
}

export default function ChatHeader({
    isDm,
    titleName,
    isPeerOnline,
    dmPeerInfo,
    activeRoom,
    isConnected,
    showMembers,
    setShowMembers,
    setIsSettingsOpen,
    pinnedMessage,
    onOpenSearch,
    batchMode,
    setBatchMode,
}: ChatHeaderProps) {
    return (
        <div className="border-b border-[var(--bg-sidebar)] shrink-0 shadow-sm z-10 w-full bg-[var(--bg-main)]">
            <div className="h-12 flex items-center px-4">
                <span className="text-xl text-[var(--text-muted)] mr-2 leading-none font-light">
                    {isDm ? '@' : '#'}
                </span>
                <span className="font-semibold text-[var(--text-main)] text-base mr-2">{titleName}</span>
                {isDm && (
                    <div
                        title={isPeerOnline ? '在线' : '离线'}
                        className={`w-2.5 h-2.5 rounded-full shrink-0 mr-4 ${isPeerOnline ? 'bg-[var(--color-discord-green-500)]' : 'bg-gray-500 border border-black/20'}`}
                    />
                )}
                <div className="w-[1px] h-6 bg-[var(--bg-secondary)] mx-2" />
                <span className="text-sm font-medium text-[var(--text-muted)] hidden sm:block whitespace-nowrap overflow-hidden text-ellipsis mr-4">
                    {isDm
                        ? (dmPeerInfo?.username ? `你与 ${dmPeerInfo.username} 的私聊` : '私聊')
                        : (activeRoom?.description || '这里是频道讨论区。')}
                </span>
                {!isConnected && (
                    <span className="text-xs font-semibold text-[var(--color-discord-red-400)] border border-[var(--color-discord-red-400)]/30 bg-[var(--color-discord-red-500)]/10 px-2 py-0.5 rounded animate-pulse">
                        连接中...
                    </span>
                )}
                <div className="ml-auto flex items-center space-x-4">
                    {!isDm && (
                        <button
                            className="text-[var(--text-muted)] hover:text-[var(--text-main)] transition-colors"
                            onClick={() => onOpenSearch?.()}
                            title="搜索消息"
                        >
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-4.35-4.35m1.35-5.65a7 7 0 11-14 0 7 7 0 0114 0z" />
                            </svg>
                        </button>
                    )}
                    {!isDm && setBatchMode && (
                        <button
                            className={`text-sm px-2 py-1 rounded ${batchMode ? 'bg-[var(--bg-secondary)] text-[var(--text-main)]' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'}`}
                            onClick={() => setBatchMode(!batchMode)}
                            title="批量管理"
                        >
                            批量
                        </button>
                    )}
                    {!isDm && (
                        <button
                            className="text-[var(--text-muted)] hover:text-[var(--text-main)] transition-colors"
                            onClick={() => setIsSettingsOpen(true)}
                            title="设置"
                        >
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                            </svg>
                        </button>
                    )}
                    {!isDm && (
                        <button
                            className={`transition-colors ${showMembers ? 'text-[var(--text-main)]' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'}`}
                            onClick={() => setShowMembers(prev => !prev)}
                            title="成员列表"
                        >
                            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                                <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z" />
                            </svg>
                        </button>
                    )}
                </div>
            </div>
            {pinnedMessage && !isDm && (
                <div className="px-4 py-2 border-t border-[var(--bg-secondary)] text-sm text-[var(--text-muted)] flex items-center gap-2">
                    <span className="text-[var(--accent)]">置顶</span>
                    <span className="truncate">{pinnedMessage.content || '[附件]'}</span>
                </div>
            )}
        </div>
    );
}
