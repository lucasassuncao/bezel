package overlay

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

type pickedMsg struct{ i int }

func TestChoiceRunsTheOptionAndCloses(t *testing.T) {
	ran := ""
	c := NewChoice("Apply?", "one\ntwo", []Option{
		Opt("enter/y", "apply", func() tea.Msg { ran = "apply"; return nil }),
		Opt("d", "dry-run", func() tea.Msg { ran = "dry"; return nil }),
	}, theme.Resolved{})
	require.Contains(t, ansi.Strip(c.View(layout.Rect{W: 60, H: 20})), "dry-run (d)", "each option is a button naming its key")

	_, r := c.Update(press("d"))
	cmd := r.Cmd
	msgs := drain(cmd())
	require.Equal(t, "dry", ran)
	require.Contains(t, msgs, tea.Msg(CloseMsg{}))

	_, r = c.Update(press("z"))
	cmd = r.Cmd
	require.Nil(t, cmd, "an unbound key does nothing")
}

func TestPickMovesAndPicks(t *testing.T) {
	p := NewPick("Which?", []string{"winget", "scoop"},
		func(i int) tea.Cmd { return func() tea.Msg { return pickedMsg{i} } }, theme.Resolved{})
	down := tea.KeyPressMsg{Code: tea.KeyDown}
	o, _ := p.Update(down)
	p = o.(Pick)
	require.Equal(t, 1, p.Cursor())
	o, _ = p.Update(down)
	p = o.(Pick)
	require.Equal(t, 1, p.Cursor(), "clamped")
	require.Contains(t, ansi.Strip(p.View(layout.Rect{W: 60, H: 20})), "▸ scoop")

	_, r := p.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	cmd := r.Cmd
	require.Contains(t, drain(cmd()), tea.Msg(pickedMsg{1}))
}

// Only the arrows move: letters stay free for the app.
func TestPickIgnoresVimKeys(t *testing.T) {
	p := NewPick("Which?", []string{"a", "b"}, nil, theme.Resolved{})
	for _, k := range []string{"j", "k"} {
		o, _ := p.Update(press(k))
		require.Equal(t, 0, o.(Pick).Cursor(), k)
	}
}

func TestPickWithNoRowsKeepsTheCursorAtZero(t *testing.T) {
	p := NewPick("Which?", nil, nil, theme.Resolved{})
	o, _ := p.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, 0, o.(Pick).Cursor())
}

func TestPickWithCursorStartsThereAndClamps(t *testing.T) {
	p := NewPick("Which?", []string{"a", "b", "c"}, nil, theme.Resolved{})
	require.Equal(t, 1, p.WithCursor(1).Cursor())
	require.Equal(t, 2, p.WithCursor(9).Cursor())
	require.Equal(t, 0, p.WithCursor(-1).Cursor())
	empty := NewPick("Which?", nil, nil, theme.Resolved{})
	require.Equal(t, 0, empty.WithCursor(3).Cursor())
}

// rowsOf is "row 01" to "row n", so no row's name contains another's.
func rowsOf(n int) []string {
	rows := make([]string, n)
	for i := range rows {
		rows[i] = fmt.Sprintf("row %02d", i+1)
	}
	return rows
}

// A Pick taller than the body scrolls: the cursor row stays on screen and the
// rows under it are counted, never cut out of the middle.
func TestPickScrollsALongList(t *testing.T) {
	body := layout.Rect{W: 60, H: 16}
	o := NewPick("Which?", rowsOf(30), nil, theme.Resolved{}).Resize(body)
	for range 20 {
		o, _ = o.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	v := ansi.Strip(o.View(body))
	require.Contains(t, v, "▸ row 21")
	require.Contains(t, v, "↓ 9 more")
	require.NotContains(t, v, "resize to see them")
	require.LessOrEqual(t, strings.Count(v, "\n")+1, body.H)
}

// Back up inside the window it stays put: Update scrolls with the rows the
// last Resize allowed, rather than View pinning the cursor to the bottom.
func TestPickKeepsItsWindowWhileTheCursorMovesInside(t *testing.T) {
	body := layout.Rect{W: 60, H: 16}
	o := NewPick("Which?", rowsOf(30), nil, theme.Resolved{}).Resize(body)
	for range 20 {
		o, _ = o.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	for range 3 {
		o, _ = o.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	}
	v := ansi.Strip(o.View(body))
	require.Contains(t, v, "▸ row 18")
	require.Contains(t, v, "row 21", "the window moved with the cursor")
}

// A cursor started deep in the list is on screen, sized or not.
func TestPickStartsWithItsCursorInView(t *testing.T) {
	body := layout.Rect{W: 60, H: 16}
	p := NewPick("Which?", rowsOf(30), nil, theme.Resolved{}).WithCursor(25)
	require.Contains(t, ansi.Strip(p.View(body)), "▸ row 26")
	require.Contains(t, ansi.Strip(p.Resize(body).View(body)), "▸ row 26")
}

// Pick has no buttons, so it names its keys inside the box.
func TestPickShowsItsKeysInItsBox(t *testing.T) {
	p := NewPick("t", []string{"a", "b"}, nil, theme.Resolved{})
	require.Contains(t, ansi.Strip(p.View(layout.Rect{W: 80, H: 20})), "[enter] pick")
}

func TestTextClosesOnAnyKey(t *testing.T) {
	x := NewText("Help", "a line", theme.Resolved{})
	require.Contains(t, ansi.Strip(x.View(layout.Rect{W: 40, H: 10})), "a line")
	_, r := x.Update(press("q"))
	cmd := r.Cmd
	require.IsType(t, CloseMsg{}, cmd())
}

func TestTextLinesReadCRLFAndDropTrailingBlanks(t *testing.T) {
	require.Equal(t, []string{"a", "", "b"}, textLines("a\r\n\nb\n\n"))
	require.Nil(t, textLines(""))
	require.Nil(t, textLines("\n\n"))
}

func TestPickAnswersToTheKeysItWasGiven(t *testing.T) {
	keys := DefaultPickKeys()
	keys.Up, keys.Down = bind("k", "move"), key.New("j")
	p := NewPick("t", []string{"a", "b"}, nil, theme.Resolved{}).WithKeys(keys)

	o, _ := p.Update(press("j"))
	require.Equal(t, 1, o.(Pick).Cursor())
	o, _ = o.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	require.Equal(t, 1, o.(Pick).Cursor(), "the arrow is no longer bound")
	require.Contains(t, ansi.Strip(p.View(layout.Rect{W: 60, H: 20})), "[k] move")
}

func TestChoiceRunsTheOptionClicked(t *testing.T) {
	ran := ""
	opt := func(label, k string) Option {
		return Opt(k, label, func() tea.Msg { ran = label; return nil })
	}
	s := Stack{}.Push(NewChoice("Save?", "unsaved", []Option{opt("Save", "s"), opt("Discard", "d")}, theme.Resolved{}))
	x, y := cellOf(t, s, "Discard")
	_, r := s.Click(x, y, clickBody)
	cmd := r.Cmd
	require.True(t, isClose(cmd))
	require.Equal(t, "Discard", ran)
}
