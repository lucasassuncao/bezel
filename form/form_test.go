package form

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/theme"
)

type submitted struct{}

func th() theme.Resolved { return theme.Resolve(theme.ThemePlain, true) }

// sample is one of each control, in the order the tests index them.
func sample(t *testing.T) Form {
	t.Helper()
	f, _ := New(
		NewInput("Name", "my-project", 24, th()),
		NewSelect("Image", []string{"ubuntu", "debian", "alpine"}, 0, th()),
		NewRadio("Shell", []string{"bash", "zsh", "fish"}, 0, th()),
		NewToggle("Docker", false, th()),
		NewCheckbox("Forward port 3000", false, th()),
		NewButton("Create", func() tea.Msg { return submitted{} }, th()),
	)
	return f
}

func press(f Form, keys ...string) (Form, tea.Cmd) {
	var cmd tea.Cmd
	for _, k := range keys {
		f, cmd = f.Update(bezeltest.Key(k))
	}
	return f, cmd
}

func TestTabMovesTheFocusAndWraps(t *testing.T) {
	f := sample(t)
	require.Equal(t, 0, f.Focused())
	f, _ = press(f, "tab", "tab")
	require.Equal(t, 2, f.Focused())
	f, _ = press(f, "shift+tab", "shift+tab", "shift+tab")
	require.Equal(t, 5, f.Focused(), "shift+tab from the first wraps to the last")
}

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

func TestViewLinesUpTheValues(t *testing.T) {
	rows := strings.Split(ansi.Strip(sample(t).View(60)), "\n")
	require.Len(t, rows, 6)
	for i, r := range rows {
		require.Equal(t, 60, ansi.StringWidth(r), "row %d", i)
	}
	// Columns, not bytes: the focus marker is one cell and three bytes.
	column := func(r, chars string) int { return ansi.StringWidth(r[:strings.IndexAny(r, chars)]) }
	col := column(rows[0], "[")
	for _, r := range rows[1:5] {
		require.Equal(t, col, column(r, "[‹(○─"), "every value starts in one column:\n%s", strings.Join(rows, "\n"))
	}
	require.True(t, strings.HasPrefix(rows[0], "› Name"), "the focus marker leads the focused row")
}
