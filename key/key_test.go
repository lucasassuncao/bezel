package key

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func press(code rune, mod tea.KeyMod) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code, Mod: mod} }

func TestNewBindingCarriesItsKeysAndHelp(t *testing.T) {
	b := NewBinding(WithKeys("ctrl+s", "f2"), WithHelp("ctrl+s", "save"))
	require.Equal(t, []string{"ctrl+s", "f2"}, b.Keys())
	require.Equal(t, Help{Key: "ctrl+s", Desc: "save"}, b.Help())
	require.True(t, b.Enabled())
}

func TestMatchesAnyKeyOfAnyBinding(t *testing.T) {
	save := NewBinding(WithKeys("ctrl+s"))
	up := NewBinding(WithKeys("up", "k"))
	require.True(t, Matches(press('s', tea.ModCtrl), save))
	require.True(t, Matches(press('k', 0), save, up), "the second binding")
	require.False(t, Matches(press('j', 0), save, up))
	require.False(t, Matches(press('s', 0)), "no bindings, no match")
}

func TestADisabledOrEmptyBindingNeverMatches(t *testing.T) {
	off := NewBinding(WithKeys("q"), WithDisabled())
	require.False(t, off.Enabled())
	require.False(t, Matches(press('q', 0), off))

	off.SetEnabled(true)
	require.True(t, Matches(press('q', 0), off), "switched back on")

	require.False(t, NewBinding().Enabled(), "no keys: not enabled")
}

func TestSettersChangeTheBindingInPlace(t *testing.T) {
	b := NewBinding(WithKeys("a"), WithHelp("a", "add"))
	b.SetKeys("n")
	b.SetHelp("n", "new")
	require.Equal(t, []string{"n"}, b.Keys())
	require.Equal(t, "new", b.Help().Desc)

	b.Unbind()
	require.Nil(t, b.Keys())
	require.Equal(t, Help{}, b.Help())
	require.False(t, Matches(press('n', 0), b))
}
