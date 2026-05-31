package release

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManifestValidatesAndroidBuildAndIncludesArtifact(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.aab")
	if err := os.WriteFile(path, []byte("bundle"), 0o644); err != nil {
		t.Fatal(err)
	}
	version, err := ParseVersion("2.0.0")
	if err != nil {
		t.Fatal(err)
	}

	manifest, err := NewManifest("android", version, "42", []string{path}, "notes", time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	if manifest.Platform != PlatformAndroid {
		t.Fatalf("Platform = %q, want android", manifest.Platform)
	}
	if manifest.Version != "2.0.0" {
		t.Fatalf("Version = %q, want 2.0.0", manifest.Version)
	}
	if len(manifest.Artifacts) != 1 {
		t.Fatalf("Artifacts length = %d, want 1", len(manifest.Artifacts))
	}
}

func TestNewManifestRejectsNonIntegerAndroidBuild(t *testing.T) {
	version, err := ParseVersion("2.0.0")
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewManifest("android", version, "1.2.3", []string{"missing.aab"}, "", time.Now())
	if err == nil {
		t.Fatal("expected error for non-integer Android build")
	}
}
