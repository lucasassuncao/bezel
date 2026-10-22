package form

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/theme"
)

type submitted struct{}

func th() theme.Resolved { return theme.Resolve(theme.Plain, true) }

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

// Get reads a field by its type; a wrong index or type is false, not a panic.
func TestGetReadsAFieldAsItsType(t *testing.T) {
	f, _ := press(sample(t), "h", "i")
	in, ok := Get[Input](f, 0)
	require.True(t, ok)
	require.Equal(t, "hi", in.Value())

	_, ok = Get[Select](f, 0)
	require.False(t, ok, "field 0 is an Input")
	_, ok = Get[Input](f, 99)
	require.False(t, ok)
	_, ok = Get[Input](f, -1)
	require.False(t, ok)
}

func TestTabMovesTheFocusAndWraps(t *testing.T) {
	f := sample(t)
	require.Equal(t, 0, f.Focused())
	f, _ = press(f, "tab", "tab")
	require.Equal(t, 2, f.Focused())
	f, _ = press(f, "shift+tab", "shift+tab", "shift+tab")
	require.Equal(t, 5, f.Focused(), "shift+tab from the first wraps to the last")
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

// A Form is a value: updating one must not reach the copy it came from, nor
// the slice the fields were handed in.
func TestUpdateLeavesEarlierCopiesAlone(t *testing.T) {
	fields := []Field{NewCheckbox("a", false, th())}
	before, _ := New(fields...)
	after, _ := press(before, "space")

	require.True(t, after.Field(0).(Checkbox).Checked())
	require.False(t, before.Field(0).(Checkbox).Checked(), "the earlier copy changed")
	require.False(t, fields[0].(Checkbox).Checked(), "the caller's slice changed")
}

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

// A keystroke into the focused input and the form's frame.
func BenchmarkTypingInAForm(b *testing.B) {
	f, _ := New(NewInput("Name", "", 30, th()), NewCheckbox("Port", false, th()), NewToggle("Docker", false, th()))
	back := tea.KeyPressMsg{Code: tea.KeyBackspace}
	for b.Loop() {
		f, _ = f.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
		f, _ = f.Update(back)
		_ = f.View(60)
	}
}
