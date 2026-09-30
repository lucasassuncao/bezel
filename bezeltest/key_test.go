package bezeltest

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

// What Key builds must read back as the name it was given, since that
// string is what key.Matches compares against a binding.
func TestKeyReadsBackAsItsName(t *testing.T) {
	names := []string{
		"enter", "esc", "tab", "space", "backspace", "delete", "insert",
		"up", "down", "left", "right", "home", "end", "pgup", "pgdown",
		"f1", "f5", "f12",
		"a", "V", "?", ":", "/", "4", "+",
		"ctrl+s", "shift+tab", "alt+x", "ctrl+alt+shift+a", "shift+a", "ctrl++",
	}
	for _, n := range names {
		require.Equal(t, n, Key(n).String(), n)
	}
}

// A text input inserts what the message carries, so a printable key needs
// its Text, and a key with a modifier or a named key must have none.
func TestKeyCarriesTextOnlyForAPlainCharacter(t *testing.T) {
	require.Equal(t, "a", Key("a").Text)
	require.Equal(t, "V", Key("V").Text)
	require.Equal(t, " ", Key(" ").Text)
	require.Equal(t, " ", Key("space").Text)
	require.Equal(t, rune(tea.KeySpace), Key(" ").Code)
	for _, n := range []string{"enter", "tab", "ctrl+s", "shift+a", "alt+x"} {
		require.Empty(t, Key(n).Text, n)
	}
}

func TestKeyMatchesTheBindingItNames(t *testing.T) {
	for _, n := range []string{"ctrl+s", "shift+tab", "V", "enter", " "} {
		b := key.NewBinding(key.WithKeys(Key(n).String()))
		require.True(t, key.Matches(Key(n), b), n)
	}
	require.False(t, key.Matches(Key("v"), key.NewBinding(key.WithKeys("V"))))
}

// A typo in a test must fail loudly, not send some other key.
func TestKeyPanicsOnAnUnknownName(t *testing.T) {
	for _, n := range []string{"entr", "ctrl+", "hyperr+a", "ab", ""} {
		require.Panics(t, func() { Key(n) }, "%q", n)
	}
}
