package accordion

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"os"
	"strings"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Plain(), true)

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
	m, _ = m.Advance(start)
	m, _ = m.Advance(start.Add(d))
	return m
}

func TestASectionOpensALineAtATime(t *testing.T) {
	m := New(sections())
	assert.Equal(t, "▸ Install\n▸ Usage", ansi.Strip(m.SetWidth(20).View(th)))
	m, cmd := withCmd(m.Update(bezeltest.Key("enter")))
	require.NotNil(t, cmd)
	assert.True(t, m.IsOpen(0))
	half := ansi.Strip(opened(m, 100*time.Millisecond).SetWidth(20).View(th))
	assert.Equal(t, "▾ Install\n  one\n  two\n▸ Usage", half)
	full := opened(m, time.Second)
	assert.False(t, full.Active())
	assert.Len(t, strings.Split(full.SetWidth(20).View(th), "\n"), 6)
}

func TestASectionClosesFromWhereItIs(t *testing.T) {
	m, _ := New(sections()).Toggle(0)
	m = opened(m, 100*time.Millisecond)
	m, _ = m.Toggle(0)
	assert.False(t, m.IsOpen(0))
	assert.InDelta(t, 0.5, m.folds[0].share(), 1e-9, "it starts closing from halfway")
	shut := opened(m, time.Second)
	assert.Equal(t, "▸ Install\n▸ Usage", ansi.Strip(shut.SetWidth(20).View(th)))
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
	assert.Equal(t, "▸ Install\n▾ Usage\n  run it", ansi.Strip(m.SetWidth(20).View(th)))
	_, cmd = m.Toggle(9)
	assert.Nil(t, cmd)
}

func TestTheSectionsTakeOnlyTheirOwnFrames(t *testing.T) {
	m, own := New(sections()).WithKeys(DefaultKeys()).Toggle(0)
	_, other := New(sections()).Toggle(0)
	_, cmd := withCmd(m.Update(other()))
	assert.Nil(t, cmd)
	_, cmd = withCmd(m.Update(own()))
	assert.NotNil(t, cmd)
}

func TestAdvanceSettlesEverySection(t *testing.T) {
	m, _ := New(sections()).Toggle(0)
	m, _ = m.Toggle(1)
	m, done := m.Advance(time.Unix(10, 0))
	assert.False(t, done)
	m, done = m.Advance(time.Unix(11, 0))
	assert.True(t, done)
	assert.True(t, m.IsOpen(0) && m.IsOpen(1))
}

func TestWithStylesRecoloursTheHeadings(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m := New([]Section{{Title: "a"}, {Title: "b"}}).WithStyles(func(s Styles) Styles { s.Heading = red; return s })
	require.Equal(t, red.Render("▸ b"), strings.Split(m.SetWidth(20).View(th), "\n")[1])
}

func TestToggleAndMovesAreHandled(t *testing.T) {
	m := New(sections())
	for _, k := range []string{"enter", "down", "up"} {
		_, r := m.Update(bezeltest.Key(k))
		assert.True(t, r.Handled, k)
	}
	_, r := m.Update(bezeltest.Key("x"))
	assert.False(t, r.Handled)
}

func TestOwnFrameIsHandledAndAnotherIsNot(t *testing.T) {
	m, own := New(sections()).Toggle(0)
	_, other := New(sections()).Toggle(0)
	_, r := m.Update(own())
	assert.True(t, r.Handled)
	_, r = m.Update(other())
	assert.False(t, r.Handled)
}

// withCmd keeps the command of an Update, for the tests that run it.
func withCmd[T any](m T, r bezel.Result) (T, tea.Cmd) { return m, r.Cmd }

// Open or shut, every row of the accordion fits the width it was given.
func FuzzAccordionStaysInItsWidth(f *testing.F) {
	f.Add("Question", "An answer that runs on past the width.", 20, true)
	f.Add("全角の題", "本文\n二行目", 6, true)
	f.Add("", "", 1, false)
	f.Add("", "0", 1, true) // the body's indent alone outgrew one cell
	f.Fuzz(func(t *testing.T, title, body string, w int, open bool) {
		if !appText(title, body) {
			t.Skip()
		}
		w %= 80
		m := New([]Section{{Title: title, Body: body}, {Title: "second", Body: body}}).SetWidth(w)
		if open {
			m, _ = m.Toggle(0)
			m = opened(m, time.Hour)
		}
		for i, l := range strings.Split(m.View(th), "\n") {
			require.LessOrEqual(t, ansi.StringWidth(l), max(w, 0), "row %d", i)
		}
	})
}

// appText reports text an app hands a component as it is: valid UTF-8 with
// no control characters but line breaks. Anything else goes through
// draw.Sanitize first, the app's job.
func appText(ss ...string) bool {
	for _, s := range ss {
		if !utf8.ValidString(s) || strings.ContainsFunc(s, func(r rune) bool { return unicode.IsControl(r) && r != '\n' }) {
			return false
		}
	}
	return true
}
