package release

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Artifact struct {
	Path      string `json:"path"`
	FileName  string `json:"fileName"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"sizeBytes"`
}

func NewFileArtifact(path string) (Artifact, error) {
	hash, size, err := FileSHA256(path)
	if err != nil {
		return Artifact{}, err
	}
	return Artifact{
		Path:      path,
		FileName:  filepath.Base(path),
		SHA256:    hash,
		SizeBytes: size,
	}, nil
}

func FileSHA256(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", 0, err
	}
	if info.IsDir() {
		return "", 0, fmt.Errorf("%s is a directory", path)
	}

	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(digest.Sum(nil)), info.Size(), nil
}
