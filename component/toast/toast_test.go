package toast

import (
	"os"
	"strings"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/draw"
)

const ms = time.Millisecond

// TestMain clears the motion switches, so a CI with TERM=dumb still tests motion.
func TestMain(m *testing.M) {
	_ = os.Setenv("BEZEL_NO_MOTION", "")
	_ = os.Setenv("BEZEL_MOTION", "")
	_ = os.Setenv("TERM", "xterm-256color")
	os.Exit(m.Run())
}

func TestAToastSlidesInStaysAndLeaves(t *testing.T) {
	m, cmd := New().WithWidth(20).WithTTL(time.Second).Push("saved", draw.Success)
	require.NotNil(t, cmd)
	start := time.Unix(10, 0)
	m, _ = m.Advance(start)
	assert.Equal(t, 21, m.offset(m.items[0]), "it arrives from beyond the edge")
	m, _ = m.Advance(start.Add(500 * ms))
	assert.Equal(t, 0, m.offset(m.items[0]), "in place")
	assert.Contains(t, ansi.Strip(m.View()), "saved")
	leaving, _ := m.Advance(start.Add(900 * ms))
	assert.Positive(t, leaving.offset(leaving.items[0]))
	gone, _ := m.Advance(start.Add(1100 * ms))
	assert.Zero(t, gone.Len())
}

func TestToastsStackNewestFirst(t *testing.T) {
	m, _ := New().WithWidth(20).Push("first", draw.Info)
	m, cmd := m.Push("second", draw.Danger)
	assert.Nil(t, cmd, "one chain of frames serves the stack")
	m, _ = m.Advance(time.Unix(10, 0))
	m, _ = m.Advance(time.Unix(11, 0))
	out := ansi.Strip(m.View())
	assert.Less(t, strings.Index(out, "second"), strings.Index(out, "first"))
	for _, l := range strings.Split(out, "\n") {
		assert.Equal(t, 20, ansi.StringWidth(l))
	}
}

func TestOverDrawsOnTheTopRightCorner(t *testing.T) {
	screen := strings.TrimSuffix(strings.Repeat(strings.Repeat(".", 40)+"\n", 8), "\n")
	m, _ := New().WithWidth(16).Push("hi", draw.Info)
	m, _ = m.Advance(time.Unix(10, 0))
	m, _ = m.Advance(time.Unix(11, 0))
	lines := strings.Split(ansi.Strip(m.Over(screen)), "\n")
	require.Len(t, lines, 8)
	assert.Equal(t, strings.Repeat(".", 40), lines[0], "a row is left above it")
	assert.True(t, strings.HasSuffix(lines[2], "│ hi           │."), "%q", lines[2])
	assert.True(t, strings.HasPrefix(lines[2], "......."), "the screen shows beside it")
	for _, l := range lines {
		assert.Equal(t, 40, ansi.StringWidth(l))
	}
}

func TestTheStackTakesOnlyItsOwnFrames(t *testing.T) {
	m, _ := New().Push("a", draw.Info)
	other, _ := New().Push("b", draw.Info)
	_, cmd := withCmd(m.Update(other.FrameAt(time.Unix(1, 0))))
	assert.Nil(t, cmd)
	m, cmd = withCmd(m.Update(m.FrameAt(time.Unix(1, 0))))
	assert.NotNil(t, cmd)
	m, cmd = withCmd(m.Update(m.FrameAt(time.Unix(30, 0))))
	assert.Nil(t, cmd, "every toast is gone: no more frames")
	assert.Zero(t, m.Len())
}

func TestStillToastsDoNotSlideAndStillExpire(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	m, cmd := New().WithTTL(10*ms).Push("now", draw.Warning)
	require.NotNil(t, cmd)
	now, _ := m.Advance(time.Unix(10, 0))
	assert.Equal(t, 0, now.offset(now.items[0]))
	m, _ = m.Update(cmd())
	assert.Zero(t, m.Len())
}

// With motion off no frame ever stamps a toast, and it still shows.
func TestAStillToastShowsWithoutAFrame(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	m, _ := New().Push("saved", draw.Success)
	assert.Contains(t, ansi.Strip(m.View()), "saved")
}

func TestAZeroToastWorks(t *testing.T) {
	var m Model
	m, cmd := m.Push("x", draw.Info)
	assert.NotNil(t, cmd)
	assert.Equal(t, 1, m.Len())
}

func TestOwnFrameIsHandledAndAnotherIsNot(t *testing.T) {
	m, _ := New().Push("a", draw.Info)
	other, _ := New().Push("b", draw.Info)
	_, r := m.Update(m.FrameAt(time.Unix(1, 0)))
	assert.True(t, r.Handled)
	assert.NotNil(t, r.Cmd)
	_, r = m.Update(other.FrameAt(time.Unix(1, 0)))
	assert.False(t, r.Handled)
}

// withCmd keeps the command of an Update, for the tests that run it.
func withCmd[T any](m T, r bezel.Result) (T, tea.Cmd) { return m, r.Cmd }

// A toast's edge is its Kind's fixed colour: no theme reaches it.
func TestTheEdgeIsTheKindsFixedColour(t *testing.T) {
	m, _ := New().WithWidth(20).Push("failed", draw.Danger)
	require.Contains(t, m.View(), "38;2;239;68;68", "Danger's #EF4444")
}

// Toasts drawn over a screen leave it the size it was, at any point of
// their slide.
func FuzzOverKeepsTheScreenSize(f *testing.F) {
	f.Add("Saved notes.md", 80, 24, 100)
	f.Add("全角の通知", 10, 3, 0)
	f.Add("", 1, 1, 999)
	f.Add("x", 10, 3, 300) // a screen narrower than the toast
	f.Fuzz(func(t *testing.T, text string, w, h, ms int) {
		if !utf8.ValidString(text) || strings.ContainsFunc(text, unicode.IsControl) {
			t.Skip()
		}
		w, h, ms = 1+(w%120+120)%120, 1+(h%40+40)%40, (ms%3000+3000)%3000
		row := strings.Repeat(".", w)
		screen := strings.TrimSuffix(strings.Repeat(row+"\n", h), "\n")
		m, _ := New().WithWidth(30).Push(text, draw.Info)
		m, _ = m.Advance(time.Unix(0, 0))
		m, _ = m.Advance(time.Unix(0, 0).Add(time.Duration(ms) * time.Millisecond))
		lines := strings.Split(m.Over(screen), "\n")
		require.Len(t, lines, h)
		for i, l := range lines {
			require.Equal(t, w, ansi.StringWidth(l), "row %d: %q", i, l)
		}
	})
}
