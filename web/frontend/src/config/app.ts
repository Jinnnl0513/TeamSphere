const isDev = import.meta.env.DEV;

const trustedDomains = (import.meta.env.VITE_TRUSTED_DOMAINS || '')
  .split(',')
  .map((d: string) => d.trim().toLowerCase())
  .filter(Boolean);

export const API_BASE_URL = isDev ? 'http://localhost:8080/api/v1' : '/api/v1';

export const FILE_BASE_URL = API_BASE_URL.replace(/\/api\/v1\/?$/, '');

export const WS_BASE_URL = isDev
  ? 'ws://localhost:8080/api/v1/ws'
  : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/ws`;

export const APP_NAME = 'TeamSphere';

export const TRUSTED_DOMAINS = trustedDomains;
