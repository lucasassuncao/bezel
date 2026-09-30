package overlay

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/legend"
)

func TestConfirmRunsOnYesAndClosesOnNo(t *testing.T) {
	fired := false
	onYes := func() tea.Msg { fired = true; return nil }
	c := NewConfirm("Delete?", "this cannot be undone", onYes, lipgloss.NewStyle(), legend.Style{})

	_, cmd := c.Update(press("n"))
	require.IsType(t, CloseMsg{}, cmd())
	require.False(t, fired)

	_, cmd = c.Update(press("y"))
	// The yes command is a batch of onYes then Close: run it the way tea would.
	msgs := drain(cmd())
	require.True(t, fired)
	require.Contains(t, msgs, tea.Msg(CloseMsg{}))
}

func TestConfirmIgnoresVimKeys(t *testing.T) {
	c := NewConfirm("Delete?", "", nil, lipgloss.NewStyle(), legend.Style{})
	for _, k := range []string{"h", "l"} {
		o, _ := c.Update(press(k))
		require.False(t, o.(Confirm).no, k)
	}
	o, _ := c.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	require.True(t, o.(Confirm).no, "the arrows still move")
}
