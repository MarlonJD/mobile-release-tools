package release

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

type Command struct {
	Dir  string
	Name string
	Args []string
}

func (command Command) String() string {
	parts := make([]string, 0, len(command.Args)+1)
	parts = append(parts, shellQuote(command.Name))
	for _, arg := range command.Args {
		parts = append(parts, shellQuote(arg))
	}
	if command.Dir == "" || command.Dir == "." {
		return strings.Join(parts, " ")
	}
	return "(cd " + shellQuote(command.Dir) + " && " + strings.Join(parts, " ") + ")"
}

type AndroidPackageOptions struct {
	ProjectDir string
	Gradle     string
	Module     string
	Version    Version
	Build      string
	Channel    string
	Signing    string
	SkipTests  bool
	IncludeAPK bool
}

const (
	AndroidSigningEnv      = "env"
	AndroidSigningExternal = "external"
	AndroidSigningUnsigned = "unsigned"
)

var AndroidSigningEnvironmentKeys = []string{
	"EMSI_ANDROID_RELEASE_STORE_FILE",
	"EMSI_ANDROID_RELEASE_STORE_PASSWORD",
	"EMSI_ANDROID_RELEASE_KEY_ALIAS",
	"EMSI_ANDROID_RELEASE_KEY_PASSWORD",
}

func AndroidPackageCommands(options AndroidPackageOptions) ([]Command, error) {
	if options.ProjectDir == "" {
		options.ProjectDir = "."
	}
	if options.Gradle == "" {
		options.Gradle = "./gradlew"
	}
	if options.Module == "" {
		options.Module = ":app"
	}
	if options.Channel == "" {
		options.Channel = "production"
	}
	if options.Signing == "" {
		options.Signing = AndroidSigningEnv
	}
	if options.Build == "" {
		return nil, fmt.Errorf("android build is required")
	}
	if _, err := strconv.Atoi(options.Build); err != nil {
		return nil, fmt.Errorf("android build must be an integer versionCode: %w", err)
	}

	properties := []string{
		"-Pemsi.versionName=" + options.Version.String(),
		"-Pemsi.versionCode=" + options.Build,
		"-Pemsi.distributionChannel=" + options.Channel,
	}

	commands := make([]Command, 0, 3)
	if !options.SkipTests {
		commands = append(commands, Command{
			Dir:  options.ProjectDir,
			Name: options.Gradle,
			Args: []string{options.Module + ":testDebugUnitTest"},
		})
	}
	commands = append(commands, Command{
		Dir:  options.ProjectDir,
		Name: options.Gradle,
		Args: append([]string{options.Module + ":bundleRelease"}, properties...),
	})
	if options.IncludeAPK {
		commands = append(commands, Command{
			Dir:  options.ProjectDir,
			Name: options.Gradle,
			Args: append([]string{options.Module + ":assembleRelease"}, properties...),
		})
	}
	return commands, nil
}

func ValidateAndroidSigning(signing string, lookup func(string) (string, bool)) error {
	if signing == "" {
		signing = AndroidSigningEnv
	}
	switch signing {
	case AndroidSigningEnv:
		missing := make([]string, 0)
		for _, key := range AndroidSigningEnvironmentKeys {
			value, ok := lookup(key)
			if !ok || strings.TrimSpace(value) == "" {
				missing = append(missing, key)
			}
		}
		if len(missing) > 0 {
			return fmt.Errorf("android signing env is incomplete; missing %s. Use --signing external when Gradle properties provide signing, or --signing unsigned for QA-only unsigned artifacts", strings.Join(missing, ", "))
		}
		return nil
	case AndroidSigningExternal, AndroidSigningUnsigned:
		return nil
	default:
		return fmt.Errorf("unsupported android signing mode %q; expected env, external, or unsigned", signing)
	}
}

type IOSPackageOptions struct {
	ProjectPath              string
	Scheme                   string
	Configuration            string
	Version                  Version
	Build                    string
	ArchivePath              string
	ExportPath               string
	ExportOptions            string
	ArchiveDest              string
	TestDestination          string
	SkipTesting              string
	SkipTests                bool
	AllowProvisioningUpdates bool
}

func IOSPackageCommands(options IOSPackageOptions) ([]Command, error) {
	if options.ProjectPath == "" {
		return nil, fmt.Errorf("ios project path is required")
	}
	if options.Scheme == "" {
		return nil, fmt.Errorf("ios scheme is required")
	}
	if options.Build == "" {
		return nil, fmt.Errorf("ios build is required")
	}
	if options.Configuration == "" {
		options.Configuration = "Release"
	}
	if options.ArchiveDest == "" {
		options.ArchiveDest = "generic/platform=iOS"
	}
	if options.TestDestination == "" {
		options.TestDestination = "platform=iOS Simulator,name=iPhone 17"
	}
	releaseID := options.Version.String() + "+" + options.Build
	if options.ArchivePath == "" {
		options.ArchivePath = filepath.Join("build", "releases", "ios", releaseID, options.Scheme+".xcarchive")
	}
	if options.ExportPath == "" {
		options.ExportPath = filepath.Join("build", "releases", "ios", releaseID, "export")
	}
	if options.ExportOptions == "" {
		return nil, fmt.Errorf("ios export options plist is required")
	}

	commands := make([]Command, 0, 3)
	if !options.SkipTests {
		args := []string{
			"test",
			"-project", options.ProjectPath,
			"-scheme", options.Scheme,
			"-destination", options.TestDestination,
		}
		if options.SkipTesting != "" {
			args = append(args, "-skip-testing:"+options.SkipTesting)
		}
		args = append(args, "CODE_SIGNING_ALLOWED=NO")
		commands = append(commands, Command{Name: "xcodebuild", Args: args})
	}

	commands = append(commands, Command{
		Name: "xcodebuild",
		Args: []string{
			"archive",
			"-project", options.ProjectPath,
			"-scheme", options.Scheme,
			"-configuration", options.Configuration,
			"-destination", options.ArchiveDest,
			"-archivePath", options.ArchivePath,
			"MARKETING_VERSION=" + options.Version.String(),
			"CURRENT_PROJECT_VERSION=" + options.Build,
		},
	})
	if options.AllowProvisioningUpdates {
		commands[len(commands)-1].Args = append(commands[len(commands)-1].Args, "-allowProvisioningUpdates")
	}
	commands = append(commands, Command{
		Name: "xcodebuild",
		Args: []string{
			"-exportArchive",
			"-archivePath", options.ArchivePath,
			"-exportPath", options.ExportPath,
			"-exportOptionsPlist", options.ExportOptions,
		},
	})
	if options.AllowProvisioningUpdates {
		commands[len(commands)-1].Args = append(commands[len(commands)-1].Args, "-allowProvisioningUpdates")
	}
	return commands, nil
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if strings.ContainsAny(value, " \t\n\"'\\$`") {
		return strconv.Quote(value)
	}
	return value
}
