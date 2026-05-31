package release

import (
	"fmt"
	"strconv"
	"time"
)

const ManifestSchemaVersion = 1

type Platform string

const (
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
)

type Manifest struct {
	SchemaVersion int        `json:"schemaVersion"`
	Platform      Platform   `json:"platform"`
	Version       string     `json:"version"`
	Build         string     `json:"build"`
	Artifacts     []Artifact `json:"artifacts"`
	Changelog     string     `json:"changelog,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

func NewManifest(platformInput string, version Version, build string, artifactPaths []string, changelog string, createdAt time.Time) (Manifest, error) {
	platform, err := ParsePlatform(platformInput)
	if err != nil {
		return Manifest{}, err
	}
	if build == "" {
		return Manifest{}, fmt.Errorf("build is required")
	}
	if platform == PlatformAndroid {
		if _, err := strconv.Atoi(build); err != nil {
			return Manifest{}, fmt.Errorf("android build must be an integer versionCode: %w", err)
		}
	}
	if len(artifactPaths) == 0 {
		return Manifest{}, fmt.Errorf("at least one artifact is required")
	}

	artifacts := make([]Artifact, 0, len(artifactPaths))
	for _, path := range artifactPaths {
		artifact, err := NewFileArtifact(path)
		if err != nil {
			return Manifest{}, err
		}
		artifacts = append(artifacts, artifact)
	}

	return Manifest{
		SchemaVersion: ManifestSchemaVersion,
		Platform:      platform,
		Version:       version.String(),
		Build:         build,
		Artifacts:     artifacts,
		Changelog:     changelog,
		CreatedAt:     createdAt,
	}, nil
}

func ParsePlatform(input string) (Platform, error) {
	switch Platform(input) {
	case PlatformAndroid:
		return PlatformAndroid, nil
	case PlatformIOS:
		return PlatformIOS, nil
	default:
		return "", fmt.Errorf("unsupported platform %q", input)
	}
}
