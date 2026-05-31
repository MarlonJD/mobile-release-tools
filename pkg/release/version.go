package release

import (
	"fmt"
	"regexp"
	"strconv"
)

var versionPattern = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
var shortVersionPattern = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

type BumpLevel string

const (
	BumpMajor BumpLevel = "major"
	BumpMinor BumpLevel = "minor"
	BumpPatch BumpLevel = "patch"
)

type Version struct {
	Major         int
	Minor         int
	Patch         int
	Prerelease    string
	BuildMetadata string
}

func ParseVersion(input string) (Version, error) {
	matches := versionPattern.FindStringSubmatch(input)
	if matches == nil {
		return Version{}, fmt.Errorf("invalid semantic version %q", input)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	return Version{
		Major:         major,
		Minor:         minor,
		Patch:         patch,
		Prerelease:    matches[4],
		BuildMetadata: matches[5],
	}, nil
}

func ParsePlatformVersion(input string) (Version, error) {
	version, err := ParseVersion(input)
	if err == nil {
		return version, nil
	}
	matches := shortVersionPattern.FindStringSubmatch(input)
	if matches == nil {
		return Version{}, err
	}
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	return Version{Major: major, Minor: minor}, nil
}

func (version Version) String() string {
	value := fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)
	if version.Prerelease != "" {
		value += "-" + version.Prerelease
	}
	if version.BuildMetadata != "" {
		value += "+" + version.BuildMetadata
	}
	return value
}

func (version Version) Bump(level BumpLevel) (Version, error) {
	next := Version{
		Major: version.Major,
		Minor: version.Minor,
		Patch: version.Patch,
	}

	switch level {
	case BumpMajor:
		next.Major++
		next.Minor = 0
		next.Patch = 0
	case BumpMinor:
		next.Minor++
		next.Patch = 0
	case BumpPatch:
		next.Patch++
	default:
		return Version{}, fmt.Errorf("unsupported bump level %q", level)
	}
	return next, nil
}
