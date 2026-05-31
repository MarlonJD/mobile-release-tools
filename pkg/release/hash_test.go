package release

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileArtifactCalculatesSHA256AndSize(t *testing.T) {
	path := filepath.Join(t.TempDir(), "artifact.txt")
	if err := os.WriteFile(path, []byte("release"), 0o644); err != nil {
		t.Fatal(err)
	}

	artifact, err := NewFileArtifact(path)
	if err != nil {
		t.Fatal(err)
	}

	const wantSHA256 = "a4d451ec23463726f72c43d64c710968f6b602cd653b4de8adee1b556240a829"
	if artifact.SHA256 != wantSHA256 {
		t.Fatalf("SHA256 = %q, want %q", artifact.SHA256, wantSHA256)
	}
	if artifact.SizeBytes != 7 {
		t.Fatalf("SizeBytes = %d, want 7", artifact.SizeBytes)
	}
}
