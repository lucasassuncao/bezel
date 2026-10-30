package key

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func press(code rune, mod tea.KeyMod) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code, Mod: mod} }

func TestNewCarriesItsKeysAndHelp(t *testing.T) {
	b := New("ctrl+s", "f2").WithHelp("ctrl+s", "save")
	require.Equal(t, []string{"ctrl+s", "f2"}, b.Keys())
	require.Equal(t, Help{Key: "ctrl+s", Desc: "save"}, b.Help())
	require.True(t, b.Enabled())
}

func TestMatchesAnyKeyOfAnyBinding(t *testing.T) {
	save := New("ctrl+s")
	up := New("up", "k")
	require.True(t, Matches(press('s', tea.ModCtrl), save))
	require.True(t, Matches(press('k', 0), save, up), "the second binding")
	require.False(t, Matches(press('j', 0), save, up))
	require.False(t, Matches(press('s', 0)), "no bindings, no match")
}

func TestADisabledOrEmptyBindingNeverMatches(t *testing.T) {
	off := New("q")
	off.Disabled = true
	require.False(t, off.Enabled())
	require.False(t, Matches(press('q', 0), off))

	off.Disabled = false
	require.True(t, Matches(press('q', 0), off), "switched back on")

	require.False(t, New().Enabled(), "no keys: not enabled")
}

// A binding is a value: switching off a copy leaves the one it came from on,
// and a new key or help is a new binding.
func TestABindingIsAValue(t *testing.T) {
	keys := struct{ Save Binding }{Save: New("ctrl+s").WithHelp("ctrl+s", "save")}
	off := keys.Save
	off.Disabled = true
	require.False(t, off.Enabled())
	require.True(t, keys.Save.Enabled(), "the copy was switched off, not the field")

	keys.Save.Disabled = true
	require.False(t, Matches(press('s', tea.ModCtrl), keys.Save))
	require.Equal(t, Help{Key: "ctrl+s", Desc: "save"}, keys.Save.Help(), "switched off, its help stays")
}

func TestNewCopiesItsKeys(t *testing.T) {
	keys := []string{"x"}
	b := New(keys...)
	keys[0] = "y"
	require.Equal(t, []string{"x"}, b.Keys())
	require.False(t, New([]string{}...).Enabled(), "an empty list: not enabled")
}
