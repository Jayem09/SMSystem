import { describe, expect, it } from 'vitest';
import {
  INITIAL_UPDATER_PROGRESS_STATE,
  formatBytes,
  updaterProgressReducer,
} from '../../../components/update/updaterProgress';

describe('updaterProgress helpers', () => {
  it('formats bytes into human-readable units', () => {
    expect(formatBytes(0)).toBe('0 B');
    expect(formatBytes(1024)).toBe('1.0 KB');
    expect(formatBytes(1536)).toBe('1.5 KB');
    expect(formatBytes(1024 * 1024)).toBe('1.0 MB');
  });

  it('tracks download progress when total size is known', () => {
    const started = updaterProgressReducer(INITIAL_UPDATER_PROGRESS_STATE, {
      type: 'download-started',
      contentLength: 1024,
    });

    const progressed = updaterProgressReducer(started, {
      type: 'download-progress',
      chunkLength: 256,
    });

    expect(progressed.phase).toBe('downloading');
    expect(progressed.totalBytes).toBe(1024);
    expect(progressed.downloadedBytes).toBe(256);
    expect(progressed.percent).toBe(25);
  });

  it('keeps percentage empty when the updater does not expose total size', () => {
    const started = updaterProgressReducer(INITIAL_UPDATER_PROGRESS_STATE, {
      type: 'download-started',
      contentLength: undefined,
    });

    const progressed = updaterProgressReducer(started, {
      type: 'download-progress',
      chunkLength: 512,
    });

    expect(progressed.totalBytes).toBeNull();
    expect(progressed.percent).toBeNull();
    expect(progressed.downloadedBytes).toBe(512);
  });

  it('switches to installing when download finishes', () => {
    const downloading = {
      ...INITIAL_UPDATER_PROGRESS_STATE,
      phase: 'downloading' as const,
      downloadedBytes: 1024,
      totalBytes: 1024,
      percent: 100,
    };

    const finished = updaterProgressReducer(downloading, {
      type: 'download-finished',
    });

    expect(finished.phase).toBe('installing');
    expect(finished.percent).toBe(100);
  });
});