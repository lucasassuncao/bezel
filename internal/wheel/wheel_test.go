package wheel

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestRowsReadsAVerticalTurn(t *testing.T) {
	for _, c := range []struct {
		msg  tea.Msg
		rows int
		ok   bool
	}{
		{tea.MouseWheelMsg{Button: tea.MouseWheelUp}, -1, true},
		{tea.MouseWheelMsg{Button: tea.MouseWheelDown}, 1, true},
		{tea.MouseWheelMsg{Button: tea.MouseWheelDown, Mod: tea.ModShift}, 0, false},
		{tea.MouseWheelMsg{Button: tea.MouseWheelLeft}, 0, false},
		{tea.KeyPressMsg{Code: tea.KeyDown}, 0, false},
	} {
		rows, ok := Rows(c.msg)
		require.Equal(t, c.rows, rows)
		require.Equal(t, c.ok, ok)
	}
}
