package form

import (
	"charm.land/lipgloss/v2"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/theme"
)

type submitted struct{}

func th() theme.Resolved { return theme.Resolve(theme.Plain(), true) }

// sample is one of each control, in the order the tests index them.
func sample(t *testing.T) Form {
	t.Helper()
	f := New(th(),
		NewInput("Name").WithPlaceholder("my-project").WithWidth(24),
		NewSelect("Image", []string{"ubuntu", "debian", "alpine"}),
		NewRadio("Shell", []string{"bash", "zsh", "fish"}),
		NewToggle("Docker"),
		NewCheckbox("Forward port 3000"),
		NewButton("Create", func() tea.Msg { return submitted{} }),
	)
	return f
}

func press(f Form, keys ...string) (Form, tea.Cmd) {
	var cmd tea.Cmd
	for _, k := range keys {
		var r bezel.Result
		f, r = f.Update(bezeltest.Key(k))
		cmd = r.Cmd
	}
	return f, cmd
}

// click is a left click on row y of the form's body.
func click(f Form, y int) (Form, bezel.Result) {
	return f.Update(bezel.ClickMsg{Pane: "form", MouseClickMsg: tea.MouseClickMsg{Button: tea.MouseLeft, Y: y}})
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
	rows := strings.Split(ansi.Strip(sample(t).SetWidth(60).View()), "\n")
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
	fields := []Field{NewCheckbox("a")}
	before := New(th(), fields...)
	after, _ := press(before, "space")

	require.True(t, after.Field(0).(Checkbox).Checked())
	require.False(t, before.Field(0).(Checkbox).Checked(), "the earlier copy changed")
	require.False(t, fields[0].(Checkbox).Checked(), "the caller's slice changed")
}

func binding(keys ...string) key.Binding { return key.New(keys...) }

// One call rebinds the form's focus keys and the keys of every control in it.
func TestFormWithKeysRebindsTheFocusAndItsControls(t *testing.T) {
	keys := DefaultKeys()
	keys.Next, keys.Press = binding("ctrl+n"), binding("x")
	f := New(th(), NewCheckbox("a"), NewButton("go", func() tea.Msg { return submitted{} }))
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
	f := New(th(), NewInput("Name").WithWidth(30), NewCheckbox("Port"), NewToggle("Docker"))
	back := tea.KeyPressMsg{Code: tea.KeyBackspace}
	for b.Loop() {
		f, _ = f.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
		f, _ = f.Update(back)
		_ = f.SetWidth(60).View()
	}
}

func TestNilFieldsAreLeftOutAndFieldIsBounded(t *testing.T) {
	f := New(th(), nil, NewCheckbox("a"), nil)
	require.NotPanics(t, func() { _ = f.SetWidth(40).View() })
	require.NotNil(t, f.Field(0), "the checkbox is first once the nil is gone")
	require.Nil(t, f.Field(1))
	require.Nil(t, f.Field(-1))
}

func TestClickFocusesTheFieldUnderIt(t *testing.T) {
	f := sample(t).SetWidth(60)
	f, r := click(f, 5)
	require.True(t, r.Handled)
	require.Equal(t, 5, f.Focused())
	require.NotNil(t, r.Cmd)
	require.Equal(t, submitted{}, r.Cmd())

	f, r = click(f, 3)
	require.True(t, r.Handled)
	toggle, _ := Get[Toggle](f, 3)
	require.True(t, toggle.On(), "a toggle flips")
	f, _ = click(f, 4)
	box, _ := Get[Checkbox](f, 4)
	require.True(t, box.Checked())

	f, _ = click(f, 0)
	require.Equal(t, 0, f.Focused(), "an input only takes the focus")
}

func TestClickPastTheFieldsIsNotHandled(t *testing.T) {
	f, r := click(sample(t).SetWidth(60), 99)
	require.Equal(t, bezel.Result{}, r)
	require.Equal(t, 0, f.Focused())
}

func TestTabIsHandled(t *testing.T) {
	f, r := sample(t).Update(bezeltest.Key("tab"))
	require.True(t, r.Handled)
	require.Nil(t, r.Event)
	require.Equal(t, 1, f.Focused())
	_, r = f.Update(bezeltest.Key("shift+tab"))
	require.True(t, r.Handled)
}

func TestForeignKeyIsNotHandled(t *testing.T) {
	f := New(th(), NewCheckbox("a"))
	_, r := f.Update(bezeltest.Key("f5"))
	require.Equal(t, bezel.Result{}, r)
	_, r = New(th()).Update(bezeltest.Key("tab"))
	require.Equal(t, bezel.Result{}, r, "no fields")
}

