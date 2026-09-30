package overlay

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
)

type pickedMsg struct{ i int }

func TestChoiceRunsTheOptionAndCloses(t *testing.T) {
	ran := ""
	c := NewChoice("Apply?", []string{"one", "two"}, []Option{
		Opt("enter/y", "apply", func() tea.Msg { ran = "apply"; return nil }),
		Opt("d", "dry-run", func() tea.Msg { ran = "dry"; return nil }),
	}, lipgloss.NewStyle(), legend.Style{})
	require.Contains(t, ansi.Strip(c.View(layout.Rect{W: 60, H: 20})), "dry-run (d)", "each option is a button naming its key")

	_, cmd := c.Update(press("d"))
	msgs := drain(cmd())
	require.Equal(t, "dry", ran)
	require.Contains(t, msgs, tea.Msg(CloseMsg{}))

	_, cmd = c.Update(press("z"))
	require.Nil(t, cmd, "an unbound key does nothing")
}

func TestPickMovesAndPicks(t *testing.T) {
	p := NewPick("Which?", []string{"winget", "scoop"}, nil,
		func(i int) tea.Cmd { return func() tea.Msg { return pickedMsg{i} } }, lipgloss.NewStyle(), legend.Style{})
	down := tea.KeyPressMsg{Code: tea.KeyDown}
	o, _ := p.Update(down)
	p = o.(Pick)
	require.Equal(t, 1, p.Cursor())
	o, _ = p.Update(down)
	p = o.(Pick)
	require.Equal(t, 1, p.Cursor(), "clamped")
	require.Contains(t, ansi.Strip(p.View(layout.Rect{W: 60, H: 20})), "▸ scoop")

	_, cmd := p.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Contains(t, drain(cmd()), tea.Msg(pickedMsg{1}))
}

// Only the arrows move: letters stay free for the app.
func TestPickIgnoresVimKeys(t *testing.T) {
	p := NewPick("Which?", []string{"a", "b"}, nil, nil, lipgloss.NewStyle(), legend.Style{})
	for _, k := range []string{"j", "k"} {
		o, _ := p.Update(press(k))
		require.Equal(t, 0, o.(Pick).Cursor(), k)
	}
}

func TestPickWithNoRowsKeepsTheCursorAtZero(t *testing.T) {
	p := NewPick("Which?", nil, nil, nil, lipgloss.NewStyle(), legend.Style{})
	o, _ := p.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, 0, o.(Pick).Cursor())
}

// Pick has no buttons, so it names its keys inside the box.
func TestPickShowsItsKeysInItsBox(t *testing.T) {
	p := NewPick("t", []string{"a", "b"}, nil, nil, lipgloss.NewStyle(), legend.Style{})
	require.Contains(t, ansi.Strip(p.View(layout.Rect{W: 80, H: 20})), "[enter] pick")
}

func TestTextClosesOnAnyKey(t *testing.T) {
	x := NewText("Help", []string{"a line"}, lipgloss.NewStyle(), legend.Style{})
	require.Contains(t, ansi.Strip(x.View(layout.Rect{W: 40, H: 10})), "a line")
	_, cmd := x.Update(press("q"))
	require.IsType(t, CloseMsg{}, cmd())
}
