import axios from 'axios';
import type { AxiosRequestConfig, AxiosError } from 'axios';
import { API_BASE_URL } from '../config/app';

const apiClient = axios.create({
    baseURL: API_BASE_URL,
    timeout: 10000,
});

let isRefreshing = false;
let failedQueue: Array<{
    resolve: (value: unknown) => void;
    reject: (reason: unknown) => void;
}> = [];

const processQueue = (error: AxiosError | null, token: string | null) => {
    failedQueue.forEach(prom => {
        if (error) {
            prom.reject(error);
        } else {
            prom.resolve(token);
        }
    });
    failedQueue = [];
};

apiClient.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
}, (error) => {
    return Promise.reject(error);
});

apiClient.interceptors.response.use((response) => {
    if (response.data && typeof response.data === 'object') {
        if (response.data.code === 0) {
            return response.data.data;
        }
        if (typeof response.data.code === 'number' && response.data.code !== 0) {
            return Promise.reject(new ApiError(response.data.message || 'Request failed', response.data.code, response.data.data));
        }
    }
    return response.data;
}, async (error) => {
    const originalRequest = error.config;
    if (!originalRequest) return Promise.reject(error);

    if (error.response?.status === 401 && !originalRequest._retry) {
        const requestUrl = (originalRequest.url || '').toString();
        const isAuthEndpoint = requestUrl.includes('/auth/login') || requestUrl.includes('/auth/2fa');

        const refreshToken = localStorage.getItem('refresh_token');
        
        if (error.response?.data?.code === 40102 && refreshToken) {
            if (isRefreshing) {
                return new Promise((resolve, reject) => {
                    failedQueue.push({ resolve, reject });
                }).then(token => {
                    originalRequest.headers.Authorization = `Bearer ${token}`;
                    return apiClient(originalRequest);
                }).catch(err => {
                    return Promise.reject(err);
                });
            }

            originalRequest._retry = true;
            isRefreshing = true;

            try {
                const response = await axios.post(API_BASE_URL + '/auth/refresh', {
                    refresh_token: refreshToken,
                });

                const { token, refresh_token } = response.data.data || response.data;
                
                localStorage.setItem('token', token);
                if (refresh_token) {
                    localStorage.setItem('refresh_token', refresh_token);
                }
                
                processQueue(null, token);
                
                originalRequest.headers.Authorization = `Bearer ${token}`;
                return apiClient(originalRequest);
            } catch (refreshError) {
                processQueue(refreshError as AxiosError, null);
                localStorage.removeItem('token');
                localStorage.removeItem('refresh_token');
                window.location.href = '/login';
                return Promise.reject(refreshError);
            } finally {
                isRefreshing = false;
            }
        }

        if (!isAuthEndpoint) {
            localStorage.removeItem('token');
            localStorage.removeItem('refresh_token');
            window.location.href = '/login';
        }
    }

    const serverMsg = error.response?.data?.message;
    if (serverMsg) {
        return Promise.reject(new ApiError(serverMsg, error.response.data.code ?? -1, error.response.data?.data));
    }
    return Promise.reject(error);
});

export class ApiError extends Error {
    public readonly code: number;
    public readonly data?: unknown;
    constructor(message: string, code: number, data?: unknown) {
        super(message);
        this.code = code;
        this.data = data;
        this.name = 'ApiError';
    }
}

export type ApiClientType = {
    get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T>;
    post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>;
    put<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>;
    delete<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T>;
};

export default apiClient as unknown as ApiClientType;

