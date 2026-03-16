import apiClient from '../../api/client';

export type SetupStatus = {
    needed: boolean;
    db_configured: boolean;
};

export type TestDbPayload = {
    host: string;
    port: number;
    user: string;
    password: string;
    dbname: string;
};

export type TestEmailPayload = {
    smtp_host: string;
    smtp_port: number;
    username: string;
    password: string;
    from_address: string;
    from_name: string;
    to: string;
};

export type TestConnectionPayload = {
    db: TestDbPayload;
    redis_enabled?: boolean;
    redis?: {
        host: string;
        port: number;
        password: string;
        db: number;
    };
};

export type SetupPayload = {
    db?: TestDbPayload;
    email_enabled?: boolean;
    email?: Omit<TestEmailPayload, 'to'>;
    redis_enabled?: boolean;
    redis?: {
        host: string;
        port: number;
        password: string;
        db: number;
    };
    admin_username: string;
    admin_password: string;
    admin_email: string;
};

export type SetupResponse = {
    token: string;
    refresh_token?: string;
};

export const setupApi = {
    getStatus: () => apiClient.get<SetupStatus>('/setup/status'),
    testDb: (payload: TestDbPayload) => apiClient.post('/setup/test-db', payload),
    testConnection: (payload: TestConnectionPayload) => apiClient.post('/setup/test-connection', payload),
    testEmail: (payload: TestEmailPayload) => apiClient.post('/setup/test-email', payload),
    setup: (payload: SetupPayload) => apiClient.post<SetupResponse>('/setup', payload),
};
