package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/theme"
)

var at = time.Date(2026, 10, 28, 13, 11, 21, 0, time.UTC)

func newTest() (Handler, *bytes.Buffer) {
	var buf bytes.Buffer
	return New(&buf, theme.Resolve(theme.Mint(), true)).WithIcons(icon.Unicode()), &buf
}

func handle(t *testing.T, h slog.Handler, buf *bytes.Buffer, l slog.Level, msg string, args ...any) string {
	t.Helper()
	buf.Reset()
	r := slog.NewRecord(at, l, msg, 0)
	r.Add(args...)
	require.NoError(t, h.Handle(context.Background(), r))
	return buf.String()
}

func TestLoggerPutsOneAttributeUnderTheMessage(t *testing.T) {
	h, buf := newTest()
	out := handle(t, h, buf, slog.LevelWarn, "[installers] 1 .msi file to move", "name", "go1.27.3.windows-amd64.msi")
	require.Equal(t, ""+
		"2026-10-28 13:11:21 WARN  [installers] 1 .msi file to move\n"+
		"                          └─ name=go1.27.3.windows-amd64.msi\n", out)
}

func TestLoggerBranchesEveryAttribute(t *testing.T) {
	h, buf := newTest()
	out := handle(t, h, buf, slog.LevelInfo, "run complete", "moved", 5, "size", "874.19 MB")
	require.Equal(t, ""+
		"2026-10-28 13:11:21 INFO  run complete\n"+
		"                          ├─ moved=5\n"+
		"                          └─ size=874.19 MB\n", out)
}

func TestLoggerNamesEveryLevel(t *testing.T) {
	h, buf := newTest()
	h = h.WithLevel(LevelTrace).WithTimeFormat("")
	for l, want := range map[slog.Level]string{
		LevelTrace: "TRACE", slog.LevelDebug: "DEBUG", slog.LevelInfo: "INFO ",
		slog.LevelWarn: "WARN ", slog.LevelError: "ERROR", slog.LevelError + 4: "ERROR",
	} {
		require.Equal(t, want+" hi\n", handle(t, h, buf, l, "hi"))
	}
}

func TestLoggerSkipsWhatIsBelowItsLevel(t *testing.T) {
	h, _ := newTest()
	require.False(t, h.Enabled(context.Background(), slog.LevelDebug))
	var lv slog.LevelVar
	h = h.WithLevel(&lv)
	lv.Set(slog.LevelDebug)
	require.True(t, h.Enabled(context.Background(), slog.LevelDebug))
}

func TestLoggerNestsGroups(t *testing.T) {
	h, buf := newTest()
	var sh slog.Handler = h.WithTimeFormat("")
	sh = sh.WithAttrs([]slog.Attr{slog.String("batch", "b1")}).WithGroup("run")
	sh = sh.WithAttrs([]slog.Attr{slog.Int("moved", 5)})
	out := handle(t, sh, buf, slog.LevelInfo, "done", slog.Group("files", "skipped", 0, "kept", 2), "took", "2s")
	require.Equal(t, ""+
		"INFO  done\n"+
		"      ├─ batch=b1\n"+
		"      └─ run\n"+
		"         ├─ moved=5\n"+
		"         ├─ files\n"+
		"         │  ├─ skipped=0\n"+
		"         │  └─ kept=2\n"+
		"         └─ took=2s\n", out)
}

// slog drops empty attributes and groups, and inlines a group with no key.
func TestLoggerFollowsSlogRules(t *testing.T) {
	h, buf := newTest()
	var sh slog.Handler = h.WithTimeFormat("")
	sh = sh.WithGroup("empty")
	out := handle(t, sh, buf, slog.LevelInfo, "hi", slog.Attr{}, slog.Group("none"))
	require.Equal(t, "INFO  hi\n", out)

	out = handle(t, h.WithTimeFormat(""), buf, slog.LevelInfo, "hi", slog.Group("", "a", 1))
	require.Equal(t, "INFO  hi\n      └─ a=1\n", out)
}

