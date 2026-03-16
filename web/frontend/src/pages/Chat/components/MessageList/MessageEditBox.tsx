import type { RefObject } from 'react';
import { MAX_MESSAGE_LENGTH } from '../../../../constants';

export default function MessageEditBox({
    editingValue,
    isSavingEdit,
    editingBoxRef,
    editingTextareaRef,
    onChange,
    onCancel,
    onSave,
}: {
    editingValue: string;
    isSavingEdit: boolean;
    editingBoxRef: RefObject<HTMLDivElement | null>;
    editingTextareaRef: RefObject<HTMLTextAreaElement | null>;
    onChange: (val: string) => void;
    onCancel: () => void;
    onSave: () => Promise<void>;
}) {
    const editingLen = Array.from(editingValue).length;
    const overLimit = editingLen > MAX_MESSAGE_LENGTH;

    return (
        <div ref={editingBoxRef} className="mt-1">
            <textarea
                ref={editingTextareaRef}
                value={editingValue}
                onChange={(e) => onChange(e.target.value)}
                disabled={isSavingEdit}
                rows={3}
                className="w-full rounded-md bg-[var(--bg-secondary)] border border-[var(--bg-sidebar)] px-3 py-2 text-[14px] outline-none focus:border-[var(--accent)] resize-y custom-scrollbar"
                onKeyDown={(e) => {
                    if (e.key === 'Escape') {
                        e.preventDefault();
                        onCancel();
                    }
                    if (e.key === 'Enter' && !e.shiftKey) {
                        e.preventDefault();
                        void onSave();
                    }
                }}
            />
            <div className="mt-2 flex items-center gap-2">
                <button
                    type="button"
                    onClick={() => void onSave()}
                    disabled={isSavingEdit || overLimit || editingLen === 0}
                    className="px-3 py-1.5 rounded bg-[var(--accent)] text-white text-xs font-medium hover:opacity-90 disabled:opacity-60"
                >
                    {isSavingEdit ? '保存中...' : '保存'}
                </button>
                <button
                    type="button"
                    onClick={onCancel}
                    disabled={isSavingEdit}
                    className="px-3 py-1.5 rounded bg-[var(--bg-secondary)] text-[var(--text-muted)] text-xs font-medium hover:text-[var(--text-main)] disabled:opacity-60"
                >
                    取消
                </button>
                <span className={`text-[11px] ${overLimit ? 'text-[#ff6b6b]' : 'text-[var(--text-muted)]'}`}>
                    {editingLen}/{MAX_MESSAGE_LENGTH} · Enter 保存，Shift+Enter 换行
                </span>
            </div>
        </div>
    );
}