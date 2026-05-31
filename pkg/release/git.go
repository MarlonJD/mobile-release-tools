package release

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	gitFieldSeparator  = "\x1f"
	gitRecordSeparator = "\x1e"
)

func LoadGitCommits(repoPath, fromRef, toRef string) ([]Commit, error) {
	if repoPath == "" {
		repoPath = "."
	}
	if toRef == "" {
		toRef = "HEAD"
	}

	revision := toRef
	if fromRef != "" {
		revision = fromRef + ".." + toRef
	}

	args := []string{
		"-C", repoPath,
		"log",
		"--no-merges",
		"--format=%H%x1f%s%x1f%b%x1e",
		revision,
	}
	output, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("load git commits: %w", err)
	}

	records := strings.Split(string(output), gitRecordSeparator)
	commits := make([]Commit, 0, len(records))
	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}
		fields := strings.SplitN(record, gitFieldSeparator, 3)
		if len(fields) < 2 {
			continue
		}
		body := ""
		if len(fields) == 3 {
			body = fields[2]
		}
		commits = append(commits, ParseCommit(fields[0], fields[1], body))
	}
	return commits, nil
}

func LatestVersionTag(repoPath, tagPrefix string) (string, Version, bool, error) {
	if repoPath == "" {
		repoPath = "."
	}
	output, err := exec.Command("git", "-C", repoPath, "tag", "--list", "--sort=-v:refname").Output()
	if err != nil {
		return "", Version{}, false, fmt.Errorf("load git tags: %w", err)
	}

	for _, line := range strings.Split(string(output), "\n") {
		tag := strings.TrimSpace(line)
		if tag == "" {
			continue
		}
		if tagPrefix != "" && !strings.HasPrefix(tag, tagPrefix) {
			continue
		}
		versionInput := tag
		if tagPrefix != "" {
			versionInput = strings.TrimPrefix(tag, tagPrefix)
		}
		version, err := ParseVersion(versionInput)
		if err != nil {
			continue
		}
		return tag, version, true, nil
	}
	return "", Version{}, false, nil
}

func GitCommitCount(repoPath, ref string) (int, error) {
	if repoPath == "" {
		repoPath = "."
	}
	if ref == "" {
		ref = "HEAD"
	}
	output, err := exec.Command("git", "-C", repoPath, "rev-list", "--count", ref).Output()
	if err != nil {
		return 0, fmt.Errorf("count git commits: %w", err)
	}
	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("parse git commit count: %w", err)
	}
	return count, nil
}
