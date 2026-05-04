import React, { act } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createRoot, type Root } from 'react-dom/client';

const { mockApiGet, mockCheck, mockRelaunch } = vi.hoisted(() => ({
  mockApiGet: vi.fn(),
  mockCheck: vi.fn(),
  mockRelaunch: vi.fn(),
}));

vi.mock('../../api/axios', () => ({
  default: {
    get: mockApiGet,
  },
}));

vi.mock('@tauri-apps/plugin-updater', () => ({
  check: mockCheck,
}));

vi.mock('@tauri-apps/plugin-process', () => ({
  relaunch: mockRelaunch,
}));

import MaintenanceGuard from '../../components/MaintenanceGuard';

describe('MaintenanceGuard updater flow', () => {
  let root: Root | null = null;

  beforeEach(() => {
    vi.useFakeTimers();
    vi.clearAllMocks();
    document.body.innerHTML = '<div id="root"></div>';
    mockApiGet.mockResolvedValue({
      data: {
        maintenance: false,
        min_version: '5.0.99',
      },
    });
  });

  afterEach(() => {
    if (root) {
      act(() => root?.unmount());
      root = null;
    }

    vi.useRealTimers();
    document.body.innerHTML = '';
  });

  async function renderGuard() {
    const container = document.getElementById('root');

    if (!container) {
      throw new Error('Missing root container');
    }

    root = createRoot(container);

    await act(async () => {
      root?.render(
        <MaintenanceGuard>
          <div>App Content</div>
        </MaintenanceGuard>,
      );
      await Promise.resolve();
      await Promise.resolve();
    });
  }

  it('renders only the in-app updater action when an update is required', async () => {
    await renderGuard();

    const updateButton = Array.from(document.querySelectorAll('button')).find((button) =>
      button.textContent?.includes('DOWNLOAD NEW VERSION'),
    );

    expect(updateButton).toBeTruthy();
    expect(document.body.textContent).not.toContain('Manual Download');
  });

  it('shows progress updates and relaunches after a successful install', async () => {
    let emitEvent:
      | ((event: {
          event: 'Started' | 'Progress' | 'Finished';
          data: { contentLength?: number; chunkLength?: number };
        }) => void)
      | null = null;
    let resolveInstall = () => {};

    const installPromise = new Promise<void>((resolve) => {
      resolveInstall = resolve;
    });

    const downloadAndInstall = vi.fn((onEvent) => {
      emitEvent = onEvent;
      onEvent({ event: 'Started', data: { contentLength: 1024 } });
      onEvent({ event: 'Progress', data: { chunkLength: 256 } });
      return installPromise;
    });

    mockCheck.mockResolvedValue({
      available: true,
      version: '5.0.29',
      downloadAndInstall,
    });

    await renderGuard();

    const updateButton = Array.from(document.querySelectorAll('button')).find((button) =>
      button.textContent?.includes('DOWNLOAD NEW VERSION'),
    ) as HTMLButtonElement;

    await act(async () => {
      updateButton.click();
      await Promise.resolve();
    });

    expect(document.body.textContent).toContain('Downloading update...');
    expect(document.body.textContent).toContain('25%');
    expect(document.body.textContent).toContain('256 B / 1.0 KB');

    await act(async () => {
      emitEvent?.({ event: 'Finished', data: {} });
      await Promise.resolve();
    });

    expect(document.body.textContent).toContain('Installing update...');

    await act(async () => {
      resolveInstall();
      await Promise.resolve();
    });

    expect(document.body.textContent).toContain('Update installed successfully. Restarting...');

    await act(async () => {
      vi.advanceTimersByTime(250);
      await Promise.resolve();
    });

    expect(mockRelaunch).toHaveBeenCalledTimes(1);
  });
});
