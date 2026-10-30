package bezel

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestZeroResultIsNotHandled(t *testing.T) {
	var r Result
	require.False(t, r.Handled)
	require.Nil(t, r.Event)
	require.Nil(t, r.Cmd)
}

type counter int

func (c counter) Update(tea.Msg) (counter, Result) { return c + 1, Result{Handled: true} }

func TestModelIsSatisfiedByAValue(t *testing.T) {
	var m Model[counter] = counter(0)
	next, r := m.Update(nil)
	require.Equal(t, counter(1), next)
	require.True(t, r.Handled)
}

func TestClickMsgCarriesThePaneAndTheCell(t *testing.T) {
	c := ClickMsg{Pane: "list", MouseClickMsg: tea.MouseClickMsg{X: 2, Y: 3}}
	require.Equal(t, 3, c.Y)
}
