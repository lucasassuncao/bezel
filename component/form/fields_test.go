package form

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/theme"
)

func TestInputTakesWhatIsTyped(t *testing.T) {
	f, _ := press(sample(t), "a", "p", "p")
	require.Equal(t, "app", f.Field(0).(Input).Value())
}

func TestSelectHoldsTheKeysWhileOpen(t *testing.T) {
	f, _ := press(sample(t), "tab", "enter")
	require.True(t, f.Field(1).(Select).Holding())
	require.Contains(t, ansi.Strip(f.SetWidth(60).View()), "alpine", "the options open under it")

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
	for _, fl := range []Field{NewSelect("Image", nil), NewRadio("Shell", nil)} {
		f := New(th(), fl)
		require.NotPanics(t, func() {
			f, _ = press(f, "enter", "down", "up", "right", "left", "enter")
			_ = f.SetWidth(40).View()
		})
	}
	require.Empty(t, NewSelect("Image", nil).Value())
	require.Empty(t, NewRadio("Shell", nil).Value())
}

// The environment only picks the default: a host that chose a set keeps it.
func TestWithIconsOverridesTheEnvironment(t *testing.T) {
	t.Setenv("BEZEL_ASCII", "1")
	r := NewRadio("Shell", []string{"bash"})
	require.Contains(t, ansi.Strip(r.View(false, 5, 40)), "(*)")
	require.Contains(t, ansi.Strip(r.WithIcons(icon.Unicode()).View(false, 5, 40)), "(•)")

	sw := NewToggle("Docker").WithOn(true).WithIcons(icon.Unicode())
	require.Contains(t, ansi.Strip(sw.View(false, 6, 40)), "──●")
}

func TestInputCaretBlinksAtTheThemesPace(t *testing.T) {
	cmd := New(theme.Resolved{CaretBlink: time.Nanosecond}, NewInput("Name").WithWidth(10)).Init()
	require.NotNil(t, cmd, "Init starts the focused input's blink")
	start := time.Now()
	cmd()
	require.Less(t, time.Since(start), time.Second, "the blink came from the theme, not the half-second default")
}

func TestInputMaskHidesTheValue(t *testing.T) {
	in := NewInput("Token").WithWidth(10).WithValue("s3cret").WithMask('*')
	view := ansi.Strip(in.View(false, 6, 40))
	require.Contains(t, view, "******")
	require.NotContains(t, view, "s3cret")
	require.Equal(t, "s3cret", in.Value())
}

func TestEachControlTakesItsOwnKeys(t *testing.T) {
	keys := DefaultKeys()
	keys.Right, keys.Down, keys.Close = binding("n"), binding("ctrl+j"), binding("q")

	r, _ := NewRadio("r", []string{"a", "b"}).WithKeys(keys).Update(bezelKey("right"))
	require.Equal(t, 0, r.(Radio).Selected())
	r, _ = r.Update(bezelKey("n"))
	require.Equal(t, 1, r.(Radio).Selected())

	s := Field(NewSelect("s", []string{"a", "b"}).WithKeys(keys))
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

func TestWithsOpenTheControlsOnSavedValues(t *testing.T) {
	require.True(t, NewCheckbox("a").WithChecked(true).Checked())
	require.True(t, NewToggle("a").WithOn(true).On())
	require.Equal(t, "zsh", NewRadio("r", []string{"bash", "zsh"}).WithSelected(1).Value())
	require.Equal(t, 1, NewSelect("s", []string{"a", "b"}).WithSelected(9).Selected(), "clamped to the last")
	require.Equal(t, 0, NewSelect("s", nil).WithSelected(3).Selected(), "no options: 0")

	in := NewInput("Name").WithPlaceholder("my-project").WithWidth(12)
	require.Contains(t, ansi.Strip(in.View(false, 4, 40)), "my-project")
}
