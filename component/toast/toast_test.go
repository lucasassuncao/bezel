package toast

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/theme"
)

const ms = time.Millisecond

var th = theme.Resolve(theme.Plain(), true)

// TestMain clears the motion switches, so a CI with TERM=dumb still tests motion.
func TestMain(m *testing.M) {
	_ = os.Setenv("BEZEL_NO_MOTION", "")
	_ = os.Setenv("BEZEL_MOTION", "")
	_ = os.Setenv("TERM", "xterm-256color")
	os.Exit(m.Run())
}

func TestAToastSlidesInStaysAndLeaves(t *testing.T) {
	m, cmd := New().WithWidth(20).WithTTL(time.Second).Push("saved", Success)
	require.NotNil(t, cmd)
	start := time.Unix(10, 0)
	m, _ = m.Advance(start)
	assert.Equal(t, 21, m.offset(m.items[0]), "it arrives from beyond the edge")
	m, _ = m.Advance(start.Add(500 * ms))
	assert.Equal(t, 0, m.offset(m.items[0]), "in place")
	assert.Contains(t, ansi.Strip(m.View(th)), "saved")
	leaving, _ := m.Advance(start.Add(900 * ms))
	assert.Positive(t, leaving.offset(leaving.items[0]))
	gone, _ := m.Advance(start.Add(1100 * ms))
	assert.Zero(t, gone.Len())
}

func TestToastsStackNewestFirst(t *testing.T) {
	m, _ := New().WithWidth(20).Push("first", Info)
	m, cmd := m.Push("second", Error)
	assert.Nil(t, cmd, "one chain of frames serves the stack")
	m, _ = m.Advance(time.Unix(10, 0))
	m, _ = m.Advance(time.Unix(11, 0))
	out := ansi.Strip(m.View(th))
	assert.Less(t, strings.Index(out, "second"), strings.Index(out, "first"))
	for _, l := range strings.Split(out, "\n") {
		assert.Equal(t, 20, ansi.StringWidth(l))
	}
}

func TestOverDrawsOnTheTopRightCorner(t *testing.T) {
	screen := strings.TrimSuffix(strings.Repeat(strings.Repeat(".", 40)+"\n", 8), "\n")
	m, _ := New().WithWidth(16).Push("hi", Info)
	m, _ = m.Advance(time.Unix(10, 0))
	m, _ = m.Advance(time.Unix(11, 0))
	lines := strings.Split(ansi.Strip(m.Over(screen, th)), "\n")
	require.Len(t, lines, 8)
	assert.Equal(t, strings.Repeat(".", 40), lines[0], "a row is left above it")
	assert.True(t, strings.HasSuffix(lines[2], "│ hi           │."), "%q", lines[2])
	assert.True(t, strings.HasPrefix(lines[2], "......."), "the screen shows beside it")
	for _, l := range lines {
		assert.Equal(t, 40, ansi.StringWidth(l))
	}
}

func TestTheStackTakesOnlyItsOwnFrames(t *testing.T) {
	m, _ := New().Push("a", Info)
	other, _ := New().Push("b", Info)
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
	m, cmd := New().WithTTL(10*ms).Push("now", Warning)
	require.NotNil(t, cmd)
	now, _ := m.Advance(time.Unix(10, 0))
	assert.Equal(t, 0, now.offset(now.items[0]))
	m, _ = m.Update(cmd())
	assert.Zero(t, m.Len())
}

// With motion off no frame ever stamps a toast, and it still shows.
func TestAStillToastShowsWithoutAFrame(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	m, _ := New().Push("saved", Success)
	assert.Contains(t, ansi.Strip(m.View(theme.Resolve(theme.Plain(), true))), "saved")
}

func TestAZeroToastWorks(t *testing.T) {
	var m Model
	m, cmd := m.Push("x", Info)
	assert.NotNil(t, cmd)
	assert.Equal(t, 1, m.Len())
}

func TestOwnFrameIsHandledAndAnotherIsNot(t *testing.T) {
	m, _ := New().Push("a", Info)
	other, _ := New().Push("b", Info)
	_, r := m.Update(m.FrameAt(time.Unix(1, 0)))
	assert.True(t, r.Handled)
	assert.NotNil(t, r.Cmd)
	_, r = m.Update(other.FrameAt(time.Unix(1, 0)))
	assert.False(t, r.Handled)
}

// withCmd keeps the command of an Update, for the tests that run it.
func withCmd[T any](m T, r bezel.Result) (T, tea.Cmd) { return m, r.Cmd }
