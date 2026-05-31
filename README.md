# Mobile Release Tools

`mobile-release-tools` provides `mobile-release`, a local-first release CLI for
iOS and Android applications. It exists to make mobile shipping repeatable
without forcing the project to depend on paid CI, a hosted release service, or
manual release-note/version bookkeeping.

The CLI keeps the release mechanics that should be shared across platforms in
one portable Go binary:

- Semantic Versioning bump calculation.
- Conventional Commits changelog generation.
- Artifact SHA-256 hashing.
- Release manifest generation.
- Android `.aab` packaging through Gradle.
- iOS `.ipa` packaging through Xcode archive/export.

It does not replace the official platform toolchains:

- Android signing is still handled by Gradle and the configured upload key.
- iOS signing and provisioning are still handled by Xcode.
- Google Play and App Store Connect upload/review/release remain explicit
  release-owner steps.

## Why This Project Exists

Mobile releases usually fail in boring, expensive ways: reused build numbers,
missing changelog notes, unsigned or incorrectly signed artifacts, unclear
artifact paths, and local steps that differ from CI steps. This project keeps
those decisions in a small CLI so the release process is the same whether it is
run on a developer machine or later moved into CI.

The intended model is:

1. Prepare release metadata locally.
2. Build the signed platform artifact locally.
3. Hash and summarize the produced artifact.
4. Upload manually to App Store Connect or Google Play.
5. Optionally run the same commands in CI later.

## Why Go

Go is used because this is a cross-platform release orchestrator, not an app
runtime dependency. The CLI can be shipped as one binary for macOS, Linux, and
Windows, while platform-specific packaging stays delegated to Xcode and Gradle.

## Install

Until a tagged release is published, install from source:

```bash
go install github.com/MarlonJD/mobile-release-tools/cmd/mobile-release@latest
```

Or clone and run directly:

```bash
git clone https://github.com/MarlonJD/mobile-release-tools.git
cd mobile-release-tools
go test ./...
go run ./cmd/mobile-release --help
```

Planned package-manager channels are documented in
[docs/installation.md](docs/installation.md):

- macOS: Homebrew tap
- Windows: Scoop first, WinGet after the first stable release
- Linux: Homebrew/Linuxbrew, `.deb`, and `.rpm`
- Direct: GitHub Releases archives for every supported OS/architecture

## Expected App Repository Setup

`mobile-release` is normally run from the app monorepo root. The current default
paths match this layout:

```text
apps/
  android/
    gradlew
    app/build.gradle.kts
  ios/
    emsi_ios.xcodeproj
    release/ExportOptions-app-store.plist
```

You can override all important paths with command flags, so the CLI can be used
outside this exact layout.

## iOS Setup From Scratch

Requirements:

- macOS with Xcode installed.
- An Apple Developer account available to Xcode.
- A valid distribution certificate/provisioning setup for the app bundle ID.
- An Xcode project and shared scheme.
- An export options plist committed in the app repository.

Create the export options file:

```bash
mkdir -p apps/ios/release
```

`apps/ios/release/ExportOptions-app-store.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "https://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>destination</key>
	<string>export</string>
	<key>manageAppVersionAndBuildNumber</key>
	<false/>
	<key>method</key>
	<string>app-store-connect</string>
	<key>signingStyle</key>
	<string>automatic</string>
	<key>stripSwiftSymbols</key>
	<true/>
	<key>teamID</key>
	<string>YOUR_TEAM_ID</string>
	<key>uploadSymbols</key>
	<true/>
</dict>
</plist>
```

Replace `YOUR_TEAM_ID` with the Apple Developer Team ID used by the app target.

Package a signed `.ipa`:

```bash
mobile-release mobile package ios \
  --project apps/ios/emsi_ios.xcodeproj \
  --scheme emsi_ios \
  --version 1.5.0 \
  --build 105 \
  --export-options apps/ios/release/ExportOptions-app-store.plist \
  --allow-provisioning-updates
```

What this runs:

```bash
xcodebuild test ...
xcodebuild archive ...
xcodebuild -exportArchive ...
```

Default output:

```text
build/releases/ios/1.5.0+105/
  emsi_ios.xcarchive
  export/*.ipa
```

Use `--dry-run` to print the commands without running Xcode.

## Android Setup From Scratch

Requirements:

- JDK and Android SDK installed.
- A Gradle wrapper in the Android project.
- An Android app module, usually `:app`.
- Release signing configured through environment variables or Gradle
  properties.
