package input

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

// keys is each rune of s typed.
func keys(s string) []tea.Msg {
	var out []tea.Msg
	for _, r := range s {
		out = append(out, typed(string(r)))
	}
	return out
}

func TestFormatWritesTheLiteralsAndDropsWhatDoesNotFit(t *testing.T) {
	m := run(focused().WithFormat("9999-99-99"), keys("2026x11")...)
	require.Equal(t, "2026-11", m.Value(), "the dash written, the x dropped")
	require.False(t, m.Complete())
	m = run(m, keys("-0299")...)
	require.Equal(t, "2026-11-02", m.Value(), "a typed literal is taken, the rest cut")
	require.True(t, m.Complete())
}

func TestFormatHintsWhatIsLeft(t *testing.T) {
	m := run(focused().WithFormat("999.999.999-99"), keys("123")...)
	require.Equal(t, "123.___.___-__", strings.TrimRight(ansi.Strip(m.View()), " "))
}

func TestErasingStepsBackOverTheLiterals(t *testing.T) {
	m := run(focused().WithFormat("99:99"), keys("123")...)
	require.Equal(t, "12:3", m.Value())
	m = run(m, press(tea.KeyBackspace, 0))
	require.Equal(t, "12:", m.Value())
	m = run(m, press(tea.KeyBackspace, 0))
	require.Equal(t, "12", m.Value())
}

func TestFormatClassesAndEscapes(t *testing.T) {
	m := run(focused().WithFormat(`aa\9*`), keys("1ab2c")...)
	require.Equal(t, "ab92", m.Value(), "letters, a literal 9, then anything")
	require.Equal(t, "2026-11", New(th).WithFormat("9999-99").SetValue("202611").Value())
	require.True(t, New(th).SetValue("x").Complete(), "no format: any text")
}

func TestAPasteIsLaidOnTheFormat(t *testing.T) {
	m := run(focused().WithFormat("999.999.999.999"), tea.PasteMsg{Content: "192168001001"})
	require.Equal(t, "192.168.001.001", m.Value())
}
