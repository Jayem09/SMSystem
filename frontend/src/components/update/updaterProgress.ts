export type UpdaterPhase =
  | 'idle'
  | 'checking'
  | 'downloading'
  | 'installing'
  | 'restarting'
  | 'error';

export type UpdaterProgressState = {
  phase: UpdaterPhase;
  downloadedBytes: number;
  totalBytes: number | null;
  percent: number | null;
  errorMessage: string | null;
};

export type UpdaterProgressAction =
  | { type: 'reset' }
  | { type: 'checking' }
  | { type: 'download-started'; contentLength?: number | null }
  | { type: 'download-progress'; chunkLength: number }
  | { type: 'download-finished' }
  | { type: 'restarting' }
  | { type: 'error'; message: string };

export const INITIAL_UPDATER_PROGRESS_STATE: UpdaterProgressState = {
  phase: 'idle',
  downloadedBytes: 0,
  totalBytes: null,
  percent: null,
  errorMessage: null,
};

export function formatBytes(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`;
  }

  const units = ['KB', 'MB', 'GB'];
  let value = bytes / 1024;
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }

  return `${value.toFixed(1)} ${units[unitIndex]}`;
}

export function calculatePercent(
  downloadedBytes: number,
  totalBytes: number | null,
): number | null {
  if (!totalBytes || totalBytes <= 0) {
    return null;
  }

  return Math.min(100, Math.round((downloadedBytes / totalBytes) * 100));
}

export function getVisiblePercent(state: UpdaterProgressState): number | null {
  if (state.phase === 'installing' || state.phase === 'restarting') {
    return 100;
  }

  return state.percent;
}

export function updaterProgressReducer(
  state: UpdaterProgressState,
  action: UpdaterProgressAction,
): UpdaterProgressState {
  switch (action.type) {
    case 'reset':
      return INITIAL_UPDATER_PROGRESS_STATE;
    case 'checking':
      return {
        ...INITIAL_UPDATER_PROGRESS_STATE,
        phase: 'checking',
      };
    case 'download-started': {
      const totalBytes = action.contentLength && action.contentLength > 0
        ? action.contentLength
        : null;

      return {
        phase: 'downloading',
        downloadedBytes: 0,
        totalBytes,
        percent: calculatePercent(0, totalBytes),
        errorMessage: null,
      };
    }
    case 'download-progress': {
      const downloadedBytes = state.downloadedBytes + action.chunkLength;

      return {
        ...state,
        phase: 'downloading',
        downloadedBytes,
        percent: calculatePercent(downloadedBytes, state.totalBytes),
      };
    }
    case 'download-finished':
      return {
        ...state,
        phase: 'installing',
        downloadedBytes: state.totalBytes ?? state.downloadedBytes,
        percent: 100,
      };
    case 'restarting':
      return {
        ...state,
        phase: 'restarting',
        percent: 100,
        errorMessage: null,
      };
    case 'error':
      return {
        ...INITIAL_UPDATER_PROGRESS_STATE,
        phase: 'error',
        errorMessage: action.message,
      };
    default:
      return state;
  }
}