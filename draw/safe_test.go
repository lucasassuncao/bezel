package draw

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// The terminal draws a stray byte as U+FFFD, but ansi measures it with the
// bytes after it: an invalid byte would knock a padded cell out of line.
func TestSanitizeEscapesInvalidUTF8(t *testing.T) {
	cases := map[string]string{
		"0\xe8":        `0\xe8`,
		"a\xffb":       `a\xffb`,
		"caf\xc3":      `caf\xc3`,
		"ok \uFFFD":    "ok \uFFFD",
		"\xe8\x1b[31m": `\xe8\e[31m`,
	}
	for in, want := range cases {
		if got := Sanitize(in); got != want {
			t.Errorf("Sanitize(%q) = %q, want %q", in, got, want)
		}
	}
}

// Control sequences from untrusted text are shown, never executed: an OSC 52
// that would set the clipboard comes out as its escaped bytes.
func TestSanitizeShowsControlSequences(t *testing.T) {
	in, want := "name\x1b]52;c;aGk=\x07\n", `name\e]52;c;aGk=\x07\n`
	if got := Sanitize(in); got != want {
		t.Errorf("Sanitize(%q) = %q, want %q", in, got, want)
	}
}

// Whatever bytes a value carries, what comes out of sanitize is safe to draw:
// no rune the terminal would act on, and nothing lost that was safe to begin
// with. Invalid UTF-8 is included, because a secret is not obliged to be text.
func FuzzSanitizeIsSafeToDraw(f *testing.F) {
	rtlOverride := string(rune(0x202e))
	for _, s := range []string{"", "plain", "a\x1b[31mb", rtlOverride, "tab\there", "\xff\xfe", "caf" + string(rune(0xe9))} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		out := Sanitize(s)
		if strings.ContainsFunc(out, unsafeToDraw) || !utf8.ValidString(out) {
			t.Fatalf("sanitize(%q) = %q still holds a rune the terminal acts on", s, out)
		}
		if utf8.ValidString(s) && !strings.ContainsFunc(s, unsafeToDraw) && out != s {
			t.Fatalf("sanitize(%q) = %q changed a value that was safe", s, out)
		}
		if Sanitize(out) != out {
			t.Fatalf("sanitize is not idempotent on %q", s)
		}
	})
}

func BenchmarkSanitize(b *testing.B) {
	clean := strings.Repeat("a plain line of server text ", 40)
	dirty := strings.Repeat("evil \x1b]52;c;aGk=\x07 text \u202e ", 40)
	b.Run("clean", func(b *testing.B) {
		for b.Loop() {
			_ = Sanitize(clean)
		}
	})
	b.Run("dirty", func(b *testing.B) {
		for b.Loop() {
			_ = Sanitize(dirty)
		}
	})
}
