package tree

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/spinner"
	"github.com/lucasassuncao/bezel/theme"
)

// job is a leaf that may be working.
type job struct {
	busy bool
}

func (j job) TreeInfo() Info  { return Info{Leaf: true, Busy: j.busy} }
func (j job) Expand(bool) job { return j }
func jobs(busy ...bool) []job {
	out := make([]job, len(busy))
	for i, b := range busy {
		out[i] = job{busy: b}
	}
	return out
}

func TestBusyMarkTurnsWhileANodeIsBusy(t *testing.T) {
	th := theme.Resolve(theme.Plain(), true)
	m := New(jobs(false, true))
	require.True(t, m.spin.Running())
	require.Equal(t, "⠋", ansi.Strip(m.BusyMark(th)))

	m, r := m.Update(m.spin.FrameAt(time.Now()))
	require.True(t, r.Handled)
	require.Equal(t, "⠙", ansi.Strip(m.BusyMark(th)))

	m = m.SetNode(1, func(j *job) { j.busy = false })
	require.False(t, m.spin.Running(), "no busy node: it stops")
	require.Nil(t, m.Tick())
	m = m.SetNodes(jobs(true))
	require.True(t, m.spin.Running())
}

func TestBusyMarkTakesTheBusyStyleAndAnotherSpinner(t *testing.T) {
	th := theme.Resolve(theme.Mint(), true)
	red := th.Accent.Bold(true)
	m := New(jobs(true)).WithSpinner(spinner.Line()).WithStyles(func(s Styles) Styles { s.Busy = red; return s })
	require.Equal(t, red.Render("|"), m.BusyMark(th))
}

func TestTheTreeAnswersAsAMovingThing(t *testing.T) {
	m := New(jobs(true))
	require.True(t, m.Active())
	m, r := m.Update(m.FrameAt(time.Now()))
	require.True(t, r.Handled)
	_, done := m.Advance(time.Now())
	require.False(t, done)
	_, done = m.SetNodes(jobs(false)).Advance(time.Now())
	require.True(t, done, "no busy node: settled")
}
