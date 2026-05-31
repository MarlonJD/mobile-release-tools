package release

import "testing"

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
	version, err := ParseVersion("1.2.3+104")
	if err != nil {
		t.Fatal(err)
	}

	if got := nextBuildNumber(version); got != "105" {
		t.Fatalf("nextBuildNumber() = %q, want 105", got)
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
