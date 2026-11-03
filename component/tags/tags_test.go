package tags

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Plain(), true)

func focused() Model {
	m, _ := New(th).Focus()
	return m
}

// typed types s and presses enter.
func typed(m Model, s string) (Model, any) {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	m, r := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	return m, r.Event
}

func backspace(m Model) (Model, any) {
	m, r := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	return m, r.Event
}

func TestEnterAddsTheTextAsATag(t *testing.T) {
	m, ev := typed(focused(), " go ")
	require.Equal(t, Added{Tag: "go"}, ev)
	m, _ = typed(m, "tui")
	require.Equal(t, []string{"go", "tui"}, m.Tags())
	require.Empty(t, m.Value(), "the field empties for the next")

	m, ev = typed(m, "go")
	require.Nil(t, ev, "a tag is there once")
	require.Equal(t, "go", m.Value(), "the text stays to be changed")
}

func TestBackspaceOnAnEmptyFieldRemovesTheLastTag(t *testing.T) {
	m := focused().SetTags([]string{"go", "tui", "", "go"})
	require.Equal(t, []string{"go", "tui"}, m.Tags())
	m, ev := backspace(m)
	require.Equal(t, Removed{Tag: "tui"}, ev)

	m, _ = m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	m, ev = backspace(m)
	require.Nil(t, ev, "with text, backspace edits it")
	require.Equal(t, []string{"go"}, m.Tags())
}

func TestTheChipsComeBeforeTheField(t *testing.T) {
	m := focused().SetTags([]string{"go", "tui"}).SetWidth(30)
	v := m.View()
	require.Contains(t, v, DefaultStyles(th).Chip.Render(" go "))
	require.Equal(t, 30, ansi.StringWidth(v))
	require.True(t, strings.HasPrefix(ansi.Strip(v), " go   tui  "))
}

func TestChipsThatDoNotFitAreCounted(t *testing.T) {
	m := focused().SetTags([]string{"alpha", "beta", "gamma", "delta"}).SetWidth(24)
	v := ansi.Strip(m.View())
	require.True(t, strings.HasPrefix(v, "+3  delta "), v)
	require.Equal(t, 24, ansi.StringWidth(v))
}

func TestABlurredFieldTakesNoKeys(t *testing.T) {
	_, r := New(th).SetTags([]string{"go"}).Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	require.False(t, r.Handled)
	require.Len(t, focused().Legend(), 3)
}
