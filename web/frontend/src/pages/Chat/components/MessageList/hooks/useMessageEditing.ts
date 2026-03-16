import { useState, useRef, useEffect, useCallback } from 'react';
import type { RefObject } from 'react';
import type { ChatMessage } from '../../../../../stores/chatStore';
import { MAX_MESSAGE_LENGTH } from '../../../../../constants';
import toast from 'react-hot-toast';

function getDraftKey(isDm: boolean, roomId: number | null, dmId: number | null): string {
    if (isDm && dmId) return `edit_draft_dm_${dmId}`;
    if (!isDm && roomId) return `edit_draft_room_${roomId}`;
    return 'edit_draft_unknown';
}

export interface UseMessageEditingOptions {
    messages: ChatMessage[];
    isDm: boolean;
    roomId: number | null;
    dmId: number | null;
    editMessage: (
        msgId: number,
        content: string,
        roomId: number | null,
        dmUserId: number | null
    ) => Promise<void>;
}

export interface UseMessageEditingReturn {
    editingMsgId: number | null;
    editingValue: string;
    isSavingEdit: boolean;
    editingBoxRef: RefObject<HTMLDivElement | null>;
    editingTextareaRef: RefObject<HTMLTextAreaElement | null>;
    setEditingValue: (val: string) => void;
    startEdit: (msg: ChatMessage) => void;
    cancelEdit: () => void;
    saveEdit: () => Promise<void>;
}

export function useMessageEditing({
    messages,
    isDm,
    roomId,
    dmId,
    editMessage,
}: UseMessageEditingOptions): UseMessageEditingReturn {
    const [editingMsgId, setEditingMsgId] = useState<number | null>(null);
    const [editingValue, setEditingValue] = useState('');
    const [isSavingEdit, setIsSavingEdit] = useState(false);
    const editingBoxRef = useRef<HTMLDivElement | null>(null);
    const editingTextareaRef = useRef<HTMLTextAreaElement | null>(null);

    useEffect(() => {
        setEditingMsgId(null);
        setEditingValue('');
        setIsSavingEdit(false);
    }, [roomId, dmId, isDm]);

    useEffect(() => {
        if (editingMsgId === null) return;
        window.requestAnimationFrame(() => {
            const textarea = editingTextareaRef.current;
            if (!textarea) return;
            const len = textarea.value.length;
            textarea.focus();
            textarea.setSelectionRange(len, len);
        });
    }, [editingMsgId]);

    useEffect(() => {
        if (editingMsgId === null) return;
        const onMouseDown = (e: MouseEvent) => {
            if (isSavingEdit) return;
            const target = e.target as Node | null;
            if (editingBoxRef.current && target && !editingBoxRef.current.contains(target)) {
                if (editingValue.trim()) {
                    const key = getDraftKey(isDm, roomId, dmId);
                    sessionStorage.setItem(key, JSON.stringify({ msgId: editingMsgId, value: editingValue }));
                }
                setEditingMsgId(null);
                setEditingValue('');
            }
        };
        window.addEventListener('mousedown', onMouseDown);
        return () => window.removeEventListener('mousedown', onMouseDown);
    }, [editingMsgId, isSavingEdit, editingValue, isDm, roomId, dmId]);

    const startEdit = useCallback((msg: ChatMessage) => {
        const key = getDraftKey(isDm, roomId, dmId);
        let initialValue = msg.content;
        try {
            const raw = sessionStorage.getItem(key);
            if (raw) {
                const draft = JSON.parse(raw) as { msgId: number; value: string };
                if (draft.msgId === msg.id && draft.value) {
                    initialValue = draft.value;
                }
            }
        } catch {
            // ignore
        }
        setEditingMsgId(msg.id);
        setEditingValue(initialValue);
    }, [isDm, roomId, dmId]);

    const cancelEdit = useCallback(() => {
        if (isSavingEdit) return;
        const key = getDraftKey(isDm, roomId, dmId);
        sessionStorage.removeItem(key);
        setEditingMsgId(null);
        setEditingValue('');
    }, [isSavingEdit, isDm, roomId, dmId]);

    const saveEdit = useCallback(async () => {
        if (editingMsgId === null) return;
        const next = editingValue.trim();
        const nextLen = Array.from(next).length;
        const target = messages.find(m => m.id === editingMsgId);

        if (!target) {
            cancelEdit();
            return;
        }

        if (!next || next === target.content) {
            cancelEdit();
            return;
        }

        if (nextLen > MAX_MESSAGE_LENGTH) {
            toast.error(`内容字数需在 ${MAX_MESSAGE_LENGTH} 字以内`);
            return;
        }

        try {
            setIsSavingEdit(true);
            await editMessage(editingMsgId, next, roomId, dmId);
            const key = getDraftKey(isDm, roomId, dmId);
            sessionStorage.removeItem(key);
            setEditingMsgId(null);
            setEditingValue('');
        } catch (e: any) {
            toast.error(e?.response?.data?.message || e?.message || '编辑失败');
        } finally {
            setIsSavingEdit(false);
        }
    }, [editingMsgId, editingValue, messages, editMessage, roomId, dmId, isDm, cancelEdit]);

    return {
        editingMsgId,
        editingValue,
        isSavingEdit,
        editingBoxRef,
        editingTextareaRef,
        setEditingValue,
        startEdit,
        cancelEdit,
        saveEdit,
    };
}
