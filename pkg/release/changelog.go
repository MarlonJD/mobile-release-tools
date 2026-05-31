package release

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var conventionalSubjectPattern = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9-]*)(?:\(([^)]+)\))?(!)?:\s+(.+)$`)

// Commit is the normalized release-relevant shape of a git commit message.
type Commit struct {
	Hash        string
	Type        string
	Scope       string
	Description string
	Body        string
	Breaking    bool
	RawSubject  string
}

type ChangelogOptions struct {
	IncludeInternal bool
}

func ParseCommit(hash, subject, body string) Commit {
	subject = strings.TrimSpace(subject)
	body = strings.TrimSpace(body)
	commit := Commit{
		Hash:        strings.TrimSpace(hash),
		Type:        "other",
		Description: subject,
		Body:        body,
		RawSubject:  subject,
		Breaking:    hasBreakingFooter(body),
	}

	matches := conventionalSubjectPattern.FindStringSubmatch(subject)
	if matches == nil {
		return commit
	}

	commit.Type = strings.ToLower(matches[1])
	commit.Scope = matches[2]
	commit.Breaking = commit.Breaking || matches[3] == "!"
	commit.Description = matches[4]
	return commit
}

func GenerateChangelog(version string, date time.Time, commits []Commit, options ChangelogOptions) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("## %s - %s\n\n", version, date.Format("2006-01-02")))

	sections := []struct {
		title string
		match func(Commit) bool
	}{
		{"Breaking Changes", func(commit Commit) bool { return commit.Breaking }},
		{"Features", func(commit Commit) bool { return !commit.Breaking && commit.Type == "feat" }},
		{"Bug Fixes", func(commit Commit) bool { return !commit.Breaking && commit.Type == "fix" }},
		{"Performance", func(commit Commit) bool { return !commit.Breaking && commit.Type == "perf" }},
		{"Security", func(commit Commit) bool { return !commit.Breaking && commit.Type == "security" }},
		{"Other Changes", func(commit Commit) bool {
			return !commit.Breaking && options.IncludeInternal && !isPrimaryType(commit.Type)
		}},
	}

	wroteAny := false
	for _, section := range sections {
		items := matchingCommits(commits, section.match, options)
		if len(items) == 0 {
			continue
		}

		wroteAny = true
		builder.WriteString("### ")
		builder.WriteString(section.title)
		builder.WriteString("\n\n")
		for _, item := range items {
			builder.WriteString("- ")
			if item.Scope != "" {
				builder.WriteString("**")
				builder.WriteString(item.Scope)
				builder.WriteString(":** ")
			}
			builder.WriteString(item.Description)
			if item.Hash != "" {
				builder.WriteString(" (")
				builder.WriteString(shortHash(item.Hash))
				builder.WriteString(")")
			}
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	if !wroteAny {
		builder.WriteString("No notable changes.\n")
	}
	return builder.String()
}

func matchingCommits(commits []Commit, match func(Commit) bool, options ChangelogOptions) []Commit {
	items := make([]Commit, 0)
	for _, commit := range commits {
		if !options.IncludeInternal && isInternalType(commit.Type) {
			continue
		}
		if match(commit) {
			items = append(items, commit)
		}
	}
	return items
}

func hasBreakingFooter(body string) bool {
	return strings.Contains(body, "BREAKING CHANGE:") || strings.Contains(body, "BREAKING-CHANGE:")
}

func isPrimaryType(commitType string) bool {
	switch commitType {
	case "feat", "fix", "perf", "security":
		return true
	default:
		return false
	}
}

func isInternalType(commitType string) bool {
	switch commitType {
	case "build", "chore", "ci", "docs", "refactor", "style", "test":
		return true
	default:
		return false
	}
}

func shortHash(hash string) string {
	if len(hash) <= 7 {
		return hash
	}
	return hash[:7]
}
