package form

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/tags"
	"github.com/lucasassuncao/bezel/theme"
)

func TestTagsTurnTheTextIntoChips(t *testing.T) {
	f := New(theme.Resolve(theme.Plain(), true), NewTags("Labels").WithTags([]string{"go"}), NewCheckbox("x"))
	f, _ = press(f, "t", "u", "i")
	f, cmd := press(f, "enter")
	require.Nil(t, cmd)
	require.Equal(t, []string{"go", "tui"}, f.Field(0).(Tags).Tags())
	require.Contains(t, ansi.Strip(f.SetWidth(60).View()), "[ go   tui  ")

	f, _ = press(f, "backspace")
	require.Equal(t, []string{"go"}, f.Field(0).(Tags).Tags())
	f, _ = press(f, "tab")
	require.Equal(t, 1, f.Focused(), "tab still moves on")
	require.Equal(t, "enter", New(theme.Resolve(theme.Plain(), true), NewTags("t")).Legend()[2].Help.Key)
	require.Equal(t, tags.DefaultKeys().Add.Help(), New(theme.Resolve(theme.Plain(), true), NewTags("t")).Legend()[2].Help)
}
