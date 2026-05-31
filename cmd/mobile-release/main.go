package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/MarlonJD/mobile-release-tools/pkg/release"
)

const usage = `mobile-release manages mobile release metadata.

Usage:
  mobile-release bump --current 1.4.2 --level patch|minor|major
  mobile-release changelog --repo . --from v1.4.2 --to HEAD --version 1.4.3 [--output RELEASE_NOTES.md]
  mobile-release hash --file path/to/artifact
  mobile-release manifest --platform ios|android --version 1.4.3 --build 104 --artifact path [--notes RELEASE_NOTES.md] [--output manifest.json]
  mobile-release package android --channel production
  mobile-release package ios
`

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(stdout, usage)
		return nil
	}

	switch args[0] {
	case "bump":
		return runBump(args[1:], stdout)
	case "changelog":
		return runChangelog(args[1:], stdout)
	case "hash":
		return runHash(args[1:], stdout)
	case "manifest":
		return runManifest(args[1:], stdout)
	case "package":
		return runPackage(args[1:], stdout)
	case "mobile":
		return runMobile(args[1:], stdout)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func runMobile(args []string, stdout io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: mobile-release mobile package ios|android")
	}
	if args[0] != "package" {
		return fmt.Errorf("unknown mobile command %q; supported command: package", args[0])
	}

	return runPackage(args[1:], stdout)
}

func runPackage(args []string, stdout io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mobile-release package ios|android")
	}

	switch args[0] {
	case "android":
		return runMobilePackageAndroid(args[1:], stdout)
	case "ios":
		return runMobilePackageIOS(args[1:], stdout)
	default:
		return fmt.Errorf("unknown package platform %q; expected ios or android", args[0])
	}
}

func runBump(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("bump", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	current := fs.String("current", "", "current semantic version")
	level := fs.String("level", "patch", "bump level: patch, minor, or major")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *current == "" {
		return fmt.Errorf("missing --current")
	}

	version, err := release.ParseVersion(*current)
	if err != nil {
		return err
	}
	bumped, err := version.Bump(release.BumpLevel(*level))
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, bumped.String())
	return nil
}

