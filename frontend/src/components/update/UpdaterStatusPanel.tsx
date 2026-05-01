import { AlertCircle, CheckCircle2, Loader2 } from 'lucide-react';
import {
  formatBytes,
  getVisiblePercent,
  type UpdaterProgressState,
} from './updaterProgress';

type UpdaterStatusPanelProps = {
  state: UpdaterProgressState;
};

const STATUS_COPY = {
  checking: 'Checking for update...',
  downloading: 'Downloading update...',
  installing: 'Installing update...',
  restarting: 'Update installed successfully. Restarting...',
} satisfies Record<'checking' | 'downloading' | 'installing' | 'restarting', string>;

export default function UpdaterStatusPanel({
  state,
}: UpdaterStatusPanelProps) {
  if (state.phase === 'idle') {
    return null;
  }

  if (state.phase === 'error') {
    return (
      <div
        role="alert"
        className="p-3 bg-red-50 rounded-xl text-xs text-red-600 border border-red-200"
      >
        <div className="flex items-start gap-2">
          <AlertCircle className="w-4 h-4 mt-0.5 shrink-0" />
          <span>{state.errorMessage}</span>
        </div>
      </div>
    );
  }

  const percent = getVisiblePercent(state);
  const isSuccess = state.phase === 'restarting';
  const byteText =
    state.totalBytes === null
      ? `${formatBytes(state.downloadedBytes)} downloaded`
      : `${formatBytes(state.downloadedBytes)} / ${formatBytes(state.totalBytes)}`;

  return (
    <div
      className={`rounded-xl border p-4 text-xs ${
        isSuccess
          ? 'bg-emerald-50 border-emerald-200 text-emerald-700'
          : 'bg-gray-50 border-gray-200 text-gray-600'
      }`}
    >
      <div className="flex items-center gap-2 text-sm font-semibold">
        {isSuccess ? (
          <CheckCircle2 className="w-4 h-4 shrink-0" />
        ) : (
          <Loader2 className="w-4 h-4 shrink-0 animate-spin" />
        )}
        <span>{STATUS_COPY[state.phase]}</span>
      </div>

      {(state.phase === 'downloading' ||
        state.phase === 'installing' ||
        state.phase === 'restarting') && (
        <>
          <div
            className="mt-3 h-2 overflow-hidden rounded-full bg-gray-200"
            role="progressbar"
            aria-valuemin={0}
            aria-valuemax={100}
            aria-valuenow={percent ?? undefined}
          >
            <div
              className={`h-full rounded-full transition-all duration-300 ${
                isSuccess ? 'bg-emerald-500' : 'bg-gray-900'
              } ${percent === null ? 'w-1/2 animate-pulse' : ''}`}
              style={percent === null ? undefined : { width: `${percent}%` }}
            />
          </div>

          <div className="mt-2 flex items-center justify-between">
            <span>{percent === null ? 'Preparing size…' : `${percent}%`}</span>
            <span>
              {state.phase === 'downloading' ? byteText : 'Package verified'}
            </span>
          </div>
        </>
      )}
    </div>
  );
}