import React, { act } from 'react';
import { afterEach, describe, expect, it } from 'vitest';
import { createRoot, type Root } from 'react-dom/client';
import UpdaterStatusPanel from '../../../components/update/UpdaterStatusPanel';
import type { UpdaterProgressState } from '../../../components/update/updaterProgress';

describe('UpdaterStatusPanel', () => {
  let root: Root | null = null;

  afterEach(() => {
    if (root) {
      act(() => root?.unmount());
      root = null;
    }

    document.body.innerHTML = '';
  });

  async function renderPanel(state: UpdaterProgressState) {
    document.body.innerHTML = '<div id="root"></div>';
    const container = document.getElementById('root');

    if (!container) {
      throw new Error('Missing root container');
    }

    root = createRoot(container);

    await act(async () => {
      root?.render(<UpdaterStatusPanel state={state} />);
      await Promise.resolve();
    });
  }

  it('renders download progress with percent and bytes', async () => {
    await renderPanel({
      phase: 'downloading',
      downloadedBytes: 256,
      totalBytes: 1024,
      percent: 25,
      errorMessage: null,
    });

    expect(document.body.textContent).toContain('Downloading update...');
    expect(document.body.textContent).toContain('25%');
    expect(document.body.textContent).toContain('256 B / 1.0 KB');
  });

  it('renders success copy before relaunch', async () => {
    await renderPanel({
      phase: 'restarting',
      downloadedBytes: 1024,
      totalBytes: 1024,
      percent: 100,
      errorMessage: null,
    });

    expect(document.body.textContent).toContain('Update installed successfully. Restarting...');
    expect(document.body.textContent).toContain('100%');
  });

  it('renders updater errors as alert content', async () => {
    await renderPanel({
      phase: 'error',
      downloadedBytes: 0,
      totalBytes: null,
      percent: null,
      errorMessage: 'Update failed: Network timeout',
    });

    const alert = document.querySelector('[role="alert"]');
    expect(alert?.textContent).toContain('Update failed: Network timeout');
  });
});