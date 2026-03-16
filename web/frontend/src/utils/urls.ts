import { FILE_BASE_URL } from '../config/app';

export function resolveFileUrl(url: string): string {
    if (!url) return url;
    if (url.startsWith('http://') || url.startsWith('https://')) return url;
    if (url.startsWith('/uploads/')) {
        return `${FILE_BASE_URL}${url}`;
    }
    return url;
}

export function isUploadsPath(url: string): boolean {
    return url.startsWith('/uploads/');
}