func runChangelog(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("changelog", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repo", ".", "git repository path")
	from := fs.String("from", "", "previous release tag or commit")
	to := fs.String("to", "HEAD", "target release ref")
	version := fs.String("version", "Unreleased", "release version label")
	output := fs.String("output", "-", "output file path or - for stdout")
	includeInternal := fs.Bool("include-internal", false, "include internal commit types")
	if err := fs.Parse(args); err != nil {
		return err
	}

	commits, err := release.LoadGitCommits(*repo, *from, *to)
	if err != nil {
		return err
	}
	notes := release.GenerateChangelog(*version, time.Now().UTC(), commits, release.ChangelogOptions{
		IncludeInternal: *includeInternal,
	})
	return writeOutput(*output, []byte(notes), stdout)
}

func runHash(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("hash", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	file := fs.String("file", "", "artifact file path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *file == "" {
		return fmt.Errorf("missing --file")
	}

	artifact, err := release.NewFileArtifact(*file)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%s  %s\n", artifact.SHA256, artifact.Path)
	return nil
}

func runManifest(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("manifest", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	platform := fs.String("platform", "", "ios or android")
	versionInput := fs.String("version", "", "semantic version")
	build := fs.String("build", "", "platform build number or version code")
	artifactsInput := fs.String("artifact", "", "comma-separated artifact paths")
	notesPath := fs.String("notes", "", "release notes file path")
	output := fs.String("output", "-", "output file path or - for stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *platform == "" || *versionInput == "" || *build == "" || *artifactsInput == "" {
		return fmt.Errorf("missing required flags: --platform, --version, --build, and --artifact are required")
	}

	version, err := release.ParseVersion(*versionInput)
	if err != nil {
		return err
	}

	var notes string
	if *notesPath != "" {
		data, err := os.ReadFile(*notesPath)
		if err != nil {
			return err
		}
		notes = string(data)
	}

	manifest, err := release.NewManifest(*platform, version, *build, splitCSV(*artifactsInput), notes, time.Now().UTC())
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeOutput(*output, data, stdout)
}

func runMobilePackageAndroid(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mobile package android", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	projectDir := fs.String("project", "apps/android", "Android project directory")
	buildFile := fs.String("build-file", "", "Android Gradle build file used to read current version; defaults to <project>/app/build.gradle.kts")
	gradle := fs.String("gradle", "./gradlew", "Gradle executable path relative to project")
	module := fs.String("module", ":app", "Gradle app module path")
	versionInput := fs.String("version", "", "Android versionName semantic version")
	build := fs.String("build", "", "Android versionCode")
	repo := fs.String("repo", ".", "git repository path used for automatic version calculation")
	from := fs.String("from", "", "previous release tag or commit; defaults to latest SemVer tag")
	to := fs.String("to", "HEAD", "target release ref used for automatic version calculation")
	tagPrefix := fs.String("tag-prefix", "v", "release tag prefix used for automatic version calculation")
	channel := fs.String("channel", "production", "distribution channel")
	signing := fs.String("signing", release.AndroidSigningEnv, "signing source: env, external, or unsigned")
	notesOutput := fs.String("notes-output", "", "release notes output path; defaults to build/releases/android/<version+build>/RELEASE_NOTES.md")
	manifestOutput := fs.String("manifest-output", "", "release manifest output path; defaults to build/releases/android/<version+build>/release-manifest.json")
	uploadDestination := fs.String("upload-destination", "Google Play Console", "human-readable upload destination printed after packaging")
	skipTests := fs.Bool("skip-tests", false, "skip Android unit tests")
	includeAPK := fs.Bool("include-apk", false, "also build release APK for QA")
	dryRun := fs.Bool("dry-run", false, "print commands without running them")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *buildFile == "" {
		*buildFile = filepath.Join(*projectDir, "app", "build.gradle.kts")
	}
	current, err := release.ReadAndroidVersion(*buildFile)
	if err != nil && (*versionInput == "" || *build == "") {
		return err
	}
	plan, err := release.PlanNextRelease(release.ReleasePlanOptions{
		RepoPath:        *repo,
		FromRef:         *from,
		ToRef:           *to,
		TagPrefix:       *tagPrefix,
		CurrentVersion:  current.Version,
		CurrentBuild:    current.Build,
		VersionSource:   current.VersionSource,
		BuildSource:     current.BuildSource,
		VersionOverride: *versionInput,
		BuildOverride:   *build,
	})
	if err != nil {
		return err
	}
	if err := release.ValidateAndroidSigning(*signing, os.LookupEnv); err != nil {
		return err
	}
	commands, err := release.AndroidPackageCommands(release.AndroidPackageOptions{
		ProjectDir: *projectDir,
		Gradle:     *gradle,
		Module:     *module,
		Version:    plan.Version,
		Build:      plan.Build,
		Channel:    *channel,
		Signing:    *signing,
		SkipTests:  *skipTests,
		IncludeAPK: *includeAPK,
	})
	if err != nil {
		return err
	}
	if err := executeCommands(commands, *dryRun, stdout); err != nil {
		return err
	}

	releaseDir := filepath.Join("build", "releases", "android", plan.ReleaseID)
	notesPath, manifestPath := releaseMetadataPaths(releaseDir, *notesOutput, *manifestOutput)
	artifacts := androidArtifactPaths(*projectDir, *includeAPK, *dryRun)
	return finishPackage(finishPackageOptions{
		Platform:          "android",
		Plan:              plan,
		Artifacts:         artifacts,
		ReleaseDir:        releaseDir,
		NotesPath:         notesPath,
		ManifestPath:      manifestPath,
		UploadDestination: *uploadDestination,
		DryRun:            *dryRun,
		Stdout:            stdout,
	})
}

func runMobilePackageIOS(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("mobile package ios", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	projectPath := fs.String("project", "apps/ios/emsi_ios.xcodeproj", "Xcode project path")
	scheme := fs.String("scheme", "emsi_ios", "Xcode scheme")
	configuration := fs.String("configuration", "Release", "Xcode build configuration")
	versionInput := fs.String("version", "", "iOS MARKETING_VERSION semantic version")
	build := fs.String("build", "", "iOS CURRENT_PROJECT_VERSION build number")
	repo := fs.String("repo", ".", "git repository path used for automatic version calculation")
	from := fs.String("from", "", "previous release tag or commit; defaults to latest SemVer tag")
	to := fs.String("to", "HEAD", "target release ref used for automatic version calculation")
	tagPrefix := fs.String("tag-prefix", "v", "release tag prefix used for automatic version calculation")
	archivePath := fs.String("archive-path", "", "xcarchive output path")
	exportPath := fs.String("export-path", "", "IPA export output directory")
	exportOptions := fs.String("export-options", "apps/ios/release/ExportOptions-app-store.plist", "App Store export options plist")
	notesOutput := fs.String("notes-output", "", "release notes output path; defaults to build/releases/ios/<version+build>/RELEASE_NOTES.md")
	manifestOutput := fs.String("manifest-output", "", "release manifest output path; defaults to build/releases/ios/<version+build>/release-manifest.json")
	uploadDestination := fs.String("upload-destination", "App Store Connect", "human-readable upload destination printed after packaging")
	archiveDestination := fs.String("archive-destination", "generic/platform=iOS", "xcodebuild archive destination")
	testDestination := fs.String("test-destination", "platform=iOS Simulator,name=iPhone 17", "xcodebuild test destination")
	skipTesting := fs.String("skip-testing", "emsi_iosUITests", "optional xcodebuild -skip-testing target")
	skipTests := fs.Bool("skip-tests", false, "skip iOS unit tests")
	allowProvisioningUpdates := fs.Bool("allow-provisioning-updates", true, "allow Xcode to update signing assets; set false to require existing signing assets")
	dryRun := fs.Bool("dry-run", false, "print commands without running them")
	if err := fs.Parse(args); err != nil {
		return err
	}

	current, err := release.ReadIOSVersion(*projectPath)
	if err != nil && (*versionInput == "" || *build == "") {
		return err
	}
	plan, err := release.PlanNextRelease(release.ReleasePlanOptions{
		RepoPath:        *repo,
		FromRef:         *from,
		ToRef:           *to,
		TagPrefix:       *tagPrefix,
		CurrentVersion:  current.Version,
		CurrentBuild:    current.Build,
		VersionSource:   current.VersionSource,
		BuildSource:     current.BuildSource,
		VersionOverride: *versionInput,
		BuildOverride:   *build,
	})
	if err != nil {
		return err
	}
	if *archivePath == "" {
		*archivePath = filepath.Join("build", "releases", "ios", plan.ReleaseID, *scheme+".xcarchive")
	}
	if *exportPath == "" {
		*exportPath = filepath.Join("build", "releases", "ios", plan.ReleaseID, "export")
	}

	commands, err := release.IOSPackageCommands(release.IOSPackageOptions{
		ProjectPath:              *projectPath,
		Scheme:                   *scheme,
		Configuration:            *configuration,
		Version:                  plan.Version,
		Build:                    plan.Build,
		ArchivePath:              *archivePath,
		ExportPath:               *exportPath,
		ExportOptions:            *exportOptions,
		ArchiveDest:              *archiveDestination,
		TestDestination:          *testDestination,
		SkipTesting:              *skipTesting,
		SkipTests:                *skipTests,
		AllowProvisioningUpdates: *allowProvisioningUpdates,
	})
	if err != nil {
		return err
	}
	if !*dryRun {
		if err := os.MkdirAll(filepath.Dir(*archivePath), 0o755); err != nil {
			return err
		}
		if err := os.MkdirAll(*exportPath, 0o755); err != nil {
			return err
		}
	}
	if err := executeCommands(commands, *dryRun, stdout); err != nil {
		return err
	}

	releaseDir := filepath.Join("build", "releases", "ios", plan.ReleaseID)
	notesPath, manifestPath := releaseMetadataPaths(releaseDir, *notesOutput, *manifestOutput)
	artifacts := iosArtifactPaths(*exportPath, *dryRun)
	return finishPackage(finishPackageOptions{
		Platform:          "ios",
		Plan:              plan,
		Artifacts:         artifacts,
		ReleaseDir:        releaseDir,
		NotesPath:         notesPath,
		ManifestPath:      manifestPath,
		UploadDestination: *uploadDestination,
		DryRun:            *dryRun,
		Stdout:            stdout,
	})
}

type finishPackageOptions struct {
	Platform          string
	Plan              release.ReleasePlan
	Artifacts         []string
	ReleaseDir        string
	NotesPath         string
	ManifestPath      string
	UploadDestination string
	DryRun            bool
	Stdout            io.Writer
}

func finishPackage(options finishPackageOptions) error {
	notes := release.GenerateChangelog(options.Plan.Version.String(), time.Now().UTC(), options.Plan.Commits, release.ChangelogOptions{})
	if options.DryRun {
		printPackageSummary(options, nil)
		return nil
	}

	if err := os.MkdirAll(options.ReleaseDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(options.NotesPath), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(options.ManifestPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(options.NotesPath, []byte(notes), 0o644); err != nil {
		return err
	}
	manifest, err := release.NewManifest(options.Platform, options.Plan.Version, options.Plan.Build, options.Artifacts, notes, time.Now().UTC())
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(options.ManifestPath, data, 0o644); err != nil {
		return err
	}
	printPackageSummary(options, &manifest)
	return nil
}

func printPackageSummary(options finishPackageOptions, manifest *release.Manifest) {
	status := "Release package complete"
	if options.DryRun {
		status = "Release package dry run"
	}
	fmt.Fprintln(options.Stdout)
	fmt.Fprintln(options.Stdout, status)
	fmt.Fprintf(options.Stdout, "  Platform: %s\n", options.Platform)
	if options.Plan.PreviousRef != "" {
		fmt.Fprintf(options.Stdout, "  Changelog from: %s\n", options.Plan.PreviousRef)
	} else {
		fmt.Fprintln(options.Stdout, "  Changelog from: beginning of git history")
	}
	fmt.Fprintf(options.Stdout, "  Current version: %s\n", options.Plan.CurrentVersion.String())
	if options.Plan.CurrentBuild != "" {
		fmt.Fprintf(options.Stdout, "  Current build: %s\n", options.Plan.CurrentBuild)
	}
	if options.Plan.VersionSource != "" {
		fmt.Fprintf(options.Stdout, "  Current version source: %s\n", options.Plan.VersionSource)
	}
	if options.Plan.BuildSource != "" {
		fmt.Fprintf(options.Stdout, "  Current build source: %s\n", options.Plan.BuildSource)
	}
	fmt.Fprintf(options.Stdout, "  Version: %s\n", options.Plan.Version.String())
	fmt.Fprintf(options.Stdout, "  Build: %s\n", options.Plan.Build)
	fmt.Fprintf(options.Stdout, "  Release ID: %s\n", options.Plan.ReleaseID)
	fmt.Fprintf(options.Stdout, "  Bump: %s (%s)\n", options.Plan.BumpLevel, options.Plan.BumpReason)
	if options.Plan.UsedVersionFlag || options.Plan.UsedBuildFlag {
		fmt.Fprintln(options.Stdout, "  Override: manual version/build flag used")
	}
	fmt.Fprintf(options.Stdout, "  Release notes: %s\n", options.NotesPath)
	fmt.Fprintf(options.Stdout, "  Manifest: %s\n", options.ManifestPath)
	fmt.Fprintf(options.Stdout, "  Upload destination: %s\n", options.UploadDestination)
	if manifest == nil {
		for _, artifact := range options.Artifacts {
			fmt.Fprintf(options.Stdout, "  Upload artifact: %s\n", artifact)
		}
		return
	}
	for _, artifact := range manifest.Artifacts {
		fmt.Fprintf(options.Stdout, "  Upload artifact: %s\n", artifact.Path)
		fmt.Fprintf(options.Stdout, "    SHA-256: %s\n", artifact.SHA256)
	}
}

func releaseMetadataPaths(releaseDir, notesOutput, manifestOutput string) (string, string) {
	if notesOutput == "" {
		notesOutput = filepath.Join(releaseDir, "RELEASE_NOTES.md")
	}
	if manifestOutput == "" {
		manifestOutput = filepath.Join(releaseDir, "release-manifest.json")
	}
	return notesOutput, manifestOutput
}

func androidArtifactPaths(projectDir string, includeAPK bool, dryRun bool) []string {
	aabPattern := filepath.Join(projectDir, "app", "build", "outputs", "bundle", "release", "*.aab")
	paths := globOrDefault(aabPattern, filepath.Join(projectDir, "app", "build", "outputs", "bundle", "release", "app-release.aab"), dryRun)
	if includeAPK {
		apkPattern := filepath.Join(projectDir, "app", "build", "outputs", "apk", "release", "*.apk")
		paths = append(paths, globOrDefault(apkPattern, filepath.Join(projectDir, "app", "build", "outputs", "apk", "release", "app-release.apk"), dryRun)...)
	}
	return paths
}

func iosArtifactPaths(exportPath string, dryRun bool) []string {
	return globOrDefault(filepath.Join(exportPath, "*.ipa"), filepath.Join(exportPath, "*.ipa"), dryRun)
}

func globOrDefault(pattern, fallback string, dryRun bool) []string {
	if !dryRun {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return matches
		}
	}
	return []string{fallback}
}

func splitCSV(input string) []string {
	parts := strings.Split(input, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func writeOutput(path string, data []byte, stdout io.Writer) error {
	if path == "" || path == "-" {
		_, err := stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func executeCommands(commands []release.Command, dryRun bool, stdout io.Writer) error {
	for _, command := range commands {
		fmt.Fprintln(stdout, command.String())
		if dryRun {
			continue
		}
		process := exec.Command(command.Name, command.Args...)
		process.Dir = command.Dir
		process.Stdin = os.Stdin
		process.Stdout = os.Stdout
		process.Stderr = os.Stderr
		if err := process.Run(); err != nil {
			return fmt.Errorf("run %s: %w", command.String(), err)
		}
	}
	return nil
}
