package tree

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

func filterable(h int) Model[node] {
	return newTree(h).WithFilter(func(n node) string { return n.label })
}

func typeText(m Model[node], s string) Model[node] {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func labels(m Model[node]) []string {
	var out []string
	for _, i := range m.Visible() {
		out = append(out, m.Nodes()[i].label)
	}
	return out
}

func TestSlashOpensTheFilterOnlyWithWithFilter(t *testing.T) {
	_, r := newTree(10).Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	require.False(t, r.Handled, "a tree with no text to match leaves / to its host")

	m, r := filterable(10).Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	require.True(t, r.Handled)
	require.True(t, m.Filtering())
}

// A match under a closed parent shows with it, and the cursor lands on the
// match rather than on the ancestor shown for context.
func TestTheFilterShowsMatchesWithTheirAncestors(t *testing.T) {
	m, _ := filterable(10).Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m = typeText(m, "TT")
	require.Equal(t, []string{"cache", "ttl"}, labels(m))
	require.Equal(t, "ttl", label(m))
	require.Equal(t, "TT", m.Filter())

	out := ansi.Strip(m.View(theme.Resolved{}, func(n node, _ int, sel bool) string { return n.label }))
	lines := strings.Split(out, "\n")
	require.Len(t, lines, 10)
	require.Equal(t, "/TT▋", lines[9])
}

func TestEnterKeepsTheNodeAndOpensItsParents(t *testing.T) {
	m, _ := filterable(10).Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m = typeText(m, "ttl")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.False(t, m.Filtering())
	require.Equal(t, "ttl", label(m))
	require.True(t, m.Nodes()[5].expanded, "cache opened on the way")
}

func TestEscGoesBackToWhereTheCursorWas(t *testing.T) {
	m := filterable(10)
	m, _ = m.Update(press(tea.KeyDown)) // host
	m, _ = m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m = typeText(m, "ttl")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.False(t, m.Filtering())
	require.Equal(t, "host", label(m))
	require.False(t, m.Nodes()[5].expanded)
}

func TestNothingMatchingShowsTheEmptyMessageAndThePrompt(t *testing.T) {
	m, _ := filterable(4).Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m = typeText(m, "zzz")
	require.Empty(t, m.Visible())
	out := ansi.Strip(m.View(theme.Resolved{}, nil))
	require.Contains(t, out, "(empty)")
	require.True(t, strings.HasSuffix(out, "/zzz▋"))
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	require.Equal(t, "zz", m.Filter())
}
