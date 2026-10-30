package draw

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"
)

func TestEveryKindHasItsOwnFixedColour(t *testing.T) {
	require.Equal(t, lipgloss.Color("#3B82F6"), Info.Color())
	require.Equal(t, lipgloss.Color("#22C55E"), Success.Color())
	require.Equal(t, lipgloss.Color("#EAB308"), Warning.Color())
	require.Equal(t, lipgloss.Color("#EF4444"), Danger.Color())
	require.Equal(t, Info.Color(), Kind(99).Color(), "out of range reads as Info")
}
