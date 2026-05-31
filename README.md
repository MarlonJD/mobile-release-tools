# Mobile Release Tools

## Quick Install

macOS with Homebrew:

```bash
brew install --cask marlonjd/tap/mobile-release
```

The initial `v0.1.0` source formula can also be installed with:

```bash
brew install marlonjd/tap/mobile-release
```

Windows with Scoop:

```powershell
scoop bucket add marlonjd https://github.com/MarlonJD/scoop-bucket
scoop install marlonjd/mobile-release
```

Windows with WinGet, after the package is accepted into the public WinGet
community repository:

```powershell
winget install MarlonJD.MobileRelease
```

Linux can install the `.deb` or `.rpm` package from GitHub Releases:

```bash
sudo apt install ./mobile-release*.deb
sudo dnf install ./mobile-release*.rpm
```

Direct archives are published on GitHub Releases for every supported
OS/architecture:

```text
darwin/amd64, darwin/arm64
linux/amd64, linux/arm64
windows/amd64, windows/arm64
```

Any platform with Go installed:

```bash
go install github.com/MarlonJD/mobile-release-tools/cmd/mobile-release@v0.2.1
```

Verify:

```bash
mobile-release --help
```

Full installation and maintainer publishing details are in
[docs/installation.md](docs/installation.md).

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

## Install Details

Homebrew binary install:

```bash
brew install --cask marlonjd/tap/mobile-release
```

Pinned source install with Go:

```bash
go install github.com/MarlonJD/mobile-release-tools/cmd/mobile-release@v0.2.1
```

Or clone and run directly:

```bash
git clone https://github.com/MarlonJD/mobile-release-tools.git
cd mobile-release-tools
go test ./...
go run ./cmd/mobile-release --help
```

Install options and package-manager publishing details are documented in
[docs/installation.md](docs/installation.md).


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
mobile-release package ios \
  --project apps/ios/emsi_ios.xcodeproj \
  --scheme emsi_ios \
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
build/releases/ios/<version>+<build>/
  emsi_ios.xcarchive
  export/*.ipa
  RELEASE_NOTES.md
  release-manifest.json
```

The current version and build are read from the Xcode project file
(`MARKETING_VERSION` and `CURRENT_PROJECT_VERSION`). The next version is
calculated from Conventional Commits, and the next build increments the current
project build. Git tags are used only to find the changelog range. Use
`--dry-run` to print the commands and final upload summary without running
Xcode.

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
mobile-release package android \
  --project apps/android \
  --channel production
```

What this runs:

```bash
./gradlew :app:testDebugUnitTest
./gradlew :app:bundleRelease \
  -Pemsi.versionName=<computed-version> \
  -Pemsi.versionCode=<computed-build> \
  -Pemsi.distributionChannel=production
```

Default Gradle output:

```text
apps/android/app/build/outputs/bundle/release/app-release.aab
build/releases/android/<version>+<build>/RELEASE_NOTES.md
build/releases/android/<version>+<build>/release-manifest.json
```

The current version and build are read from `apps/android/app/build.gradle.kts`
by default. Use `--build-file` when the app keeps `versionName` and
`versionCode` in another Gradle file.

Signing modes:

- `--signing env`: default; requires all `EMSI_ANDROID_RELEASE_*` variables.
- `--signing external`: use when Gradle properties or CI secrets provide
  signing outside environment variables.
- `--signing unsigned`: use only for local QA artifacts that will not be
  uploaded to Google Play.

Use `--include-apk` to also run `:app:assembleRelease` for QA builds. Use
`--dry-run` to print the commands without running Gradle.

## Release Flow

1. Build the platform artifact:

   ```bash
   mobile-release package android --channel production
   ```

   ```bash
   mobile-release package ios --export-options apps/ios/release/ExportOptions-app-store.plist --allow-provisioning-updates
   ```

The package command automatically:

- Reads the current app version/build from the platform project file.
- Finds the latest SemVer git tag, using the `v` prefix by default, only for
  the changelog range.
- Infers the next SemVer bump from Conventional Commits.
- Computes the next platform build number from the current platform build.
- Passes the computed version/build into Gradle or Xcode.
- Writes `RELEASE_NOTES.md`.
- Writes `release-manifest.json` with artifact SHA-256 hashes and sizes.
- Prints the upload destination and artifact path at the end.

2. Upload the artifact listed in the terminal summary:

   - Android `.aab` to Google Play Console.
   - iOS `.ipa` to App Store Connect through Xcode Organizer, Transporter, or
     `xcrun altool`/Transporter tooling.

## Command Reference

```bash
mobile-release bump --current 1.4.2 --level patch|minor|major
mobile-release changelog --repo . --from v1.4.2 --to HEAD --version 1.4.3 --output RELEASE_NOTES.md
mobile-release hash --file path/to/artifact
mobile-release manifest --platform ios|android --version 1.4.3 --build 104 --artifact path --notes RELEASE_NOTES.md --output manifest.json
mobile-release package android --channel production [--build-file apps/android/app/build.gradle.kts]
mobile-release package ios --export-options apps/ios/release/ExportOptions-app-store.plist
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

Automatic package versioning uses these rules:

- Current iOS version/build are read from
  `apps/ios/emsi_ios.xcodeproj/project.pbxproj` by default.
- Current Android version/build are read from `apps/android/app/build.gradle.kts`
  by default.
- Git tags do not define the active app version. They define the previous
  release point for changelog generation.
- `BREAKING CHANGE` or `!`: major bump.
- `feat`: minor bump.
- `fix`, `perf`, or `security`: patch bump.
- Internal-only commits default to patch when packaging is explicitly
  requested.
- The next build is the current platform build plus one.

Use `--version` or `--build` only for an explicit release-owner override.

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
- Use the platform project file as the source of truth for the current version,
  and use the terminal summary from `mobile-release package ios` for the
  computed release candidate. Pass `--version` or `--build` only when
  intentionally overriding the automatic calculation.

## License

Copyright (C) 2026 Burak Karahan.

This project is licensed under the GNU General Public License v3.0 or later.
See [LICENSE](LICENSE) and [NOTICE](NOTICE).
