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

// What a terminal only draws comes through Clean untouched: text, line
// breaks, tabs, colours and links.
func TestCleanKeepsWhatOnlyDraws(t *testing.T) {
	for _, s := range []string{"", "plain", "日本語 👩‍💻 é", "a\tb\nc", "\x1b[31mred\x1b[0m", "\x1b[38;2;1;2;3mrgb\x1b[m",
		"\x1b]8;;https://x.dev\x1b\\link\x1b]8;;\x1b\\", "° · « » ©"} {
		if got := Clean(s); got != s {
			t.Errorf("Clean(%q) = %q, want it untouched", s, got)
		}
	}
}

// A control that moves the caret or rings the terminal draws nothing, yet
// measures 0 cells: left in, a carriage return wipes the start of the row.
func TestCleanDropsControls(t *testing.T) {
	cases := map[string]string{"abc\rXY": "abcXY", "a\x07b": "ab", "a\bb": "ab", "a\x00b": "ab", "a\x7fb": "ab", "a\u009bb": "ab", "line\r\n": "line\n"}
	for in, want := range cases {
		if got := Clean(in); got != want {
			t.Errorf("Clean(%q) = %q, want %q", in, got, want)
		}
	}
}

// A whole sequence other than a colour or a link is a command to the
// terminal: the clipboard, the title, the screen, the keyboard.
func TestCleanDropsCommands(t *testing.T) {
	for _, cmd := range []string{"\x1b]52;c;aGk=\x07", "\x1b]0;title\x1b\\", "\x1b[2J", "\x1b[5A", "\x1b[>4;1m", "\x1bP$qm\x1b\\", "\x1b_Gf=1\x1b\\", "\x1bc"} {
		if got := Clean("a" + cmd + "b"); got != "ab" {
			t.Errorf("Clean(%q) = %q, want %q", "a"+cmd+"b", got, "ab")
		}
	}
}

// An escape cut short would swallow what follows it, padding and borders
// included, so it goes with whatever it already took.
func TestCleanDropsEscapesCutShort(t *testing.T) {
	cases := map[string]string{"a\x1b]0;title": "a", "a\x1b[38;2;1": "a", "a\x1b": "a", "a\x1b[": "a", "a\x1b[3\nb": "a\nb", "a\x1b]8;;x": "a"}
	for in, want := range cases {
		if got := Clean(in); got != want {
			t.Errorf("Clean(%q) = %q, want %q", in, got, want)
		}
	}
}

// A stray byte is drawn as U+FFFD, one cell, while ansi measures it as none.
func TestCleanReplacesInvalidUTF8(t *testing.T) {
	if got := Clean("a\xffb\xc3"); got != "a�b�" {
		t.Errorf("Clean = %q", got)
	}
}

// Whatever comes in, what Clean gives is valid UTF-8 whose only controls are
// tabs, line breaks and the escapes of whole colours and links.
func FuzzCleanLeavesOnlyColourAndLinks(f *testing.F) {
	for _, s := range []string{"", "plain", "a\x1b[31mb", "a\x1b]0;t", "\xff\x1b[", "a\rb\x07", "\x1b]8;;u\x1b\\l", "\u009b2J"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		out := Clean(s)
		if !utf8.ValidString(out) {
			t.Fatalf("Clean(%q) = %q is not UTF-8", s, out)
		}
		for rest := out; rest != ""; {
			r, size := utf8.DecodeRuneInString(rest)
			if r == 0x1b {
				n, keep := escape(rest)
				if !keep {
					t.Fatalf("Clean(%q) = %q keeps the escape %q", s, out, rest[:n])
				}
				size = n
			} else if r != '\t' && r != '\n' && (r < 0x20 || (r >= 0x7f && r <= 0x9f)) {
				t.Fatalf("Clean(%q) = %q keeps the control %U", s, out, r)
			}
			rest = rest[size:]
		}
		if Clean(out) != out {
			t.Fatalf("Clean is not idempotent on %q", s)
		}
	})
}

func BenchmarkClean(b *testing.B) {
	plain := strings.Repeat("a plain line of server text ", 10)
	styled := strings.Repeat("\x1b[1;34mstyled\x1b[m text ", 10)
	b.Run("plain", func(b *testing.B) {
		for b.Loop() {
			_ = Clean(plain)
		}
	})
	b.Run("styled", func(b *testing.B) {
		for b.Loop() {
			_ = Clean(styled)
		}
	})
}
