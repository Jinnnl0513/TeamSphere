export async function copyText(text: string): Promise<void> {
    if (navigator?.clipboard?.writeText && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
        return;
    }

    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.setAttribute('readonly', 'true');
    textarea.style.position = 'fixed';
    textarea.style.top = '0';
    textarea.style.left = '0';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);
    textarea.focus();
    textarea.select();
    
    try {
        const ok = document.execCommand('copy');
        if (!ok) {
            throw new Error('copy failed');
        }
    } finally {
        document.body.removeChild(textarea);
    }
}
