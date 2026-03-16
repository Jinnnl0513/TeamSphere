import { useState, useEffect, useRef } from 'react';
import { useChatStore, type ChatMessage } from '../../../stores/chatStore';
import { useAuthStore } from '../../../stores/authStore';
import { useFriendStore } from '../../../stores/friendStore';
import MembersSidebar from './MembersSidebar';
import RoomSettingsModal from './RoomSettingsModal';
import MessageList from './MessageList';
import ChatHeader from './ChatHeader';
import ChatWelcomeBanner from './ChatWelcomeBanner';
import MessageInputArea, { type RoomMemberData } from './MessageInputArea';
import ThreadPanel from './ThreadPanel';
import SearchPanel from './SearchPanel';
import { roomsApi } from '../../../services/api/rooms';
import { chatApi } from '../../../services/api/chat';
import toast from 'react-hot-toast';

export default function ChatArea() {
    const currentUser = useAuthStore(s => s.user);

    const activeView = useChatStore(s => s.activeView);
    const activeRoomId = useChatStore(s => s.activeRoomId);
    const activeDmUserId = useChatStore(s => s.activeDmUserId);
    const isConnected = useChatStore(s => s.isConnected);
    const isConnecting = useChatStore(s => s.isConnecting);
    const connect = useChatStore(s => s.connect);
    const joinRoom = useChatStore(s => s.joinRoom);
    const leaveRoom = useChatStore(s => s.leaveRoom);
    const markAsRead = useChatStore(s => s.markAsRead);
    const setUnreadCount = useChatStore(s => s.setUnreadCount);
    const lastReadByRoom = useChatStore(s => s.lastReadByRoom);
    const lastReadByDm = useChatStore(s => s.lastReadByDm);
    const clearUnreadDmCount = useChatStore(s => s.clearUnreadDmCount);
    const mentionFilter = useChatStore(s => s.mentionFilter);
    const setMentionFilter = useChatStore(s => s.setMentionFilter);
    const fetchPinnedMessages = useChatStore(s => s.fetchPinnedMessages);
    const pinnedMessagesByRoom = useChatStore(s => s.pinnedMessages);
    const setActiveThreadMsgId = useChatStore(s => s.setActiveThreadMsgId);

    const isDm = activeView === 'dm';
    const roomId = isDm ? null : activeRoomId;
    const dmId = isDm ? activeDmUserId : null;

    const lastReadMsgId = isDm
        ? (dmId ? (lastReadByDm[dmId] ?? 0) : 0)
        : (roomId ? (lastReadByRoom[roomId] ?? 0) : 0);

    const activeRoom = useChatStore(s => roomId ? s.rooms.find(r => r.id === roomId) : null);
    const dmPeerFromConv = useChatStore(s => dmId ? s.conversations.find(c => c.user.id === dmId)?.user : null);

    const rawMessages = useChatStore(s =>
        isDm ? (dmId ? s.messagesByDm[dmId] : undefined)
            : (roomId ? s.messagesByRoom[roomId] : undefined)
    );
    const messages = rawMessages || [];

    const isPeerOnline = useFriendStore(s => (isDm && dmId) ? s.onlineFriendIds.has(dmId) : false);
    const dmPeerFromFriends = useFriendStore(s => (isDm && dmId) ? s.friends.find(f => f.user.id === dmId)?.user : null);
    const dmPeerInfo = dmPeerFromConv || dmPeerFromFriends;

    const titleName = isDm ? (dmPeerInfo?.username || '') : (activeRoom?.name || '');

    const myMutedUntilDate = useChatStore(s => roomId ? s.myMutedUntilByRoom[roomId] : null);

    const [showMembers, setShowMembers] = useState(true);
    const [isSettingsOpen, setIsSettingsOpen] = useState(false);
    const [roomState, setRoomState] = useState<{
        roomId: number | null;
        members: RoomMemberData[];
        myRole: 'owner' | 'admin' | 'member';
    }>({ roomId: null, members: [], myRole: 'member' });
    const [replyingTo, setReplyingTo] = useState<ChatMessage | null>(null);
    const [nowTs, setNowTs] = useState(() => Date.now());
    const [batchMode, setBatchMode] = useState(false);
    const [isSearchOpen, setIsSearchOpen] = useState(false);

    const myRole = roomId && roomState.roomId === roomId ? roomState.myRole : 'member';
    const roomMembers = roomId && roomState.roomId === roomId ? roomState.members : [];

    const pinnedMessage = roomId ? (pinnedMessagesByRoom[roomId]?.[0] ?? null) : null;

    useEffect(() => {
        let isActive = true;

        if (!roomId || isDm) {
            return () => {
                isActive = false;
            };
        }

        const fetchRole = async () => {
            try {
                const res = await roomsApi.listMembers(roomId);
                if (!isActive) return;

                const members = Array.isArray(res) ? res : [];
                const me = members.find((member: RoomMemberData) => member.user_id === currentUser?.id);

                setRoomState({
                    roomId,
                    members,
                    myRole: me && (me.role === 'owner' || me.role === 'admin') ? me.role : 'member',
                });
                useChatStore.getState().setMyMutedUntil(roomId, me?.muted_until || null);
            } catch (err) {
                console.error('Failed to fetch my role', err);
            }
        };

        void fetchRole();

        return () => {
            isActive = false;
        };
    }, [roomId, isDm, currentUser?.id]);

    useEffect(() => {
        if (!roomId || isDm) return;
        fetchPinnedMessages(roomId);
    }, [roomId, isDm, fetchPinnedMessages]);

    useEffect(() => {
        if (isConnected && roomId && !isDm) {
            joinRoom(roomId);
        }
        return () => {
            if (isConnected && roomId && !isDm) {
                leaveRoom(roomId);
            }
        };
    }, [isConnected, roomId, isDm, joinRoom, leaveRoom]);

    useEffect(() => {
        if (!isDm && roomId && !isConnected && !isConnecting) {
            void connect();
        }
    }, [isDm, roomId, isConnected, isConnecting, connect]);

    useEffect(() => {
        if (!messages.length) return;
        const lastMsg = messages[messages.length - 1];
        if (lastMsg.id <= 0) return;
        markAsRead(isDm, roomId, dmId, lastMsg.id);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [roomId, dmId, isDm]);

    const lastMarkedRoomIdRef = useRef<number | null>(null);
    useEffect(() => {
        if (isDm || !roomId) return;
        if (!messages.length) return;
        if (lastMarkedRoomIdRef.current === roomId) return;
        const lastMsg = messages[messages.length - 1];
        if (lastMsg.id <= 0) return;
        lastMarkedRoomIdRef.current = roomId;
        markAsRead(false, roomId, null, lastMsg.id);
        setUnreadCount(roomId, 0);
        roomsApi.markRead(roomId, lastMsg.id).catch((err: any) => {
            console.error('Failed to mark read', err);
        });
    }, [isDm, roomId, messages.length, markAsRead, setUnreadCount]);

    const lastMarkedDmIdRef = useRef<number | null>(null);
    useEffect(() => {
        if (!isDm || !dmId) return;
        if (!messages.length) return;
        if (lastMarkedDmIdRef.current === dmId) return;
        const lastMsg = messages[messages.length - 1];
        if (lastMsg.id <= 0) return;
        lastMarkedDmIdRef.current = dmId;
        markAsRead(true, null, dmId, lastMsg.id);
        clearUnreadDmCount(dmId);
        chatApi.markDmRead(dmId, lastMsg.id).catch((err: any) => {
            console.error('Failed to mark dm read', err);
        });
    }, [isDm, dmId, messages.length, markAsRead, clearUnreadDmCount]);

    useEffect(() => {
        const timer = window.setInterval(() => setNowTs(Date.now()), 30 * 1000);
        return () => window.clearInterval(timer);
    }, []);

    useEffect(() => {
        if (isDm) {
            setMentionFilter(false);
            setBatchMode(false);
            setActiveThreadMsgId(null);
        }
    }, [isDm, setMentionFilter, setActiveThreadMsgId]);

    if (!roomId && !dmId) {
        return (
            <div className="flex-1 flex flex-col items-center justify-center bg-[var(--bg-main)] min-w-0 min-h-0 z-0 h-full text-[var(--text-muted)]">
                <div className="w-24 h-24 mb-6 opacity-30">
                    <svg fill="currentColor" viewBox="0 0 24 24">
                        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm0-14c-3.31 0-6 2.69-6 6s2.69 6 6 6 6-2.69 6-6-2.69-6-6-6zm0 10c-2.21 0-4-1.79-4-4s1.79-4 4-4 4 1.79 4 4-1.79 4-4 4z" />
                    </svg>
                </div>
                <h2 className="text-xl font-semibold mb-2">没有选择频道或对话</h2>
                <p>请在左侧选择一个可用频道或好友开始聊天。</p>
            </div>
        );
    }

    const isMuted = myMutedUntilDate ? new Date(myMutedUntilDate).getTime() > nowTs : false;
    const canBatch = !isDm && (myRole === 'owner' || myRole === 'admin' || currentUser?.role === 'admin' || currentUser?.role === 'owner' || currentUser?.role === 'system_admin');

    return (
        <div className="flex-1 flex flex-row min-w-0 min-h-0 z-0 h-full">
            <div className="flex-1 flex flex-col bg-[var(--bg-main)] min-w-0 min-h-0 z-0 h-full">
                <ChatHeader
                    isDm={isDm}
                    titleName={titleName}
                    isPeerOnline={isPeerOnline}
                    dmPeerInfo={dmPeerInfo}
                    activeRoom={activeRoom}
                    isConnected={isConnected}
                    showMembers={showMembers}
                    setShowMembers={setShowMembers}
                    setIsSettingsOpen={setIsSettingsOpen}
                    pinnedMessage={pinnedMessage}
                    onOpenSearch={() => setIsSearchOpen(true)}
                    batchMode={batchMode}
                    setBatchMode={canBatch ? setBatchMode : undefined}
                    myRole={myRole}
                    currentUserRole={currentUser?.role}
                />

                {!isDm && (
                    <div className="px-4 pt-3 pb-1 flex items-center gap-2">
                        <button
                            className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${mentionFilter ? 'text-[var(--text-muted)] hover:text-[var(--text-main)] bg-[var(--bg-secondary)]' : 'text-[var(--text-main)] bg-[var(--bg-input)]'}`}
                            onClick={() => setMentionFilter(false)}
                        >
                            全部
                        </button>
                        <button
                            className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${mentionFilter ? 'text-[var(--text-main)] bg-[var(--bg-input)]' : 'text-[var(--text-muted)] hover:text-[var(--text-main)] bg-[var(--bg-secondary)]'}`}
                            onClick={() => setMentionFilter(true)}
                        >
                            @我
                        </button>
                    </div>
                )}

                <ChatWelcomeBanner
                    isDm={isDm}
                    titleName={titleName}
                    dmPeerInfo={dmPeerInfo}
                    activeRoom={activeRoom}
                    myRole={myRole}
                    setIsSettingsOpen={setIsSettingsOpen}
                />

                <MessageList
                    messages={messages}
                    currentUserId={currentUser?.id}
                    isDm={isDm}
                    roomId={roomId}
                    dmId={dmId}
                    myRole={myRole}
                    onReply={setReplyingTo}
                    lastReadMsgId={lastReadMsgId}
                    nowTs={nowTs}
                    mentionFilter={mentionFilter}
                    onViewThread={(msg) => {
                        setActiveThreadMsgId(msg.id);
                        if (!msg.id) toast.error('无效消息');
                    }}
                    batchMode={canBatch ? batchMode : false}
                    setBatchMode={canBatch ? setBatchMode : undefined}
                />

                <MessageInputArea
                    roomId={roomId}
                    dmId={dmId}
                    isDm={isDm}
                    titleName={titleName}
                    isConnected={isConnected}
                    isMuted={isMuted}
                    myMutedUntilDate={myMutedUntilDate}
                    replyingTo={replyingTo}
                    setReplyingTo={setReplyingTo}
                    roomMembers={roomMembers}
                />
            </div>

            {!isDm && <MembersSidebar isVisible={showMembers} />}

            {!isDm && roomId && <ThreadPanel roomId={roomId} />}

            {!isDm && roomId && activeRoom && (
                <RoomSettingsModal
                    isOpen={isSettingsOpen}
                    onClose={() => setIsSettingsOpen(false)}
                    roomId={roomId}
                    roomName={activeRoom.name}
                    roomDescription={activeRoom.description}
                    myRole={myRole}
                />
            )}

            {!isDm && roomId && (
                <SearchPanel
                    isOpen={isSearchOpen}
                    roomId={roomId}
                    onClose={() => setIsSearchOpen(false)}
                    onJumpToMessage={(msgId) => {
                        const el = document.getElementById(`msg-${msgId}`);
                        if (el) {
                            el.scrollIntoView({ behavior: 'smooth', block: 'center' });
                        } else {
                            toast('该消息不在当前加载范围', { icon: 'ℹ️' });
                        }
                        setIsSearchOpen(false);
                    }}
                />
            )}
        </div>
    );
}
