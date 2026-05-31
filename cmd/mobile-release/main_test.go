package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageIOSUsesShortDefaultCommand(t *testing.T) {
	repo := t.TempDir()
	runTestGit(t, repo, "init")
	runTestGit(t, repo, "config", "user.email", "release@example.com")
	runTestGit(t, repo, "config", "user.name", "Release Test")
	writeTestFile(t, filepath.Join(repo, "README.md"), "initial\n")
	runTestGit(t, repo, "add", "README.md")
	runTestGit(t, repo, "commit", "-m", "feat: first release")

	projectDir := filepath.Join(repo, "apps", "ios", "emsi_ios.xcodeproj")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(projectDir, "project.pbxproj"), `
buildSettings = {
    CURRENT_PROJECT_VERSION = 1;
    MARKETING_VERSION = 1.0;
};
`)

	var stdout bytes.Buffer
	err := run([]string{
		"package", "ios",
		"--repo", repo,
		"--project", projectDir,
		"--skip-tests",
		"--dry-run",
	}, &stdout)
	if err != nil {
		t.Fatal(err)
	}

	output := stdout.String()
	for _, want := range []string{
		"xcodebuild archive",
		"-exportOptionsPlist apps/ios/release/ExportOptions-app-store.plist",
		"-allowProvisioningUpdates",
		"Current version source: " + filepath.Join(projectDir, "project.pbxproj") + ":MARKETING_VERSION",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestPackageIOSCanDisableProvisioningUpdates(t *testing.T) {
	repo := t.TempDir()
	runTestGit(t, repo, "init")
	runTestGit(t, repo, "config", "user.email", "release@example.com")
	runTestGit(t, repo, "config", "user.name", "Release Test")
	writeTestFile(t, filepath.Join(repo, "README.md"), "initial\n")
	runTestGit(t, repo, "add", "README.md")
	runTestGit(t, repo, "commit", "-m", "fix: first patch")

	projectDir := filepath.Join(repo, "apps", "ios", "emsi_ios.xcodeproj")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(projectDir, "project.pbxproj"), `
buildSettings = {
    CURRENT_PROJECT_VERSION = 1;
    MARKETING_VERSION = 1.0;
};
`)

	var stdout bytes.Buffer
	err := run([]string{
		"package", "ios",
		"--repo", repo,
		"--project", projectDir,
		"--skip-tests",
		"--dry-run",
		"--allow-provisioning-updates=false",
	}, &stdout)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdout.String(), "-allowProvisioningUpdates") {
		t.Fatalf("output should not include -allowProvisioningUpdates:\n%s", stdout.String())
	}
}

func runTestGit(t *testing.T, dir string, args ...string) {
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
