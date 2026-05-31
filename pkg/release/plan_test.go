package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInferBumpLevelUsesConventionalCommits(t *testing.T) {
	tests := []struct {
		name    string
		commits []Commit
		want    BumpLevel
	}{
		{
			name: "breaking",
			commits: []Commit{
				ParseCommit("1", "fix!: remove legacy login", ""),
			},
			want: BumpMajor,
		},
		{
			name: "feature",
			commits: []Commit{
				ParseCommit("1", "fix: correct typo", ""),
				ParseCommit("2", "feat: add release automation", ""),
			},
			want: BumpMinor,
		},
		{
			name: "patch",
			commits: []Commit{
				ParseCommit("1", "perf: speed up hashing", ""),
			},
			want: BumpPatch,
		},
		{
			name: "internal defaults to patch when packaging is requested",
			commits: []Commit{
				ParseCommit("1", "docs: update readme", ""),
			},
			want: BumpPatch,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, _ := InferBumpLevel(test.commits)
			if got != test.want {
				t.Fatalf("InferBumpLevel() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNextBuildNumberUsesPreviousBuildMetadata(t *testing.T) {
	if got := nextBuildNumber("104"); got != "105" {
		t.Fatalf("nextBuildNumber() = %q, want 105", got)
	}
}

func TestPlanNextReleaseUsesPlatformCurrentVersionInsteadOfLatestTag(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.email", "release@example.com")
	runGit(t, repo, "config", "user.name", "Release Test")
	writeTestFile(t, filepath.Join(repo, "README.md"), "initial\n")
	runGit(t, repo, "add", "README.md")
	runGit(t, repo, "commit", "-m", "chore: initial release")
	runGit(t, repo, "tag", "v9.9.9")
	writeTestFile(t, filepath.Join(repo, "README.md"), "initial\nfeature\n")
	runGit(t, repo, "add", "README.md")
	runGit(t, repo, "commit", "-m", "feat: add mobile release")

	current, err := ParseVersion("1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := PlanNextRelease(ReleasePlanOptions{
		RepoPath:       repo,
		CurrentVersion: current,
		CurrentBuild:   "41",
	})
	if err != nil {
		t.Fatal(err)
	}

	if plan.CurrentVersion.String() != "1.2.3" {
		t.Fatalf("CurrentVersion = %q, want 1.2.3", plan.CurrentVersion.String())
	}
	if plan.Version.String() != "1.3.0" {
		t.Fatalf("Version = %q, want 1.3.0", plan.Version.String())
	}
	if plan.Build != "42" {
		t.Fatalf("Build = %q, want 42", plan.Build)
	}
	if plan.PreviousRef != "v9.9.9" {
		t.Fatalf("PreviousRef = %q, want v9.9.9", plan.PreviousRef)
	}
}

func TestParseVersionFromTagUsesPrefix(t *testing.T) {
	version, err := ParseVersionFromTag("ios/v2.3.4+88", "ios/")
	if err != nil {
		t.Fatal(err)
	}
	if version.String() != "2.3.4+88" {
		t.Fatalf("version = %q, want 2.3.4+88", version.String())
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
