# Mobile Release Tools

`mobile-release-tools` is a local-first release CLI for iOS and Android shipping.
It keeps version calculation, changelog generation, artifact hashing, and release
manifest creation in one portable Go binary.

The CLI does not replace platform signing or store delivery:

- iOS archive/export still runs through `xcodebuild` on macOS with Xcode
  signing/provisioning.
- Android bundle generation still runs through Gradle and the Android keystore.
- App Store Connect and Google Play delivery remain explicit publish steps.

## Why Go

Go is the right fit for this repository because the tool is a cross-platform
release orchestrator rather than an app runtime dependency. A single static
binary can run locally or in CI, while platform packaging commands are delegated
to native toolchains.

## Commands

```bash
mobile-release bump --current 1.4.2 --level minor
mobile-release changelog --repo . --from v1.4.2 --to HEAD --version 1.5.0 --output RELEASE_NOTES.md
mobile-release hash --file build/app-release.aab
mobile-release manifest --platform android --version 1.5.0 --build 105 --artifact build/app-release.aab --notes RELEASE_NOTES.md --output release-manifest.json
mobile-release mobile package android --version 1.5.0 --build 105 --channel production
mobile-release mobile package ios --version 1.5.0 --build 105 --export-options apps/ios/release/ExportOptions-app-store.plist --allow-provisioning-updates
```

The mobile package commands invoke the native platform toolchains:

- Android uses Gradle `:app:bundleRelease` and the app's release signing
  configuration to produce a signed `.aab`.
- iOS uses `xcodebuild archive` and `xcodebuild -exportArchive` with an export
  options plist to produce a signed `.ipa`.

Android signing defaults to `--signing env`, which requires:

- `EMSI_ANDROID_RELEASE_STORE_FILE`
- `EMSI_ANDROID_RELEASE_STORE_PASSWORD`
- `EMSI_ANDROID_RELEASE_KEY_ALIAS`
- `EMSI_ANDROID_RELEASE_KEY_PASSWORD`

Use `--signing external` when Gradle properties or CI secrets provide signing.
Use `--signing unsigned` only for local QA artifacts that will not be uploaded
to Google Play.

## Installation

See [docs/installation.md](docs/installation.md) for package manager setup.

Planned first-party channels:

- macOS: Homebrew tap
- Windows: Scoop bucket first, WinGet after the first stable release
- Linux: Homebrew/Linuxbrew, `.deb`, and `.rpm`
- Direct: GitHub Releases archives for every supported OS/architecture

## Changelog Policy

Release notes are generated from Conventional Commit subjects between the last
released tag and the target ref.

User-facing sections are generated for:

- `feat`: Features
- `fix`: Bug Fixes
- `perf`: Performance
- `security`: Security
- `!` or `BREAKING CHANGE`: Breaking Changes

Internal-only commit types such as `chore`, `ci`, `build`, `test`, `docs`,
`style`, and `refactor` are skipped by default. Use `--include-internal` when an
internal release log is needed.

## Version Policy

The CLI follows Semantic Versioning for the shared marketing version and keeps
platform build counters separate:

- iOS: `CFBundleShortVersionString` is SemVer; `CFBundleVersion` is a monotonic
  build number.
- Android: `versionName` is SemVer; `versionCode` is a monotonic integer.

CI is optional. The same commands are designed to run locally first, then in CI
as a repeatable verification path when budget and approval are available.

## License

Copyright (C) 2026 Burak Karahan.

This project is licensed under the GNU General Public License v3.0 or later.
See [LICENSE](LICENSE) and [NOTICE](NOTICE).
