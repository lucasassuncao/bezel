package list

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/spinner"
)

func busyRows() []Row[int] {
	return []Row[int]{{Label: "api", Mark: "●"}, {Label: "web", Mark: "●", Busy: true}, {Label: "db", Busy: true}}
}

func plainLines(m Model[int]) []string {
	return strings.Split(ansi.Strip(m.View(th)), "\n")
}

func TestABusyRowShowsTheSpinnerInPlaceOfItsMark(t *testing.T) {
	m := New(busyRows()).SetHeight(3)
	require.True(t, m.spin.Running())
	lines := plainLines(m)
	require.Equal(t, "▶ ●  api", lines[0])
	require.Equal(t, "  ⠋  web", lines[1], "the mark gives way")
	require.Equal(t, "⠋ db", lines[2], "no mark: the lead does")

	m, r := m.Update(m.spin.FrameAt(time.Now()))
	require.True(t, r.Handled)
	require.Equal(t, "  ⠙  web", plainLines(m)[1], "its own frame turns it")
}

func TestTheSpinnerStopsWhenNoRowIsBusy(t *testing.T) {
	m := New(busyRows()).SetHeight(3)
	m = m.SetRows([]Row[int]{{Label: "api"}})
	require.False(t, m.spin.Running())
	m = m.SetRow(0, Row[int]{Label: "api", Busy: true})
	require.True(t, m.spin.Running(), "marking a row starts it")
}

func TestSetRowsRestartsTheSpinnerSoOneTickTurnsIt(t *testing.T) {
	m := New(busyRows()).SetHeight(3)
	stale := m.spin.FrameAt(time.Now())
	m = m.SetRows(busyRows())
	m, r := m.Update(stale)
	require.True(t, r.Handled, "a spent frame is still the list's")
	require.Nil(t, r.Cmd, "but asks for no other")
}

func TestWithSpinnerAndTheBusyStyle(t *testing.T) {
	m := New(busyRows()).SetHeight(3).WithSpinner(spinner.Line())
	require.Equal(t, "| db", plainLines(m)[2])
	red := th.Accent.Bold(true)
	m = m.WithStyles(func(s Styles) Styles { s.Busy = red; return s })
	require.Contains(t, m.View(th), red.Render("|"))
}

func TestABusyRowKeepsItsDescriptionUnderTheLabel(t *testing.T) {
	m := New([]Row[int]{{Label: "web", Mark: "●", Busy: true, Description: "deploying"}}).SetHeight(2)
	require.Equal(t, []string{"▶ ⠋  web", "     deploying"}, plainLines(m))
}
