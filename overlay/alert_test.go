package overlay

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

func TestAlertClosesOnAnyKey(t *testing.T) {
	a := NewAlert(draw.Info, "Saved", "written to disk", theme.Resolved{})
	_, r := a.Update(press("x"))
	cmd := r.Cmd
	require.NotNil(t, cmd)
	require.IsType(t, CloseMsg{}, cmd())
	require.Contains(t, ansi.Strip(a.View(layout.Rect{W: 60, H: 20})), "written to disk")
}