func TestLoggerWrapsALongValueWithoutLosingIt(t *testing.T) {
	h, buf := newTest()
	h = h.WithTimeFormat("").WithWidth(60)
	path := `C:\Users\lucas\Downloads\Installers\exe\tor-browser-windows-x86_64-portable-15.0.26.exe`
	out := handle(t, h, buf, slog.LevelInfo, "moved", "destination", path)
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	require.Equal(t, []string{
		"INFO  moved",
		`      └─ destination=C:\Users\lucas\Downloads\Installers\`,
		`                     exe\tor-browser-windows-x86_64-`,
		`                     portable-15.0.26.exe`,
	}, lines)
	for _, l := range lines {
		require.LessOrEqual(t, ansi.StringWidth(l), 60)
	}
}

func TestLoggerKeepsTheTreeAlongAWrappedValue(t *testing.T) {
	h, buf := newTest()
	h = h.WithTimeFormat("").WithWidth(30)
	out := handle(t, h, buf, slog.LevelInfo, "m", "a", "one two three four five", "b", 1)
	require.Equal(t, ""+
		"INFO  m\n"+
		"      ├─ a=one two three four \n"+
		"      │    five\n"+
		"      └─ b=1\n", out)
}

func TestLoggerAlignsAValuesOwnLines(t *testing.T) {
	h, buf := newTest()
	out := handle(t, h.WithTimeFormat(""), buf, slog.LevelError, "failed", "err", "first\nsecond")
	require.Equal(t, "ERROR failed\n      └─ err=first\n             second\n", out)
}

func TestLoggerWrapsTheMessageUnderItself(t *testing.T) {
	h, buf := newTest()
	out := handle(t, h.WithTimeFormat("").WithWidth(20), buf, slog.LevelInfo, "one two three four")
	require.Equal(t, "INFO  one two three\n      four\n", out)
}

func TestLoggerNeverWrapsBelowZero(t *testing.T) {
	h, buf := newTest()
	long := strings.Repeat("x", 300)
	out := handle(t, h.WithTimeFormat("").WithWidth(-1), buf, slog.LevelInfo, "m", "k", long)
	require.Equal(t, "INFO  m\n      └─ k="+long+"\n", out)
}

func TestLoggerIsPlainOffATerminal(t *testing.T) {
	h, buf := newTest()
	out := handle(t, h, buf, slog.LevelInfo, "\x1b[36m[images]\x1b[0m none", "k", "\x1b[31mv\x1b[0m")
	require.NotContains(t, out, "\x1b[")
	require.Contains(t, out, "[images] none")
}

func TestLoggerColoursWhenAsked(t *testing.T) {
	h, buf := newTest()
	out := handle(t, h.WithColor(true), buf, slog.LevelInfo, "m", "k", "v")
	require.Contains(t, out, "\x1b[")
	require.Equal(t, "2026-10-28 13:11:21 INFO  m\n                          └─ k=v\n", ansi.Strip(out))
}

func TestLoggerTakesStyles(t *testing.T) {
	h, buf := newTest()
	h = h.WithColor(true).WithStyles(func(s Styles) Styles {
		s.Key = s.Key.Underline(true)
		return s
	})
	require.Contains(t, handle(t, h, buf, slog.LevelInfo, "m", "k", "v"), "\x1b[4")
}

func TestLoggerDrawsInASCII(t *testing.T) {
	h, buf := newTest()
	out := handle(t, h.WithIcons(icon.ASCII()).WithTimeFormat(""), buf, slog.LevelInfo, "m", "a", 1, "b", 2)
	require.Equal(t, "INFO  m\n      |- a=1\n      `- b=2\n", out)
}

func TestLoggerNamesTheCaller(t *testing.T) {
	var buf bytes.Buffer
	h := New(&buf, theme.Resolve(theme.Mint(), true)).WithTimeFormat("").WithCaller(true)
	slog.New(h).Info("hi")
	require.Regexp(t, `^INFO  hi logger_test\.go:\d+\n$`, buf.String())
}

func TestLoggerFormatsTimeValues(t *testing.T) {
	h, buf := newTest()
	out := handle(t, h.WithTimeFormat(""), buf, slog.LevelInfo, "m", "at", at)
	require.Equal(t, "INFO  m\n      └─ at=2026-10-28T13:11:21Z\n", out)
}

func TestLoggerWritesWholeEntries(t *testing.T) {
	h, buf := newTest()
	log := slog.New(h.WithTimeFormat(""))
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() { log.Info("m", "a", 1, "b", 2) })
	}
	wg.Wait()
	require.Equal(t, strings.Repeat("INFO  m\n      ├─ a=1\n      └─ b=2\n", 20), buf.String())
}

func TestReplaceAttrNamesTrace(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: LevelTrace, ReplaceAttr: ReplaceAttr}))
	log.Log(context.Background(), LevelTrace, "deep", "k", "v")
	var got map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	require.Equal(t, "TRACE", got["level"])
	require.Equal(t, "v", got["k"])

	require.Equal(t, slog.Any(slog.LevelKey, slog.LevelInfo), ReplaceAttr(nil, slog.Any(slog.LevelKey, slog.LevelInfo)))
}

// No line the handler writes is wider than the width it was given, however
// long the message and the values, while a value has its minRoom.
func FuzzLinesStayInTheWidth(f *testing.F) {
	f.Add("moved files", "path", `C:\Users\someone\Downloads\a-rather-long-name.zip`, 40)
	f.Add("全角のメッセージ", "キー", "値の文字列がとても長いのでよく折り返される", 24)
	f.Add("", "", "", 30)
	f.Fuzz(func(t *testing.T, msg, k, v string, w int) {
		if !utf8.ValidString(msg+k+v) || strings.ContainsFunc(msg+k+v, unicode.IsControl) {
			t.Skip()
		}
		w = 30 + ansi.StringWidth(k) + (w%120+120)%120 // minRoom: a value always gets 12 cells
		var buf bytes.Buffer
		h := New(&buf, theme.Resolve(theme.Plain(), true)).WithWidth(w).WithTimeFormat("").WithColor(false)
		slog.New(h).Info(msg, k, v, "group", slog.GroupValue(slog.String(k, v)))
		for i, l := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
			require.LessOrEqual(t, ansi.StringWidth(l), w, "line %d: %q", i, l)
		}
	})
}
