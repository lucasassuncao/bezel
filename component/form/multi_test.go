package form

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/icon"
)

var tools = []string{"git", "docker", "node", "golang", "dotnet", "python", "rust", "deno"}

func multiForm(m MultiSelect) Form {
	return New(th(), m.WithIcons(icon.Unicode()), NewButton("Create", nil))
}

func multiOf(f Form) MultiSelect { return f.Field(0).(MultiSelect) }

func TestMultiSelectChecksAndKeeps(t *testing.T) {
	f, _ := press(multiForm(NewMultiSelect("Tools", tools)), "enter")
	require.True(t, multiOf(f).Holding())
	f, _ = press(f, "space", "down", "space", "tab")
	require.Equal(t, 0, f.Focused(), "tab does not leave an open list")
	f, _ = press(f, "enter")
	require.False(t, multiOf(f).Holding())
	require.Equal(t, []string{"git", "docker"}, multiOf(f).Values())
	require.Equal(t, []int{0, 1}, multiOf(f).Checked())
	require.Contains(t, ansi.Strip(f.SetWidth(60).View()), "‹ git, docker ›")
}

func TestMultiSelectEscPutsTheChecksBack(t *testing.T) {
	f, _ := press(multiForm(NewMultiSelect("Tools", tools).WithChecked(2)), "enter", "space", "down", "down", "space", "esc")
	require.False(t, multiOf(f).Holding())
	require.Equal(t, []string{"node"}, multiOf(f).Values())
}

func TestMultiSelectFiltersFuzzily(t *testing.T) {
	f, _ := press(multiForm(NewMultiSelect("Tools", tools)), "enter", "d", "t")
	require.Equal(t, "dt", multiOf(f).Filter())
	out := ansi.Strip(f.SetWidth(60).View())
	require.Contains(t, out, "/ dt")
	require.Contains(t, out, "dotnet")
	require.NotContains(t, out, "deno")

	f, _ = press(f, "space", "backspace", "backspace", "x", "y", "z")
	require.Contains(t, ansi.Strip(f.SetWidth(60).View()), "no match")
	f, _ = press(f, "space", "enter")
	require.Equal(t, []string{"dotnet"}, multiOf(f).Values(), "checking with no match does nothing")
	require.Empty(t, multiOf(f).Filter(), "closing clears the filter")
}

func TestMultiSelectChecksAllItShows(t *testing.T) {
	f, _ := press(multiForm(NewMultiSelect("Tools", tools)), "enter", "d", "o", "ctrl+a")
	require.Equal(t, []string{"docker", "dotnet", "deno"}, multiOf(f).Values())
	f, _ = press(f, "ctrl+a")
	require.Empty(t, multiOf(f).Values(), "a second ctrl+a unchecks them")
}

func TestMultiSelectScrollsItsWindow(t *testing.T) {
	f := multiForm(NewMultiSelect("Tools", tools).WithHeight(3))
	f, _ = press(f, "enter")
	out := ansi.Strip(f.SetWidth(60).View())
	require.Contains(t, out, "↓ 6 more")
	require.Len(t, strings.Split(out, "\n"), 6, "head, filter, three rows, button")

	f, _ = press(f, "down", "down", "down", "down")
	out = ansi.Strip(f.SetWidth(60).View())
	require.Contains(t, out, "→ [ ] dotnet")
	require.NotContains(t, out, "git")
}

func TestMultiSelectLettersAreFilterTextUnderVimKeys(t *testing.T) {
	f := multiForm(NewMultiSelect("Tools", tools)).WithKeys(VimKeys())
	f, _ = press(f, "enter", "j")
	require.Equal(t, "j", multiOf(f).Filter())
}

func TestMultiSelectLegendWhileOpen(t *testing.T) {
	f, _ := press(multiForm(NewMultiSelect("Tools", tools)), "enter")
	var got []string
	for _, e := range f.Legend() {
		got = append(got, e.Help.Key+" "+e.Help.Desc)
	}
	require.Equal(t, []string{"↑/↓ move", "space check", "ctrl+a check all", "enter done", "esc close"}, got)
}

func TestMultiSelectStaysInItsWidth(t *testing.T) {
	f, _ := press(multiForm(NewMultiSelect("Tools", tools).WithChecked(0, 1, 2, 3, 4, 5)), "enter")
	for _, w := range []int{1, 8, 20} {
		for i, l := range strings.Split(f.SetWidth(w).View(), "\n") {
			require.LessOrEqual(t, ansi.StringWidth(l), w, "width %d row %d", w, i)
		}
	}
}

func TestMultiSelectIgnoresOtherMessagesAndOutOfRangeChecks(t *testing.T) {
	m := NewMultiSelect("Tools", nil).WithChecked(-1, 3)
	require.Empty(t, m.Values())
	fl, r := m.Update(nil)
	require.False(t, r.Handled)
	fl, _ = fl.Update(bezelKey("enter"))
	fl, r = fl.Update(bezelKey("space"))
	require.True(t, r.Handled)
	fl, _ = fl.Update(bezelKey("ctrl+a"))
	_, r = fl.Update(bezelKey("f5"))
	require.False(t, r.Handled)
	require.Empty(t, fl.(MultiSelect).Values())
}

func TestMultiSelectBlurKeepsTheChecks(t *testing.T) {
	fl, _ := NewMultiSelect("Tools", tools).Update(bezelKey("enter"))
	fl, _ = fl.Update(bezelKey("space"))
	fl = fl.Blur()
	require.False(t, fl.(MultiSelect).Holding())
	require.Equal(t, []string{"git"}, fl.(MultiSelect).Values())
}
