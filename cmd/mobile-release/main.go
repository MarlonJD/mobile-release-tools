package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
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
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
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
