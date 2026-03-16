import { createPortal } from 'react-dom';
import type { ExternalFileAction } from './hooks/useExternalLinks';

function ExternalFileActionModal({
    fileName,
    url,
    onCancel,
    onVisit,
    onDownload,
}: {
    fileName: string;
    url: string;
    onCancel: () => void;
    onVisit: () => void;
    onDownload: () => void;
}) {
    return createPortal(
        <div
            className="fixed inset-0 z-[2000] flex items-center justify-center bg-black/70 p-4 backdrop-blur-sm"
            onClick={onCancel}
        >
            <div
                className="w-full max-w-md rounded-xl border border-[var(--bg-secondary)] bg-[var(--bg-main)] p-5 shadow-2xl"
                onClick={(e) => e.stopPropagation()}
            >
                <h3 className="text-base font-semibold text-[var(--text-main)]">外部链接操作确认</h3>
                <p className="mt-2 text-sm text-[var(--text-muted)] break-all">文件：{fileName}</p>
                <p className="mt-1 text-xs text-[var(--text-muted)] break-all opacity-80">{url}</p>
                <div className="mt-5 flex flex-wrap items-center justify-end gap-2">
                    <button
                        type="button"
                        onClick={onCancel}
                        className="rounded-md border border-[var(--bg-secondary)] px-3 py-1.5 text-sm text-[var(--text-muted)] hover:text-[var(--text-main)] transition-colors"
                    >
                        取消
                    </button>
                    <button
                        type="button"
                        onClick={onVisit}
                        className="rounded-md border border-[#5865F2]/50 px-3 py-1.5 text-sm text-[#8ea1ff] hover:bg-[#5865F2]/15 transition-colors"
                    >
                        访问
                    </button>
                    <button
                        type="button"
                        onClick={onDownload}
                        className="rounded-md bg-[var(--accent)] px-3 py-1.5 text-sm font-medium text-white hover:opacity-90 transition-opacity"
                    >
                        下载
                    </button>
                </div>
            </div>
        </div>,
        document.body
    );
}

function ExternalLinkConfirmModal({
    url,
    onCancel,
    onVisit,
}: {
    url: string;
    onCancel: () => void;
    onVisit: () => void;
}) {
    return createPortal(
        <div
            className="fixed inset-0 z-[2000] flex items-center justify-center bg-black/70 p-4 backdrop-blur-sm"
            onClick={onCancel}
        >
            <div
                className="w-full max-w-md rounded-xl border border-[var(--bg-secondary)] bg-[var(--bg-main)] p-5 shadow-2xl"
                onClick={(e) => e.stopPropagation()}
            >
                <h3 className="text-base font-semibold text-[var(--text-main)]">外部链接访问确认</h3>
                <p className="mt-2 text-sm text-[var(--text-muted)] break-all">即将访问：</p>
                <p className="mt-1 text-xs text-[var(--text-muted)] break-all opacity-80">{url}</p>
                <div className="mt-5 flex flex-wrap items-center justify-end gap-2">
                    <button
                        type="button"
                        onClick={onCancel}
                        className="rounded-md border border-[var(--bg-secondary)] px-3 py-1.5 text-sm text-[var(--text-muted)] hover:text-[var(--text-main)] transition-colors"
                    >
                        取消
                    </button>
                    <button
                        type="button"
                        onClick={onVisit}
                        className="rounded-md bg-[var(--accent)] px-3 py-1.5 text-sm font-medium text-white hover:opacity-90 transition-opacity"
                    >
                        访问
                    </button>
                </div>
            </div>
        </div>,
        document.body
    );
}

export default function ExternalLinkGuard({
    externalFileAction,
    externalLinkToVisit,
    onCancelFile,
    onVisitFile,
    onDownloadFile,
    onCancelLink,
    onVisitLink,
}: {
    externalFileAction: ExternalFileAction | null;
    externalLinkToVisit: string | null;
    onCancelFile: () => void;
    onVisitFile: () => void;
    onDownloadFile: () => void;
    onCancelLink: () => void;
    onVisitLink: () => void;
}) {
    return (
        <>
            {externalFileAction && (
                <ExternalFileActionModal
                    fileName={externalFileAction.fileName}
                    url={externalFileAction.url}
                    onCancel={onCancelFile}
                    onVisit={onVisitFile}
                    onDownload={onDownloadFile}
                />
            )}
            {externalLinkToVisit && (
                <ExternalLinkConfirmModal
                    url={externalLinkToVisit}
                    onCancel={onCancelLink}
                    onVisit={onVisitLink}
                />
            )}
        </>
    );
}
