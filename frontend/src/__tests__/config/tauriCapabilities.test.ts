import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

type CapabilityConfig = {
  permissions?: string[];
};

const readDefaultCapability = (): CapabilityConfig => {
  const filePath = resolve(import.meta.dirname, '../../../src-tauri/capabilities/default.json');
  return JSON.parse(readFileSync(filePath, 'utf8')) as CapabilityConfig;
};

describe('default Tauri capability', () => {
  it('allows process restart when updater flow is enabled', () => {
    const capability = readDefaultCapability();
    const permissions = capability.permissions ?? [];

    expect(permissions).toContain('updater:default');
    expect(permissions).toContain('process:default');
  });
});
