export default function ImagePreviewModal({
    src,
    images,
    onChange,
    onClose,
}: {
    src: string;
    images: string[];
    onChange: (nextSrc: string) => void;
    onClose: () => void;
}) {
    const currentIndex = images.findIndex((item) => item === src);
    const hasMultiple = images.length > 1;

    const goPrev = () => {
        if (!hasMultiple) return;
        const nextIndex = currentIndex <= 0 ? images.length - 1 : currentIndex - 1;
        onChange(images[nextIndex]);
    };

    const goNext = () => {
        if (!hasMultiple) return;
        const nextIndex = currentIndex < 0 || currentIndex >= images.length - 1 ? 0 : currentIndex + 1;
        onChange(images[nextIndex]);
    };

    return (
        <div
            className="fixed inset-0 z-[100] flex flex-col items-center justify-center bg-black/85 p-4 sm:p-8 backdrop-blur-sm"
            onClick={onClose}
            onKeyDown={(e) => {
                if (e.key === 'Escape') onClose();
                if (e.key === 'ArrowLeft') goPrev();
                if (e.key === 'ArrowRight') goNext();
            }}
            tabIndex={-1}
        >
            <button
                className="absolute top-4 right-4 md:top-6 md:right-6 text-white/50 hover:text-white transition-colors z-10 bg-black/20 hover:bg-black/40 rounded-full p-2"
                onClick={onClose}
                title="??"
            >
                <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>

            {hasMultiple && (
                <button
                    className="absolute left-4 md:left-8 text-white/70 hover:text-white transition-colors z-10 bg-black/30 hover:bg-black/50 rounded-full p-3"
                    onClick={(e) => {
                        e.stopPropagation();
                        goPrev();
                    }}
                    title="???"
                >
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                    </svg>
                </button>
            )}

            <div className="relative max-w-full max-h-[85vh] flex items-center justify-center">
                <img
                    src={src}
                    alt="Preview"
                    className="max-w-full max-h-full object-contain rounded shadow-2xl"
                    onClick={(e) => e.stopPropagation()}
                />
            </div>

            {hasMultiple && (
                <button
                    className="absolute right-4 md:right-8 text-white/70 hover:text-white transition-colors z-10 bg-black/30 hover:bg-black/50 rounded-full p-3"
                    onClick={(e) => {
                        e.stopPropagation();
                        goNext();
                    }}
                    title="???"
                >
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                </button>
            )}

            <a
                href={src}
                target="_blank"
                rel="noreferrer"
                className="text-sm font-medium text-gray-400 hover:text-white hover:underline mt-6 transition-colors z-10 bg-black/50 px-4 py-2 rounded-full backdrop-blur-sm"
                onClick={(e) => e.stopPropagation()}
            >
                ?????????
            </a>
        </div>
    );
}
