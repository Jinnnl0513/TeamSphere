import { useState, useEffect, useCallback } from 'react';
import toast from 'react-hot-toast';
import { resolveFileUrl } from '../../../../../utils/urls';

export function inferFileNameFromUrl(url: string): string {
    url = resolveFileUrl(url);
    const cleaned = url.split('#')[0].split('?')[0];
    const lastPart = cleaned.split('/').pop() || '';
    try {
        return decodeURIComponent(lastPart) || '下载文件';
    } catch {
        return lastPart || '下载文件';
    }
}

async function fetchBlobWithAuth(url: string): Promise<Blob> {
    const token = localStorage.getItem('token');
    const resp = await fetch(url, token ? { headers: { Authorization: `Bearer ${token}` } } : undefined);
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    return await resp.blob();
}

async function openBlobInNewTab(blob: Blob): Promise<void> {
    const objectUrl = URL.createObjectURL(blob);
    const newTab = window.open(objectUrl, '_blank', 'noopener,noreferrer');
    if (!newTab) {
        URL.revokeObjectURL(objectUrl);
        throw new Error('popup_blocked');
    }
    setTimeout(() => URL.revokeObjectURL(objectUrl), 60 * 1000);
}

async function triggerFileDownload(url: string): Promise<void> {
    url = resolveFileUrl(url);
    const fileName = inferFileNameFromUrl(url);
    try {
        const blob = await fetchBlobWithAuth(url);
        const objectUrl = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = objectUrl;
        a.download = fileName;
        document.body.appendChild(a);
        a.click();
        a.remove();
        URL.revokeObjectURL(objectUrl);
    } catch {
        const a = document.createElement('a');
        a.href = url;
        a.download = fileName;
        document.body.appendChild(a);
        a.click();
        a.remove();
    }
}

export interface ExternalFileAction {
    url: string;
    fileName: string;
}

export interface UseExternalLinksReturn {
    externalFileAction: ExternalFileAction | null;
    externalLinkToVisit: string | null;
    openFileAction: (url: string) => void;
    closeFileAction: () => void;
    visitFile: () => void;
    downloadFile: () => Promise<void>;
    openLinkConfirm: (url: string) => void;
    closeLinkConfirm: () => void;
    visitLink: () => void;
}

export function useExternalLinks(): UseExternalLinksReturn {
    const [externalFileAction, setExternalFileAction] = useState<ExternalFileAction | null>(null);
    const [externalLinkToVisit, setExternalLinkToVisit] = useState<string | null>(null);

    useEffect(() => {
        if (!externalFileAction) return;
        const onKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'Escape') setExternalFileAction(null);
        };
        window.addEventListener('keydown', onKeyDown);
        return () => window.removeEventListener('keydown', onKeyDown);
    }, [externalFileAction]);

    useEffect(() => {
        if (!externalLinkToVisit) return;
        const onKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'Escape') setExternalLinkToVisit(null);
        };
        window.addEventListener('keydown', onKeyDown);
        return () => window.removeEventListener('keydown', onKeyDown);
    }, [externalLinkToVisit]);

    const openFileAction = useCallback((url: string) => {
        const resolved = resolveFileUrl(url);
        setExternalFileAction({ url: resolved, fileName: inferFileNameFromUrl(resolved) });
    }, []);

    const closeFileAction = useCallback(() => {
        setExternalFileAction(null);
    }, []);

    const visitFile = useCallback(() => {
        if (!externalFileAction) return;
        const url = externalFileAction.url;
        fetchBlobWithAuth(url)
            .then((blob) => openBlobInNewTab(blob))
            .catch(() => {
                // Fallback to direct open (may 401 if auth required)
                window.open(url, '_blank', 'noopener,noreferrer');
            })
            .finally(() => {
                setExternalFileAction(null);
            });
    }, [externalFileAction]);

    const downloadFile = useCallback(async () => {
        if (!externalFileAction) return;
        try {
            await triggerFileDownload(externalFileAction.url);
        } catch {
            toast.error('下载失败');
        } finally {
            setExternalFileAction(null);
        }
    }, [externalFileAction]);

    const openLinkConfirm = useCallback((url: string) => {
        setExternalLinkToVisit(url);
    }, []);

    const closeLinkConfirm = useCallback(() => {
        setExternalLinkToVisit(null);
    }, []);

    const visitLink = useCallback(() => {
        if (!externalLinkToVisit) return;
        window.open(externalLinkToVisit, '_blank', 'noopener,noreferrer');
        setExternalLinkToVisit(null);
    }, [externalLinkToVisit]);

    return {
        externalFileAction,
        externalLinkToVisit,
        openFileAction,
        closeFileAction,
        visitFile,
        downloadFile,
        openLinkConfirm,
        closeLinkConfirm,
        visitLink,
    };
}
