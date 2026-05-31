package release

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePlatformVersionAcceptsShortAppleVersion(t *testing.T) {
	version, err := ParsePlatformVersion("1.0")
	if err != nil {
		t.Fatal(err)
	}
	if version.String() != "1.0.0" {
		t.Fatalf("version = %q, want 1.0.0", version.String())
	}
}

func TestReadAndroidVersionUsesGradleDefaults(t *testing.T) {
	buildFile := filepath.Join(t.TempDir(), "build.gradle.kts")
	content := `
android {
    defaultConfig {
        versionName = providers.gradleProperty("emsi.versionName").getOrElse("0.1.0")
        versionCode = providers.gradleProperty("emsi.versionCode").map(String::toInt).getOrElse(1)
    }
}
`
	if err := os.WriteFile(buildFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	current, err := ReadAndroidVersion(buildFile)
	if err != nil {
		t.Fatal(err)
	}
	if current.Version.String() != "0.1.0" {
		t.Fatalf("Version = %q, want 0.1.0", current.Version.String())
	}
	if current.Build != "1" {
		t.Fatalf("Build = %q, want 1", current.Build)
	}
	if current.VersionSource != buildFile+":versionName" {
		t.Fatalf("VersionSource = %q", current.VersionSource)
	}
}

func TestReadIOSVersionUsesProjectBuildSettings(t *testing.T) {
	projectDir := filepath.Join(t.TempDir(), "Example.xcodeproj")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	projectFile := filepath.Join(projectDir, "project.pbxproj")
	content := `
buildSettings = {
    CURRENT_PROJECT_VERSION = 7;
    MARKETING_VERSION = 1.0;
};
buildSettings = {
    CURRENT_PROJECT_VERSION = 7;
    MARKETING_VERSION = 1.0;
};
`
	if err := os.WriteFile(projectFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	current, err := ReadIOSVersion(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if current.Version.String() != "1.0.0" {
		t.Fatalf("Version = %q, want 1.0.0", current.Version.String())
	}
	if current.Build != "7" {
		t.Fatalf("Build = %q, want 7", current.Build)
	}
	if current.BuildSource != projectFile+":CURRENT_PROJECT_VERSION" {
		t.Fatalf("BuildSource = %q", current.BuildSource)
	}
}
