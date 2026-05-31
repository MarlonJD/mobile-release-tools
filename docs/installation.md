# Installation

`mobile-release` is distributed from GitHub Releases. Package managers should all
resolve to the same signed release artifacts and SHA-256 checksums.

## macOS

Preferred install path:

```bash
brew install marlonjd/tap/mobile-release
```

The Homebrew formula lives in `MarlonJD/homebrew-tap` and builds from the
tagged source archive.

Direct install is also supported by downloading the matching
`mobile-release_<version>_darwin_<arch>.tar.gz` archive from GitHub Releases and
placing `mobile-release` on `PATH`.

## Windows

Preferred install paths:

```powershell
scoop bucket add marlonjd https://github.com/MarlonJD/scoop-bucket
scoop install mobile-release
```

```powershell
winget install MarlonJD.MobileRelease
```

Scoop is the easiest first-party channel because GoReleaser can update the
bucket from the same release command. WinGet should be added after the first
stable release by submitting the package manifest to `microsoft/winget-pkgs`.

Direct install is also supported by downloading the matching
`mobile-release_<version>_windows_<arch>.zip` archive from GitHub Releases and
placing `mobile-release.exe` on `PATH`.

## Linux

Preferred install paths:

```bash
brew install marlonjd/tap/mobile-release
```

```bash
sudo dpkg -i mobile-release_<version>_linux_<arch>.deb
```

```bash
sudo rpm -i mobile-release_<version>_linux_<arch>.rpm
```

Linuxbrew gives one command that matches macOS. Debian and RPM packages are
generated for environments where native Linux package managers are preferred.

Snap can be added later if the tool needs Snap Store discovery. It is not the
first release channel because it adds store metadata, review, and sandbox
maintenance that are not needed for a local release CLI.

## Release Flow

1. Run tests locally:

   ```bash
   go test ./...
   ```

2. Create a SemVer tag:

   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

3. Publish artifacts locally or in CI with GoReleaser:

   ```bash
   goreleaser release --clean
   ```

The first release needs credentials for GitHub Releases and optional tap/bucket
repositories. Local snapshot builds can be generated without publishing:

```bash
goreleaser release --snapshot --clean
```
