import React, { useCallback, useEffect, useReducer, useState } from 'react';
import api from '../api/axios';
import packageJson from '../../package.json';
import { check } from '@tauri-apps/plugin-updater';
import { relaunch } from '@tauri-apps/plugin-process';
import UpdaterStatusPanel from './update/UpdaterStatusPanel';
import {
  INITIAL_UPDATER_PROGRESS_STATE,
  updaterProgressReducer,
} from './update/updaterProgress';

interface SystemStatus {
    maintenance: boolean;
    min_version: string;
}

const APP_VERSION = packageJson.version;

export default function MaintenanceGuard({ children }: { children: React.ReactNode }) {
    const [status, setStatus] = useState<SystemStatus | null>(null);
    const [loading, setLoading] = useState(true);
    const [updaterState, dispatchUpdater] = useReducer(
        updaterProgressReducer,
        INITIAL_UPDATER_PROGRESS_STATE,
    );

    const isUpdating =
        updaterState.phase === 'checking' ||
        updaterState.phase === 'downloading' ||
        updaterState.phase === 'installing' ||
        updaterState.phase === 'restarting';

    useEffect(() => {
        const checkStatus = async () => {
            try {
                const res = await api.get('/api/status');
                setStatus(res.data);
            } catch {
                // Silently fail to let the app continue if the backend is down
            } finally {
                setLoading(false);
            }
        };

        checkStatus();
        const interval = setInterval(checkStatus, 60000); // Check every minute
        return () => clearInterval(interval);
    }, []);

    const startUpdate = useCallback(async () => {
        dispatchUpdater({ type: 'checking' });

        try {
            const update = await check();

            if (!update?.available) {
                dispatchUpdater({
                    type: 'error',
                    message: 'You are already using the latest version!',
                });
                return;
            }

            await update.downloadAndInstall((event) => {
                switch (event.event) {
                    case 'Started':
                        dispatchUpdater({
                            type: 'download-started',
                            contentLength: event.data.contentLength,
                        });
                        break;
                    case 'Progress':
                        dispatchUpdater({
                            type: 'download-progress',
                            chunkLength: event.data.chunkLength,
                        });
                        break;
                    case 'Finished':
                        dispatchUpdater({ type: 'download-finished' });
                        break;
                }
            });

            dispatchUpdater({ type: 'restarting' });
            await new Promise((resolve) => window.setTimeout(resolve, 250));
            await relaunch();
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : JSON.stringify(error);

            dispatchUpdater({
                type: 'error',
                message: `Update failed: ${errorMessage}`,
            });
        }
    }, []);

    if (loading) return (
        <div className="fixed inset-0 bg-gray-50 flex items-center justify-center z-[99999]">
            <div className="w-8 h-8 border-2 border-indigo-500/20 border-t-indigo-500 rounded-full animate-spin" />
        </div>
    );

    if (!status) return <>{children}</>;

    const isVersionTooOld = (current: string, min: string) => {
        const c = current.split('.').map(Number);
        const m = min.split('.').map(Number);
        for (let i = 0; i < 3; i++) {
            if (c[i] < m[i]) return true;
            if (c[i] > m[i]) return false;
        }
        return false;
    };

    const needsUpdate = isVersionTooOld(APP_VERSION, status.min_version);

    if (status.maintenance || needsUpdate) {
        return (
            <div className="fixed inset-0 z-[99999] min-h-screen flex items-center justify-center bg-gray-50 px-4 font-sans select-none">
                <div className="max-w-md w-full bg-white rounded-2xl border border-gray-200 p-8 text-center shadow-xl">
                    <div className="mx-auto mb-6 flex justify-center">
                        <img src="/logo.png" alt="SMSystem Logo" className="w-20 h-20 object-contain drop-shadow-sm" />
                    </div>
                    <h1 className="text-2xl font-bold text-gray-900 mb-2">
                        {status.maintenance ? "System Maintenance" : "Update Required"}
                    </h1>
                    <p className="text-gray-500 mb-8 leading-relaxed text-sm">
                        {status.maintenance
                            ? "We are currently optimizing the SMSystem database. Access is temporarily suspended to ensure data integrity."
                            : `A mandatory update (v${status.min_version}) is required to continue using the system and prevent data conflicts.`}
                    </p>

                    <div className="space-y-3">
                        {status.maintenance ? (
                            <div className="p-4 bg-gray-50 rounded-xl text-sm text-gray-600 border border-gray-100 italic">
                                "Please check back shortly or contact the administrator."
                            </div>
                        ) : (
                            <>
                                {(updaterState.phase === 'idle' || updaterState.phase === 'error') && (
                                    <button
                                        onClick={startUpdate}
                                        disabled={isUpdating}
                                        className="w-full py-3 px-4 rounded-xl text-sm font-bold transition-colors shadow-lg uppercase tracking-widest bg-gray-900 text-white hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed"
                                    >
                                        DOWNLOAD NEW VERSION
                                    </button>
                                )}

                                {updaterState.phase !== 'idle' && (
                                    <UpdaterStatusPanel state={updaterState} />
                                )}

                            </>
                        )}
                    </div>

                    <div className="mt-8 text-[10px] font-bold text-gray-400 uppercase tracking-[0.2em]">
                        SMSystem Control v{APP_VERSION}
                    </div>

                    {/* Developer Bypass (Only visible during npm run tauri dev) */}
                    {import.meta.env.DEV && (
                        <button
                            onClick={() => {
                                setStatus(null); // Bypass the guard
                            }}
                            className="mt-6 w-full py-2 px-4 bg-red-100/50 text-red-600 border border-red-200 rounded-xl text-xs font-bold hover:bg-red-100 transition-colors uppercase tracking-wider"
                        >
                            Dev Preview
                        </button>
                    )}
                </div>
            </div>
        );
    }

    return <>{children}</>;
}
