package form

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
)

func binding(keys ...string) key.Binding { return key.NewBinding(key.WithKeys(keys...)) }

// One call rebinds the form's focus keys and the keys of every control in it.
func TestFormWithKeysRebindsTheFocusAndItsControls(t *testing.T) {
	keys := DefaultKeys()
	keys.Next, keys.Press = binding("ctrl+n"), binding("x")
	f, _ := New(NewCheckbox("a", false, th()), NewButton("go", func() tea.Msg { return submitted{} }, th()))
	f = f.WithKeys(keys)

	f, _ = press(f, "space")
	require.False(t, f.Field(0).(Checkbox).Checked(), "space is no longer a press")
	f, _ = press(f, "x")
	require.True(t, f.Field(0).(Checkbox).Checked())

	f, _ = press(f, "tab")
	require.Equal(t, 0, f.Focused(), "tab is no longer next")
	f, _ = press(f, "ctrl+n")
	require.Equal(t, 1, f.Focused())
	_, cmd := press(f, "x")
	require.Equal(t, submitted{}, cmd())
}

func TestEachControlTakesItsOwnKeys(t *testing.T) {
	keys := DefaultKeys()
	keys.Right, keys.Down, keys.Close = binding("n"), binding("ctrl+j"), binding("q")

	r, _ := NewRadio("r", []string{"a", "b"}, 0, th()).WithKeys(keys).Update(bezelKey("right"))
	require.Equal(t, 0, r.(Radio).Selected())
	r, _ = r.Update(bezelKey("n"))
	require.Equal(t, 1, r.(Radio).Selected())

	s := Field(NewSelect("s", []string{"a", "b"}, 0, th()).WithKeys(keys))
	for _, k := range []string{"enter", "down", "ctrl+j", "q"} {
		s, _ = s.Update(bezelKey(k))
	}
	require.False(t, s.(Select).Holding(), "q closes the list")
	require.Equal(t, 0, s.(Select).Selected(), "closing never picks")
	for _, k := range []string{"enter", "ctrl+j", "enter"} {
		s, _ = s.Update(bezelKey(k))
	}
	require.Equal(t, 1, s.(Select).Selected())
}

func bezelKey(k string) tea.KeyPressMsg { return bezeltest.Key(k) }
