import { isTauriRuntime } from '../utils/runtime';

const REMOTE_API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://168.144.46.137:8080';

export function getApiBaseUrl(): string {
  if (isTauriRuntime()) {
    return REMOTE_API_BASE;
  }

  return import.meta.env.DEV ? '' : REMOTE_API_BASE;
}

export function buildApiUrl(url: string): string {
  if (url.startsWith('http')) {
    return url;
  }

  return `${getApiBaseUrl()}${url}`;
}
