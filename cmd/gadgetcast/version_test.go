package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// releaseLdflag is the -X flag release.yml stamps the tag with; keep the two in step.
const releaseLdflag = "-X github.com/adeelahmad/gadgetcast/internal/server.Version="

// The v0.1.1 release binary reported 0.1.0: the stamp named a symbol that did
// not exist, which the linker ignores. Build the way the release does and check
// the binary reports the stamped version.
func TestReleaseStampSetsVersion(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "gadgetcast")
	build := exec.Command("go", "build", "-ldflags", releaseLdflag+"9.8.7-test", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	out, err := exec.Command(bin, "-version").Output()
	if err != nil {
		t.Fatalf("-version: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "9.8.7-test" {
		t.Fatalf("-version = %q, want the stamped 9.8.7-test", got)
	}
}
