package release

import (
	"strings"
	"testing"
	"time"
)

func TestParseCommitDetectsConventionalCommit(t *testing.T) {
	commit := ParseCommit("abcdef123", "feat(auth)!: add passkey login", "BREAKING CHANGE: accounts must re-authenticate")

	if commit.Type != "feat" {
		t.Fatalf("Type = %q, want feat", commit.Type)
	}
	if commit.Scope != "auth" {
		t.Fatalf("Scope = %q, want auth", commit.Scope)
	}
	if !commit.Breaking {
		t.Fatal("Breaking = false, want true")
	}
	if commit.Description != "add passkey login" {
		t.Fatalf("Description = %q", commit.Description)
	}
}

func TestGenerateChangelogGroupsUserFacingChanges(t *testing.T) {
	commits := []Commit{
		ParseCommit("111111111", "feat(profile): add avatar crop", ""),
		ParseCommit("222222222", "fix(sync): retry failed uploads", ""),
		ParseCommit("333333333", "chore: update tooling", ""),
	}

	notes := GenerateChangelog("1.2.0", time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC), commits, ChangelogOptions{})

	for _, want := range []string{
		"## 1.2.0 - 2026-05-31",
		"### Features",
		"- **profile:** add avatar crop (1111111)",
		"### Bug Fixes",
		"- **sync:** retry failed uploads (2222222)",
	} {
		if !strings.Contains(notes, want) {
			t.Fatalf("changelog missing %q:\n%s", want, notes)
		}
	}
	if strings.Contains(notes, "update tooling") {
		t.Fatalf("internal commit leaked into public changelog:\n%s", notes)
	}
}
