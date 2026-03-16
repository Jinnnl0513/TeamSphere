import apiClient from '../../api/client';

export type EmailRequiredResponse = {
  email_required?: boolean;
  email_enabled?: boolean;
  allow_register?: boolean;
};

export type LoginResponse = {
  token: string;
  refresh_token: string;
};

export type VerifyEmailResponse = {
  verification_token: string;
};

export type OAuthProvider = {
  name: string;
  label: string;
  enabled: boolean;
};

export type OAuthProvidersResponse = {
  providers: OAuthProvider[];
};

export type TOTPStatusResponse = {
  enabled: boolean;
  policy?: string;
  required?: boolean;
};

export type TOTPSetupResponse = {
  secret: string;
  otpauth_url: string;
};

export type RecoveryCodesResponse = {
  recovery_codes: string[];
};

export type SetupRequiredResponse = {
  setup_token: string;
};

export type RecoveryCodesStatusResponse = {
  remaining: number;
};

export type SessionsResponse = {
  sessions: Array<{
    id: number;
    created_at: string;
    last_used_at?: string;
    expires_at: string;
    revoked_at?: string;
    ip_address?: string;
    user_agent?: string;
    device_name?: string;
    is_current: boolean;
  }>;
};

export const authApi = {
  getEmailRequired: () => apiClient.get<EmailRequiredResponse>('/auth/email-required'),
  getOAuthProviders: () => apiClient.get<OAuthProvidersResponse>('/auth/oauth/providers'),
  getTOTPStatus: () => apiClient.get<TOTPStatusResponse>('/auth/2fa/status'),
  setupTOTP: () => apiClient.post<TOTPSetupResponse>('/auth/2fa/setup'),
  setupTOTPRequired: (payload: { setup_token: string }) =>
    apiClient.post<TOTPSetupResponse>('/auth/2fa/setup-required', payload),
  enableTOTP: (payload: { code: string }) =>
    apiClient.post<RecoveryCodesResponse>('/auth/2fa/enable', payload),
  enableTOTPRequired: (payload: { setup_token: string; code: string }) =>
    apiClient.post<RecoveryCodesResponse & LoginResponse & { user?: any }>('/auth/2fa/enable-required', payload),
  disableTOTP: (payload: { code: string }) => apiClient.post('/auth/2fa/disable', payload),
  recoveryCodesStatus: () => apiClient.get<RecoveryCodesStatusResponse>('/auth/2fa/recovery-codes/status'),
  regenRecoveryCodes: (payload: { code: string }) =>
    apiClient.post<RecoveryCodesResponse>('/auth/2fa/recovery-codes/regen', payload),
  verifyTwoFALogin: (payload: { challenge: string; code?: string; recovery_code?: string }) =>
    apiClient.post<LoginResponse>('/auth/2fa/verify-login', payload),
  listSessions: (refreshToken?: string) =>
    apiClient.get<SessionsResponse>('/auth/sessions', refreshToken ? { headers: { 'X-Refresh-Token': refreshToken } } : undefined),
  revokeSession: (payload: { session_id: number }) => apiClient.post('/auth/sessions/revoke', payload),
  revokeOtherSessions: (payload: { refresh_token: string }) =>
    apiClient.post('/auth/sessions/revoke-others', payload),
  login: (payload: { username: string; password: string; totp_code?: string; recovery_code?: string }) =>
    apiClient.post<LoginResponse>('/auth/login', payload),
  register: (payload: {
    username: string;
    password: string;
    email?: string;
    verification_token?: string;
  }) => apiClient.post('/auth/register', payload),
  sendCode: (payload: { email: string }) => apiClient.post('/auth/send-code', payload),
  verifyEmail: (payload: { email: string; code: string }) =>
    apiClient.post<VerifyEmailResponse>('/auth/verify-email', payload),
  sendPasswordResetCode: (payload: { email: string }) =>
    apiClient.post('/auth/password/reset-code', payload),
  resetPassword: (payload: { email: string; code: string; new_password: string }) =>
    apiClient.post('/auth/password/reset', payload),
  logout: (payload?: { refresh_token?: string }) =>
    apiClient.post('/auth/logout', payload ?? {}),
};
