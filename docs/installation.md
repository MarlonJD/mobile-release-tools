# Installation

`mobile-release` is currently distributed through a Homebrew source formula and
a pinned Go source install. Binary archives, Scoop, WinGet, `.deb`, and `.rpm`
packages are planned follow-up channels.

## Available Now

### macOS

Preferred install path:

```bash
brew install marlonjd/tap/mobile-release
```

Verify:

```bash
mobile-release --help
```

### Linux

Use Homebrew/Linuxbrew when available:

```bash
brew install marlonjd/tap/mobile-release
```

Or install from source with Go:

```bash
go install github.com/MarlonJD/mobile-release-tools/cmd/mobile-release@v0.1.0
```

Make sure Go's binary directory is on `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

### Windows

Use the Go install path until the Scoop/WinGet packages are published:

```powershell
go install github.com/MarlonJD/mobile-release-tools/cmd/mobile-release@v0.1.0
```

Make sure Go's binary directory is on `PATH`:

```powershell
$env:Path += ";$(go env GOPATH)\bin"
```

Verify:

```powershell
mobile-release --help
```

## Planned Channels

These channels should be added after GitHub release binaries are published:

- Windows: Scoop bucket.
- Windows: WinGet package.
- Linux: `.deb` package.
- Linux: `.rpm` package.
- Direct: GitHub Releases archives for each supported OS/architecture.

Snap can be added later if the tool needs Snap Store discovery. It is not the
first release channel because it adds store metadata, review, and sandbox
maintenance that are not needed for a local release CLI.

## Homebrew Formula

The Homebrew formula lives in `MarlonJD/homebrew-tap`:

```text
Formula/mobile-release.rb
```

It builds from the tagged source archive:

```text
https://github.com/MarlonJD/mobile-release-tools/archive/refs/tags/v0.1.0.tar.gz
```

## Source Install Version Pin

Use explicit tags instead of `latest` when installing into a release machine:

```bash
go install github.com/MarlonJD/mobile-release-tools/cmd/mobile-release@v0.1.0
```

## Maintainer Release Flow

1. Run tests locally:

   ```bash
   go test ./...
   ```

2. Create a SemVer tag:

   ```bash
   git tag v0.1.1
   git push origin v0.1.1
   ```

3. Update `MarlonJD/homebrew-tap/Formula/mobile-release.rb` to the new tag and
   source tarball SHA-256.

4. Run the tap checks:

   ```bash
   ruby -c Formula/mobile-release.rb
   brew style Formula/mobile-release.rb
   ```
