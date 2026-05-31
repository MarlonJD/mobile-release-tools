package release

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type PlatformVersion struct {
	Version       Version
	Build         string
	VersionSource string
	BuildSource   string
}

var (
	androidVersionNamePattern        = regexp.MustCompile(`providers\.gradleProperty\("emsi\.versionName"\)\.getOrElse\("([^"]+)"\)`)
	androidVersionCodePattern        = regexp.MustCompile(`(?s)providers\.gradleProperty\("emsi\.versionCode"\).*?getOrElse\(([0-9]+)\)`)
	androidLiteralVersionNamePattern = regexp.MustCompile(`versionName\s*=\s*"([^"]+)"`)
	androidLiteralVersionCodePattern = regexp.MustCompile(`versionCode\s*=\s*([0-9]+)`)
	iosMarketingVersionPattern       = regexp.MustCompile(`MARKETING_VERSION\s*=\s*([^;]+);`)
	iosCurrentProjectVersionPattern  = regexp.MustCompile(`CURRENT_PROJECT_VERSION\s*=\s*([^;]+);`)
)

func ReadAndroidVersion(buildFile string) (PlatformVersion, error) {
	if buildFile == "" {
		buildFile = filepath.Join("apps", "android", "app", "build.gradle.kts")
	}
	data, err := os.ReadFile(buildFile)
	if err != nil {
		return PlatformVersion{}, fmt.Errorf("read android build file: %w", err)
	}
	content := string(data)

	versionInput := firstMatch(content, androidVersionNamePattern)
	if versionInput == "" {
		versionInput = firstMatch(content, androidLiteralVersionNamePattern)
	}
	if versionInput == "" {
		return PlatformVersion{}, fmt.Errorf("android versionName not found in %s", buildFile)
	}
	build := firstMatch(content, androidVersionCodePattern)
	if build == "" {
		build = firstMatch(content, androidLiteralVersionCodePattern)
	}
	if build == "" {
		return PlatformVersion{}, fmt.Errorf("android versionCode not found in %s", buildFile)
	}
	version, err := ParsePlatformVersion(versionInput)
	if err != nil {
		return PlatformVersion{}, fmt.Errorf("parse android versionName from %s: %w", buildFile, err)
	}
	return PlatformVersion{
		Version:       version,
		Build:         build,
		VersionSource: buildFile + ":versionName",
		BuildSource:   buildFile + ":versionCode",
	}, nil
}

func ReadIOSVersion(projectPath string) (PlatformVersion, error) {
	projectFile := iosProjectFile(projectPath)
	data, err := os.ReadFile(projectFile)
	if err != nil {
		return PlatformVersion{}, fmt.Errorf("read ios project file: %w", err)
	}
	content := string(data)

	versionInput := mostCommonMatch(content, iosMarketingVersionPattern)
	if versionInput == "" {
		return PlatformVersion{}, fmt.Errorf("ios MARKETING_VERSION not found in %s", projectFile)
	}
	build := mostCommonMatch(content, iosCurrentProjectVersionPattern)
	if build == "" {
		return PlatformVersion{}, fmt.Errorf("ios CURRENT_PROJECT_VERSION not found in %s", projectFile)
	}
	version, err := ParsePlatformVersion(versionInput)
	if err != nil {
		return PlatformVersion{}, fmt.Errorf("parse ios MARKETING_VERSION from %s: %w", projectFile, err)
	}
	return PlatformVersion{
		Version:       version,
		Build:         build,
		VersionSource: projectFile + ":MARKETING_VERSION",
		BuildSource:   projectFile + ":CURRENT_PROJECT_VERSION",
	}, nil
}

func iosProjectFile(projectPath string) string {
	if projectPath == "" {
		projectPath = filepath.Join("apps", "ios", "emsi_ios.xcodeproj")
	}
	if strings.HasSuffix(projectPath, ".xcodeproj") {
		return filepath.Join(projectPath, "project.pbxproj")
	}
	return projectPath
}

func firstMatch(content string, pattern *regexp.Regexp) string {
	matches := pattern.FindStringSubmatch(content)
	if len(matches) < 2 {
		return ""
	}
	return cleanProjectValue(matches[1])
}

func mostCommonMatch(content string, pattern *regexp.Regexp) string {
	matches := pattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return ""
	}
	counts := map[string]int{}
	order := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		value := cleanProjectValue(match[1])
		if value == "" {
			continue
		}
		if counts[value] == 0 {
			order = append(order, value)
		}
		counts[value]++
	}
	best := ""
	for _, value := range order {
		if best == "" || counts[value] > counts[best] {
			best = value
		}
	}
	return best
}

func cleanProjectValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	return value
}
