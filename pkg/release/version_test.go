package release

import "testing"

func TestParseVersionSupportsPrereleaseAndBuildMetadata(t *testing.T) {
	version, err := ParseVersion("v1.2.3-beta.1+build.7")
	if err != nil {
		t.Fatal(err)
	}

	if version.String() != "1.2.3-beta.1+build.7" {
		t.Fatalf("String() = %q", version.String())
	}
}

func TestVersionBumpDropsPrereleaseAndMetadata(t *testing.T) {
	version, err := ParseVersion("1.2.3-beta.1+build.7")
	if err != nil {
		t.Fatal(err)
	}

	next, err := version.Bump(BumpMinor)
	if err != nil {
		t.Fatal(err)
	}

	if next.String() != "1.3.0" {
		t.Fatalf("next = %q, want 1.3.0", next.String())
	}
}

func TestParseVersionRejectsLeadingZero(t *testing.T) {
	if _, err := ParseVersion("1.02.3"); err == nil {
		t.Fatal("expected leading zero to be rejected")
	}
}
