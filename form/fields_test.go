package form

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/icon"
	"github.com/lucasassuncao/bezel/theme"
)

func TestInputTakesWhatIsTyped(t *testing.T) {
	f, _ := press(sample(t), "a", "p", "p")
	require.Equal(t, "app", f.Field(0).(Input).Value())
}

func TestSelectHoldsTheKeysWhileOpen(t *testing.T) {
	f, _ := press(sample(t), "tab", "enter")
	require.True(t, f.Field(1).(Select).Holding())
	require.Contains(t, ansi.Strip(f.View(60)), "alpine", "the options open under it")

	f, _ = press(f, "tab")
	require.Equal(t, 1, f.Focused(), "tab does not leave an open select")

	f, _ = press(f, "down", "down", "enter")
	require.Equal(t, "alpine", f.Field(1).(Select).Value())
	require.False(t, f.Field(1).(Select).Holding())

	f, _ = press(f, "enter", "up", "esc")
	require.Equal(t, "alpine", f.Field(1).(Select).Value(), "esc closes without changing it")
}

func TestChoicesFlip(t *testing.T) {
	f, _ := press(sample(t), "tab", "tab", "right", "right", "right")
	require.Equal(t, "fish", f.Field(2).(Radio).Value(), "right stops at the last option")

	f, _ = press(f, "tab", "space")
	require.True(t, f.Field(3).(Toggle).On())

	f, _ = press(f, "tab", "space")
	require.True(t, f.Field(4).(Checkbox).Checked())
	f, _ = press(f, "enter")
	require.False(t, f.Field(4).(Checkbox).Checked())
}

func TestButtonRunsItsCommand(t *testing.T) {
	f, cmd := press(sample(t), "shift+tab", "enter")
	require.Equal(t, 5, f.Focused())
	require.NotNil(t, cmd)
	require.Equal(t, submitted{}, cmd())
}

func TestEmptyChoicesDoNotPanic(t *testing.T) {
	for _, fl := range []Field{NewSelect("Image", nil, 0, th()), NewRadio("Shell", nil, 0, th())} {
		f, _ := New(fl)
		require.NotPanics(t, func() {
			f, _ = press(f, "enter", "down", "up", "right", "left", "enter")
			_ = f.View(40)
		})
	}
	require.Empty(t, NewSelect("Image", nil, 0, th()).Value())
	require.Empty(t, NewRadio("Shell", nil, 0, th()).Value())
}

// The environment only picks the default: a host that chose a set keeps it.
func TestWithIconsOverridesTheEnvironment(t *testing.T) {
	t.Setenv("BEZEL_ASCII", "1")
	r := NewRadio("Shell", []string{"bash"}, 0, th())
	require.Contains(t, ansi.Strip(r.View(false, 5, 40)), "(*)")
	require.Contains(t, ansi.Strip(r.WithIcons(icon.Unicode()).View(false, 5, 40)), "(•)")

	sw := NewToggle("Docker", true, th()).WithIcons(icon.Unicode())
	require.Contains(t, ansi.Strip(sw.View(false, 6, 40)), "──●")
}

func TestInputCaretBlinksAtTheThemesPace(t *testing.T) {
	_, cmd := NewInput("Name", "", 10, theme.Resolved{CaretBlink: time.Nanosecond}).Focus()
	start := time.Now()
	cmd()
	require.Less(t, time.Since(start), time.Second, "the blink came from the theme, not the half-second default")
}

func TestInputMaskHidesTheValue(t *testing.T) {
	in := NewInput("Token", "", 10, th()).WithValue("s3cret").WithMask('*')
	view := ansi.Strip(in.View(false, 6, 40))
	require.Contains(t, view, "******")
	require.NotContains(t, view, "s3cret")
	require.Equal(t, "s3cret", in.Value())
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
