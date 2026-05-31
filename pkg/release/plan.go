package release

import "strconv"

type ReleasePlanOptions struct {
	RepoPath        string
	FromRef         string
	ToRef           string
	TagPrefix       string
	CurrentVersion  Version
	CurrentBuild    string
	VersionSource   string
	BuildSource     string
	VersionOverride string
	BuildOverride   string
}

type ReleasePlan struct {
	PreviousRef       string
	CurrentVersion    Version
	CurrentBuild      string
	Version           Version
	Build             string
	ReleaseID         string
	BumpLevel         BumpLevel
	BumpReason        string
	Commits           []Commit
	VersionSource     string
	BuildSource       string
	UsedVersionFlag   bool
	UsedBuildFlag     bool
	NoPreviousRelease bool
}

func PlanNextRelease(options ReleasePlanOptions) (ReleasePlan, error) {
	if options.RepoPath == "" {
		options.RepoPath = "."
	}
	if options.ToRef == "" {
		options.ToRef = "HEAD"
	}
	if options.TagPrefix == "" {
		options.TagPrefix = "v"
	}

	current := options.CurrentVersion
	var previousRef string
	noPrevious := false
	if options.FromRef != "" {
		previousRef = options.FromRef
	} else {
		tag, _, ok, err := LatestVersionTag(options.RepoPath, options.TagPrefix)
		if err != nil {
			return ReleasePlan{}, err
		}
		if ok {
			previousRef = tag
		} else {
			noPrevious = true
		}
	}

	commits, err := LoadGitCommits(options.RepoPath, previousRef, options.ToRef)
	if err != nil {
		return ReleasePlan{}, err
	}

	level, reason := InferBumpLevel(commits)
	next, err := current.Bump(level)
	if err != nil {
		return ReleasePlan{}, err
	}

	usedVersionFlag := false
	if options.VersionOverride != "" {
		next, err = ParseVersion(options.VersionOverride)
		if err != nil {
			return ReleasePlan{}, err
		}
		usedVersionFlag = true
	}

	build := options.BuildOverride
	usedBuildFlag := build != ""
	if build == "" {
		build = nextBuildNumber(options.CurrentBuild)
		if build == "" {
			if current.BuildMetadata != "" {
				build = nextBuildNumber(current.BuildMetadata)
			}
			if build == "" {
				count, err := GitCommitCount(options.RepoPath, options.ToRef)
				if err != nil {
					return ReleasePlan{}, err
				}
				build = strconv.Itoa(count)
			}
		}
	}

	return ReleasePlan{
		PreviousRef:       previousRef,
		CurrentVersion:    current,
		CurrentBuild:      options.CurrentBuild,
		Version:           next,
		Build:             build,
		ReleaseID:         next.String() + "+" + build,
		BumpLevel:         level,
		BumpReason:        reason,
		Commits:           commits,
		VersionSource:     options.VersionSource,
		BuildSource:       options.BuildSource,
		UsedVersionFlag:   usedVersionFlag,
		UsedBuildFlag:     usedBuildFlag,
		NoPreviousRelease: noPrevious,
	}, nil
}

func InferBumpLevel(commits []Commit) (BumpLevel, string) {
	hasPatch := false
	for _, commit := range commits {
		if commit.Breaking {
			return BumpMajor, "breaking change commit"
		}
		if commit.Type == "feat" {
			return BumpMinor, "feature commit"
		}
		switch commit.Type {
		case "fix", "perf", "security":
			hasPatch = true
		}
	}
	if hasPatch {
		return BumpPatch, "patch-level commit"
	}
	return BumpPatch, "no release-relevant commits; defaulting to patch because packaging was requested"
}

func ParseVersionFromTag(tag, tagPrefix string) (Version, error) {
	if tagPrefix != "" && len(tag) >= len(tagPrefix) && tag[:len(tagPrefix)] == tagPrefix {
		tag = tag[len(tagPrefix):]
	}
	return ParseVersion(tag)
}

func nextBuildNumber(current string) string {
	if current == "" {
		return ""
	}
	value, err := strconv.Atoi(current)
	if err != nil {
		return ""
	}
	if value < 0 {
		return ""
	}
	return strconv.Itoa(value + 1)
}
