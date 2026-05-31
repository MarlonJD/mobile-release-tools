# Installation

`mobile-release` is distributed as a single CLI binary and as native package
manager metadata for the common desktop/server environments used by release
owners.

The first tagged source release was `v0.1.0`; it only contains GitHub's
automatic source archives. Use `v0.1.1` or newer for platform archives,
checksums, `.deb`, and `.rpm` release assets.

## Supported Install Channels

### macOS

Homebrew Cask is the preferred binary install path for tags published by the
release workflow:

```bash
brew install --cask marlonjd/tap/mobile-release
```

The initial `v0.1.0` source formula can still be installed with:

```bash
brew install marlonjd/tap/mobile-release
```

### Windows: Scoop

```powershell
scoop bucket add marlonjd https://github.com/MarlonJD/scoop-bucket
scoop install marlonjd/mobile-release
```

The Scoop manifest is generated from GitHub Release checksums and committed to
`MarlonJD/scoop-bucket` by the release workflow.

### Windows: WinGet

```powershell
winget install MarlonJD.MobileRelease
```

WinGet publishing is slower than Scoop because it goes through the public
`microsoft/winget-pkgs` review flow. The release workflow generates the WinGet
manifest and opens a pull request from `MarlonJD/winget-pkgs` to
`microsoft/winget-pkgs`.

Until the WinGet PR for a version is merged, use Scoop or the direct Windows
archive from GitHub Releases.

### Linux: Debian/Ubuntu

Download the `.deb` package from the matching GitHub Release, then install it:

```bash
sudo apt install ./mobile-release*.deb
```

### Linux: Fedora/RHEL

Download the `.rpm` package from the matching GitHub Release, then install it:

```bash
sudo dnf install ./mobile-release*.rpm
```

### Direct GitHub Releases

Every tag publishes direct archives for the supported OS/architecture matrix:

```text
mobile-release_<version>_darwin_amd64.tar.gz
mobile-release_<version>_darwin_arm64.tar.gz
mobile-release_<version>_linux_amd64.tar.gz
mobile-release_<version>_linux_arm64.tar.gz
mobile-release_<version>_windows_amd64.zip
mobile-release_<version>_windows_arm64.zip
checksums.txt
```

Manual install from a direct archive:

```bash
tar -xzf mobile-release_<version>_<os>_<arch>.tar.gz
sudo install -m 0755 mobile-release /usr/local/bin/mobile-release
```

For Windows, unzip the matching `.zip` archive and put `mobile-release.exe` on
`PATH`.

### Go Source Fallback

Use the Go path only when package-manager or archive install is not available:

```bash
go install github.com/MarlonJD/mobile-release-tools/cmd/mobile-release@v0.2.2
```

Make sure Go's binary directory is on `PATH`.

Linux/macOS:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Windows PowerShell:

```powershell
$env:Path += ";$(go env GOPATH)\bin"
```

## Verify

```bash
mobile-release --help
```

The command should print support for:

- `bump`
- `changelog`
- `hash`
- `manifest`
- `package android`
- `package ios`

## Maintainer Release Prerequisites

Before pushing a release tag, these repositories must exist:

- `MarlonJD/homebrew-tap`
- `MarlonJD/scoop-bucket`
- `MarlonJD/winget-pkgs`, forked from `microsoft/winget-pkgs`

The `mobile-release-tools` repository must also define these GitHub Actions
secrets:

- `RELEASE_PUBLISH_TOKEN`: personal access token with write access to
  `MarlonJD/homebrew-tap` and `MarlonJD/scoop-bucket`.
- `WINGET_TOKEN`: personal access token with write access to
  `MarlonJD/winget-pkgs`. If omitted, the workflow reuses
  `RELEASE_PUBLISH_TOKEN`.

## Maintainer Release Flow

1. Run tests locally:

   ```bash
   go test ./...
   ```

2. Create and push a SemVer tag:

   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```

3. Let `.github/workflows/release.yml` run GoReleaser.

The default tag workflow publishes:

- GitHub Release archives for macOS, Linux, and Windows.
- `checksums.txt`.
- Linux `.deb` packages.
- Linux `.rpm` packages.

Package-manager metadata publishing is a separate manual workflow dispatch
until the required repositories and secrets are present. Run the release
workflow manually with `publish_package_managers=true` after creating:

- Homebrew Cask update target in `MarlonJD/homebrew-tap`.
- Scoop manifest target in `MarlonJD/scoop-bucket`.
- WinGet manifest fork/PR target in `MarlonJD/winget-pkgs`.

## Notes

Snap can be added later if the tool needs Snap Store discovery. It is not the
first release channel because it adds store metadata, review, and sandbox
maintenance that are not needed for a local release CLI.
