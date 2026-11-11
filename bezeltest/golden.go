package bezeltest

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// update is `go test -update`: Golden and GoldenANSI write what they got
// instead of comparing it.
var update = flag.Bool("update", false, "rewrite the golden files under testdata")

// Golden compares got, its escape codes stripped, with
// testdata/<test name>.golden; run `go test -update` to write that file.
func Golden(t testing.TB, got string) {
	t.Helper()
	golden(t, ansi.Strip(got), ".golden", *update)
}

// GoldenANSI is Golden with the escape codes kept, so a colour that changes
// fails too; its file is testdata/<test name>.ansi.golden.
func GoldenANSI(t testing.TB, got string) {
	t.Helper()
	golden(t, got, ".ansi.golden", *update)
}

func golden(t testing.TB, got, suffix string, write bool) {
	t.Helper()
	path := goldenPath(t.Name(), suffix)
	if write {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatalf("bezeltest: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatalf("bezeltest: %v", err)
		}
		t.Logf("bezeltest: wrote %s", path)
		return
	}
	data, err := os.ReadFile(path) // #nosec G304 -- the test's own testdata file
	if errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("bezeltest: %s is missing; run go test -update to write it", path)
	}
	if err != nil {
		t.Fatalf("bezeltest: %v", err)
	}
	// A checkout that turned LF into CRLF is still the same screen.
	want := strings.ReplaceAll(string(data), "\r\n", "\n")
	if diff := lineDiff(want, got); diff != "" {
		t.Errorf("bezeltest: the screen differs from %s (go test -update rewrites it):\n%s", path, diff)
	}
}

// goldenPath is the file of the test named name, each subtest a folder.
func goldenPath(name, suffix string) string {
	clean := strings.NewReplacer(" ", "_", ":", "_", "\\", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_").Replace(name)
	return filepath.Join("testdata", filepath.FromSlash(clean)+suffix)
}

// lineDiff lists the lines where want and got differ, numbered from 1, and
// is empty when they are the same.
func lineDiff(want, got string) string {
	if want == got {
		return ""
	}
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	var b strings.Builder
	for i := range max(len(w), len(g)) {
		wl, gl := line(w, i), line(g, i)
		if wl == gl {
			continue
		}
		fmt.Fprintf(&b, "line %d\n  want: %q\n  got:  %q\n", i+1, wl, gl)
	}
	return b.String()
}

func line(lines []string, i int) string {
	if i < len(lines) {
		return lines[i]
	}
	return "(no line)"
}
