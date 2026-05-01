import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const readFile = (relativePath: string) =>
  readFileSync(resolve(import.meta.dirname, relativePath), 'utf8');

describe('Tauri process plugin wiring', () => {
  it('registers the process plugin when the updater uses relaunch', () => {
    const cargoToml = readFile('../../../src-tauri/Cargo.toml');
    const libRs = readFile('../../../src-tauri/src/lib.rs');

    expect(cargoToml).toContain('tauri-plugin-process = "2"');
    expect(libRs).toContain('.plugin(tauri_plugin_process::init())');
  });
});
