package accordion

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Plain, true)

// TestMain clears the motion switches, so a CI with TERM=dumb still tests motion.
func TestMain(m *testing.M) {
	_ = os.Setenv("BEZEL_NO_MOTION", "")
	_ = os.Setenv("BEZEL_MOTION", "")
	_ = os.Setenv("TERM", "xterm-256color")
	os.Exit(m.Run())
}

func sections() []Section {
	return []Section{
		{Title: "Install", Body: "one\ntwo\nthree\nfour"},
		{Title: "Usage", Body: "run it"},
	}
}

// opened is m's first section played from its first frame to d after it.
func opened(m Model, d time.Duration) Model {
	start := time.Unix(10, 0)
	m, _ = m.Update(m.FrameAt(0, start))
	m, _ = m.Update(m.FrameAt(0, start.Add(d)))
	return m
}

func TestASectionOpensALineAtATime(t *testing.T) {
	m := New(sections())
	assert.Equal(t, "▸ Install\n▸ Usage", ansi.Strip(m.View(th, 20)))
	m, cmd := m.Update(bezeltest.Key("enter"))
	require.NotNil(t, cmd)
	assert.True(t, m.IsOpen(0))
	half := ansi.Strip(opened(m, 100*time.Millisecond).View(th, 20))
	assert.Equal(t, "▾ Install\n  one\n  two\n▸ Usage", half)
	full := opened(m, time.Second)
	assert.False(t, full.Active())
	assert.Len(t, strings.Split(full.View(th, 20), "\n"), 6)
}

func TestASectionClosesFromWhereItIs(t *testing.T) {
	m, _ := New(sections()).Toggle(0)
	m = opened(m, 100*time.Millisecond)
	m, _ = m.Toggle(0)
	assert.False(t, m.IsOpen(0))
	assert.InDelta(t, 0.5, m.folds[0].share(), 1e-9, "it starts closing from halfway")
	shut := opened(m, time.Second)
	assert.Equal(t, "▸ Install\n▸ Usage", ansi.Strip(shut.View(th, 20)))
}

func TestTheArrowsMoveTheCursor(t *testing.T) {
	m := New(sections())
	m, _ = m.Update(bezeltest.Key("down"))
	m, _ = m.Update(bezeltest.Key("down"))
	assert.Equal(t, 1, m.Cursor(), "clamped at the last")
	m, _ = m.Update(bezeltest.Key("up"))
	assert.Equal(t, 0, m.Cursor())
	assert.NotEmpty(t, m.Legend())
}

func TestAStillAccordionSnaps(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	m, cmd := New(sections()).WithDuration(time.Second).Toggle(1)
	assert.Nil(t, cmd)
	assert.Equal(t, "▸ Install\n▾ Usage\n  run it", ansi.Strip(m.View(th, 20)))
	_, cmd = m.Toggle(9)
	assert.Nil(t, cmd)
}

func TestTheSectionsTakeOnlyTheirOwnFrames(t *testing.T) {
	m, _ := New(sections()).WithKeys(DefaultKeys()).Toggle(0)
	other, _ := New(sections()).Toggle(0)
	_, cmd := m.Update(other.FrameAt(0, time.Unix(1, 0)))
	assert.Nil(t, cmd)
	_, cmd = m.Update(m.FrameAt(0, time.Unix(1, 0)))
	assert.NotNil(t, cmd)
}
