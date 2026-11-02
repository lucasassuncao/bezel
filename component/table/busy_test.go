package table

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/spinner"
)

func busyItems() []Item[int] {
	items := pkgs(3)
	items[0].Busy, items[2].Busy = true, true
	return items
}

func TestABusyRowShowsTheSpinnerWhereTheMarkerGoes(t *testing.T) {
	m := New(pkgCols, busyItems()).SetHeight(4)
	lines := viewLines(m)
	require.Equal(t, "⠋ pkg-0  v0.0", lines[1], "the spinner takes the cursor's marker")
	require.Equal(t, "  pkg-1  v0.1", lines[2])
	require.Equal(t, "⠋ pkg-2  v0.2", lines[3])

	m, r := m.Update(m.spin.FrameAt(time.Now()))
	require.True(t, r.Handled)
	require.Equal(t, "⠙ pkg-2  v0.2", viewLines(m)[3], "its own frame turns it")
	require.True(t, sel(m).Busy, "Selected keeps the mark")
}

func TestTheTableSpinnerFollowsTheItems(t *testing.T) {
	m := New(pkgCols, busyItems()).SetHeight(4)
	require.True(t, m.spin.Running())
	m = m.SetItems(pkgs(3))
	require.False(t, m.spin.Running())
	require.Nil(t, m.Tick())
	_, r := m.Update(m.spin.FrameAt(time.Now()))
	require.Nil(t, r.Cmd, "a stopped spinner asks for nothing")
}

func TestWithSpinnerAndTheTableBusyStyle(t *testing.T) {
	red := th.Accent.Bold(true)
	m := New(pkgCols, busyItems()).SetHeight(4).WithSpinner(spinner.Line()).
		WithStyles(func(s Styles) Styles { s.Busy = red; return s })
	require.Equal(t, "| pkg-2  v0.2", viewLines(m)[3])
	require.Contains(t, m.SetWidth(40).View(th), red.Render("|"))
}
