package palette

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/legend"
)

func TestTheLegendNamesThePalettesKeys(t *testing.T) {
	line := ansi.Strip(legend.HintLine(open().Legend(), legend.Style{}))
	require.Equal(t, "[enter] run · [tab] complete · [↑/↓] choose · [esc] cancel", line)
}

func TestThePaletteAnswersToTheKeysItWasGiven(t *testing.T) {
	keys := DefaultKeys()
	keys.Up = key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p/n", "choose"))
	keys.Down = key.NewBinding(key.WithKeys("ctrl+n"))
	m := open().WithKeys(keys)

	m, _ = press(m, tea.KeyDown)
	require.Equal(t, -1, m.cursor, "the arrow is no longer bound")
	o, _ := m.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	require.Equal(t, 0, o.(Model).cursor)
	require.Contains(t, ansi.Strip(legend.HintLine(m.Legend(), legend.Style{})), "[ctrl+p/n] choose")
}
