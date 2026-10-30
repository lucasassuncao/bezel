package overlay

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

func TestConfirmRunsOnYesAndClosesOnNo(t *testing.T) {
	fired := false
	onYes := func() tea.Msg { fired = true; return nil }
	c := NewConfirm("Delete?", "this cannot be undone", onYes, theme.Resolved{})

	_, r := c.Update(press("n"))
	cmd := r.Cmd
	require.IsType(t, CloseMsg{}, cmd())
	require.False(t, fired)

	_, r = c.Update(press("y"))
	cmd = r.Cmd
	// The yes command is a batch of onYes then Close: run it the way tea would.
	msgs := drain(cmd())
	require.True(t, fired)
	require.Contains(t, msgs, tea.Msg(CloseMsg{}))
}

func TestConfirmIgnoresVimKeys(t *testing.T) {
	c := NewConfirm("Delete?", "", nil, theme.Resolved{})
	for _, k := range []string{"h", "l"} {
		o, _ := c.Update(press(k))
		require.False(t, o.(Confirm).no, k)
	}
	o, _ := c.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	require.True(t, o.(Confirm).no, "the arrows still move")
}

// The button enter would press is filled in the colour the dialog's kind gives its border.
func TestConfirmFillsItsFocusedButtonInItsKindsColour(t *testing.T) {
	th := theme.Resolve(theme.Mint(), true)
	c := NewConfirm("Delete?", "gone for good", nil, th).WithKind(Danger)
	view := c.View(layout.Rect{W: 80, H: 20})
	require.Contains(t, view, th.Button.Focused.Background(lipgloss.Color(th.Colors.Danger)).Render("  Yes  "))
	require.Contains(t, view, th.Button.Blurred.Render("  No  "))
}

func TestConfirmAnswersToTheKeysItWasGiven(t *testing.T) {
	keys := DefaultConfirmKeys()
	keys.Yes = bind("s", "sim")
	c := NewConfirm("Apagar?", "", func() tea.Msg { return nil }, theme.Resolved{}).WithKeys(keys)

	_, r := c.Update(press("y"))
	cmd := r.Cmd
	require.Nil(t, cmd, "y is no longer an answer")
	_, r = c.Update(press("s"))
	cmd = r.Cmd
	require.Contains(t, drain(cmd()), tea.Msg(CloseMsg{}))
	require.Contains(t, hint(c), "[s] sim")
}

func TestConfirmAnswersAClickOnItsButtons(t *testing.T) {
	fired := false
	c := NewConfirm("Delete?", "for good", func() tea.Msg { fired = true; return nil }, theme.Resolved{})
	s := Stack{}.Push(c)
	x, y := cellOf(t, s, "No")
	_, r := s.Click(x, y, clickBody)
	cmd := r.Cmd
	require.True(t, isClose(cmd))
	require.False(t, fired)
	x, y = cellOf(t, s, "Yes")
	_, r = s.Click(x, y, clickBody)
	cmd = r.Cmd
	require.True(t, isClose(cmd))
	require.True(t, fired)
}

func TestConfirmClickOnYesSendsItsCommand(t *testing.T) {
	fired := false
	c := NewConfirm("Delete?", "for good", func() tea.Msg { fired = true; return nil }, theme.Resolved{})
	s := Stack{}.Push(c)
	x, y := cellOf(t, s, "Yes")
	_, r := s.Click(x, y, clickBody)
	require.True(t, r.Handled)
	require.True(t, isClose(r.Cmd))
	require.True(t, fired)
}