- A Gradle build that accepts:
  - `emsi.versionName`
  - `emsi.versionCode`
  - `emsi.distributionChannel`

For environment-based signing, set:

```bash
export EMSI_ANDROID_RELEASE_STORE_FILE=/absolute/path/to/upload-key.jks
export EMSI_ANDROID_RELEASE_STORE_PASSWORD=...
export EMSI_ANDROID_RELEASE_KEY_ALIAS=...
export EMSI_ANDROID_RELEASE_KEY_PASSWORD=...
```

Package a signed `.aab`:

```bash
mobile-release mobile package android \
  --project apps/android \
  --version 1.5.0 \
  --build 105 \
  --channel production
```

What this runs:

```bash
./gradlew :app:testDebugUnitTest
./gradlew :app:bundleRelease \
  -Pemsi.versionName=1.5.0 \
  -Pemsi.versionCode=105 \
  -Pemsi.distributionChannel=production
```

Default Gradle output:

```text
apps/android/app/build/outputs/bundle/release/app-release.aab
```

Signing modes:

- `--signing env`: default; requires all `EMSI_ANDROID_RELEASE_*` variables.
- `--signing external`: use when Gradle properties or CI secrets provide
  signing outside environment variables.
- `--signing unsigned`: use only for local QA artifacts that will not be
  uploaded to Google Play.

Use `--include-apk` to also run `:app:assembleRelease` for QA builds. Use
`--dry-run` to print the commands without running Gradle.

## Release Flow

1. Decide the next version:

   ```bash
   mobile-release bump --current 1.4.2 --level minor
   ```

2. Generate release notes from Conventional Commits:

   ```bash
   mobile-release changelog \
     --repo . \
     --from v1.4.2 \
     --to HEAD \
     --version 1.5.0 \
     --output RELEASE_NOTES.md
   ```

3. Build the platform artifact:

   ```bash
   mobile-release mobile package android --version 1.5.0 --build 105 --channel production
   ```

   ```bash
   mobile-release mobile package ios --version 1.5.0 --build 105 --export-options apps/ios/release/ExportOptions-app-store.plist --allow-provisioning-updates
   ```

4. Hash the artifact:

   ```bash
   mobile-release hash --file apps/android/app/build/outputs/bundle/release/app-release.aab
   ```

   ```bash
   mobile-release hash --file build/releases/ios/1.5.0+105/export/App.ipa
   ```

5. Write a release manifest:

   ```bash
   mobile-release manifest \
     --platform android \
     --version 1.5.0 \
     --build 105 \
     --artifact apps/android/app/build/outputs/bundle/release/app-release.aab \
     --notes RELEASE_NOTES.md \
     --output release-manifest.android.json
   ```

6. Upload manually:

   - Android `.aab` to Google Play Console.
   - iOS `.ipa` to App Store Connect through Xcode Organizer, Transporter, or
     `xcrun altool`/Transporter tooling.

## Command Reference

```bash
mobile-release bump --current 1.4.2 --level patch|minor|major
mobile-release changelog --repo . --from v1.4.2 --to HEAD --version 1.4.3 --output RELEASE_NOTES.md
mobile-release hash --file path/to/artifact
mobile-release manifest --platform ios|android --version 1.4.3 --build 104 --artifact path --notes RELEASE_NOTES.md --output manifest.json
mobile-release mobile package android --version 1.4.3 --build 104 --channel production
mobile-release mobile package ios --version 1.4.3 --build 104 --export-options apps/ios/release/ExportOptions-app-store.plist
```

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

Example release id:

```text
1.5.0+105
```

## Troubleshooting

Android signing error:

```text
android signing env is incomplete
```

Set the `EMSI_ANDROID_RELEASE_*` variables, use `--signing external` if Gradle
properties provide signing, or use `--signing unsigned` only for QA artifacts.

iOS export fails because signing assets are missing:

- Confirm the app target has the correct Team ID and bundle ID.
- Sign into Xcode with the Apple Developer account.
- Pass `--allow-provisioning-updates` if automatic signing should update
  provisioning assets.

The app version changed after export:

- Keep `manageAppVersionAndBuildNumber` set to `false` in the export options
  plist.
- Pass `--version` and `--build` explicitly to `mobile-release mobile package
  ios`.

## License

Copyright (C) 2026 Burak Karahan.

This project is licensed under the GNU General Public License v3.0 or later.
See [LICENSE](LICENSE) and [NOTICE](NOTICE).
