package overlay

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
)

func TestAlertClosesOnAnyKey(t *testing.T) {
	a := NewAlert(Info, "Saved", "written to disk", lipgloss.NewStyle(), legend.Style{})
	_, cmd := a.Update(press("x"))
	require.NotNil(t, cmd)
	require.IsType(t, CloseMsg{}, cmd())
	require.Contains(t, ansi.Strip(a.View(layout.Rect{W: 60, H: 20})), "written to disk")
}
