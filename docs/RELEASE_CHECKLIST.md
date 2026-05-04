# SMSytem Desktop Release Checklist

> Working checklist for Tauri desktop releases with stable in-app updates.

## Why this exists

The desktop updater now uses a stable manifest URL:

`https://github.com/Jayem09/SMSystem/releases/download/latest/latest.json`

This checklist keeps releases aligned with that flow so new versions publish correctly and the in-app updater keeps working.

## 1) Bump the version in all desktop release files

Update the same version in:

- `frontend/package.json`
- `frontend/src-tauri/tauri.conf.json`
- `frontend/src-tauri/Cargo.toml`

Keep the Tauri updater endpoint set to:

```json
"https://github.com/Jayem09/SMSystem/releases/download/latest/latest.json"
```

## 2) Commit and tag the release

Example for `5.0.35`:

```bash
git add frontend/package.json frontend/src-tauri/tauri.conf.json frontend/src-tauri/Cargo.toml
git commit -m "Bump version to 5.0.35"
git tag v5.0.35
git push origin main v5.0.35
```

## 3) Verify GitHub Actions publish flow

Check `.github/workflows/publish.yml` run for the version tag.

Confirm the versioned release contains:

- `latest.json`
- macOS artifacts (`.dmg`, `.app.tar.gz`, `.sig`)
- Windows artifacts (`.exe`, `.sig`)

## 4) Verify the stable updater manifest moved forward

After the versioned release finishes, confirm:

- `https://github.com/Jayem09/SMSystem/releases/tag/latest` exists
- `https://github.com/Jayem09/SMSystem/releases/download/latest/latest.json` shows the new version

If `latest/latest.json` still shows the old version, wait for the `update-latest-release` job to finish before testing the updater.

## 5) Test the in-app updater

To force the update popup for version `5.0.35`:

```sql
UPDATE settings SET value='5.0.35' WHERE `key` = 'min_app_version';
```

Then:

1. Open the older installed app (for example `5.0.34`)
2. Restart the app after changing `min_app_version`
3. Confirm the update popup appears
4. Click **Download Update**
5. Verify progress bar, percentage, bytes downloaded, install success, and restart flow

## 6) macOS workaround if the app is blocked

If macOS says the app is damaged:

```bash
sudo xattr -rd com.apple.quarantine /Applications/SMSytem.app
```

If needed, also try opening the app via Finder with **Right click → Open**.

## 7) Quick failure checks

- Popup appears but app says **already latest** → check `latest/latest.json` version
- New release exists but updater does not move → confirm the installed app is a lower version than the release you are testing
- macOS app will not open → remove quarantine attribute
- Version mismatch confusion → re-check all three version files before tagging

## Related files

- `.github/workflows/publish.yml` - versioned publish + stable latest release flow
- `frontend/src-tauri/tauri.conf.json` - stable updater endpoint
- `AGENTS.md` - repo structure and release-related commands