func TestAFocusedInputHandlesTextButNotF5(t *testing.T) {
	f := sample(t)
	_, r := f.Update(bezeltest.Key("a"))
	require.True(t, r.Handled)
	_, r = f.Update(bezeltest.Key("f5"))
	require.False(t, r.Handled)
}

func TestControlsHandleOnlyTheirKeys(t *testing.T) {
	cases := []struct {
		name  string
		field Field
		yes   []string
		no    []string
	}{
		{"checkbox", NewCheckbox("a"), []string{"space", "enter"}, []string{"left", "x", "f5"}},
		{"radio", NewRadio("r", []string{"a", "b"}), []string{"left", "right"}, []string{"space", "x", "f5"}},
		{"toggle", NewToggle("t"), []string{"space", "enter", "left", "right"}, []string{"up", "x", "f5"}},
		{"select closed", NewSelect("s", []string{"a"}), []string{"enter", "space"}, []string{"up", "down", "esc", "f5"}},
		{"button", NewButton("b", nil), []string{"space", "enter"}, []string{"left", "x", "f5"}},
	}
	for _, c := range cases {
		for _, k := range c.yes {
			_, r := c.field.Update(bezeltest.Key(k))
			require.True(t, r.Handled, "%s %s", c.name, k)
			require.Nil(t, r.Event)
		}
		for _, k := range c.no {
			_, r := c.field.Update(bezeltest.Key(k))
			require.Equal(t, bezel.Result{}, r, "%s %s", c.name, k)
		}
	}
	open, _ := NewSelect("s", []string{"a", "b"}).Update(bezeltest.Key("enter"))
	for _, k := range []string{"up", "down", "enter", "esc"} {
		_, r := open.Update(bezeltest.Key(k))
		require.True(t, r.Handled, "open select %s", k)
	}
	_, r := open.Update(bezeltest.Key("f5"))
	require.False(t, r.Handled)
}

func TestInputPassesItsResultThrough(t *testing.T) {
	f, _ := NewInput("n").Focus()
	_, r := f.Update(bezeltest.Key("a"))
	require.True(t, r.Handled)
	_, r = NewInput("n").Update(bezeltest.Key("a"))
	require.False(t, r.Handled, "an unfocused input takes no text")
}

func TestAClickCountsTheRowsOfAnOpenSelect(t *testing.T) {
	f, _ := press(sample(t), "tab", "enter") // the select, open under its row
	_, h := draw.BlockSize(f.Field(1).View(true, 20, 60))
	require.Greater(t, h, 1)
	f, _ = click(f.SetWidth(60), 1+h)
	require.Equal(t, 2, f.Focused(), "the radio sits under the open list")
}

// The form hands its theme to every bezel control, at New and on SetTheme.
func TestTheFormPaintsItsControlsInItsTheme(t *testing.T) {
	bare := NewCheckbox("a").WithChecked(true).View(false, 0, 20)
	f := New(th(), NewCheckbox("a").WithChecked(true))
	require.NotEqual(t, bare, f.SetWidth(20).View(), "the form's theme reached the box")

	other := theme.Resolve(theme.Mint(), true)
	require.Equal(t, New(other, NewCheckbox("a").WithChecked(true)).SetWidth(20).View(), f.SetTheme(other).SetWidth(20).View())
}

func TestInitIsNilWithoutAnIniterInFocus(t *testing.T) {
	require.Nil(t, New(th()).Init(), "no fields")
	require.Nil(t, New(th(), NewCheckbox("a")).Init())
}

// One edit reaches every control, and outlives a theme switch.
func TestWithStylesRecoloursEveryControl(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	f := New(th(), NewCheckbox("a").WithChecked(true)).WithStyles(func(s Styles) Styles { s.Chosen = red; return s })
	f = f.SetTheme(theme.Resolve(theme.Mint(), true))
	require.Contains(t, f.SetWidth(20).View(), red.Render("[x]"))
}

func TestLegendFollowsTheFocusedControl(t *testing.T) {
	descs := func(f Form) (out []string) {
		for _, e := range f.Legend() {
			out = append(out, e.Help.Desc)
		}
		return out
	}
	f := sample(t)
	f, _ = press(f, "tab")
	require.Equal(t, []string{"next", "previous", "press"}, descs(f), "a closed select opens on a press")
	f, _ = press(f, "enter")
	require.Equal(t, []string{"move", "press", "close"}, descs(f), "an open select keeps its keys")
	f, _ = press(f, "esc", "tab")
	require.Equal(t, []string{"next", "previous", "choose"}, descs(f), "a radio")
	require.Empty(t, New(th()).Legend())
}

func TestFormIsAModel(t *testing.T) {
	var _ bezel.Model[Form] = Form{}
}
