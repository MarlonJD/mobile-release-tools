package release

import (
	"strings"
	"testing"
)

func TestAndroidPackageCommandsBuildsSignedAABCommand(t *testing.T) {
	version, err := ParseVersion("1.4.3")
	if err != nil {
		t.Fatal(err)
	}

	commands, err := AndroidPackageCommands(AndroidPackageOptions{
		ProjectDir: "apps/android",
		Version:    version,
		Build:      "1848",
		Channel:    "production",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(commands) != 2 {
		t.Fatalf("len(commands) = %d, want 2", len(commands))
	}
	bundle := commands[1].String()
	for _, want := range []string{
		"./gradlew :app:bundleRelease",
		"-Pemsi.versionName=1.4.3",
		"-Pemsi.versionCode=1848",
		"-Pemsi.distributionChannel=production",
	} {
		if !strings.Contains(bundle, want) {
			t.Fatalf("bundle command missing %q: %s", want, bundle)
		}
	}
}

func TestValidateAndroidSigningRequiresEnvByDefault(t *testing.T) {
	err := ValidateAndroidSigning("", func(string) (string, bool) {
		return "", false
	})
	if err == nil {
		t.Fatal("expected missing signing env error")
	}
}

func TestValidateAndroidSigningAllowsExternalSigning(t *testing.T) {
	err := ValidateAndroidSigning(AndroidSigningExternal, func(string) (string, bool) {
		return "", false
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSPackageCommandsBuildsArchiveAndExportCommands(t *testing.T) {
	version, err := ParseVersion("1.4.3")
	if err != nil {
		t.Fatal(err)
	}

	commands, err := IOSPackageCommands(IOSPackageOptions{
		ProjectPath:   "apps/ios/emsi_ios.xcodeproj",
		Scheme:        "emsi_ios",
		Version:       version,
		Build:         "1848",
		ExportOptions: "apps/ios/release/ExportOptions-app-store.plist",
		SkipTesting:   "emsi_iosUITests",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(commands) != 3 {
		t.Fatalf("len(commands) = %d, want 3", len(commands))
	}
	archive := commands[1].String()
	for _, want := range []string{
		"xcodebuild archive",
		"-project apps/ios/emsi_ios.xcodeproj",
		"-scheme emsi_ios",
		"MARKETING_VERSION=1.4.3",
		"CURRENT_PROJECT_VERSION=1848",
	} {
		if !strings.Contains(archive, want) {
			t.Fatalf("archive command missing %q: %s", want, archive)
		}
	}
	export := commands[2].String()
	if !strings.Contains(export, "-exportOptionsPlist apps/ios/release/ExportOptions-app-store.plist") {
		t.Fatalf("export command missing export options: %s", export)
	}
}
