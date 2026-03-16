import { useCallback, useState } from 'react';
import toast from 'react-hot-toast';
import { ApiError } from '../api/client';

export function getApiErrorMessage(err: unknown, fallback = '请求失败') {
    if (!err) return fallback;
    if (err instanceof ApiError) return err.message || fallback;
    if (err instanceof Error) return err.message || fallback;
    return fallback;
}

export function useApiCall<T>() {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const call = useCallback(async (fn: () => Promise<T>, opts?: { toastError?: boolean; errorMessage?: string }) => {
        setLoading(true);
        setError(null);
        try {
            const result = await fn();
            return result;
        } catch (err) {
            const msg = getApiErrorMessage(err, opts?.errorMessage);
            setError(msg);
            if (opts?.toastError !== false) {
                toast.error(msg);
            }
            throw err;
        } finally {
            setLoading(false);
        }
    }, []);

    return { loading, error, call };
}
