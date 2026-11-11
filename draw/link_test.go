package draw

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

const osc8 = "\x1b]8;"

func TestLinkWrapsTheTextInOSC8(t *testing.T) {
	out := Link("docs", "https://example.com/a", lipgloss.NewStyle())
	require.True(t, strings.HasPrefix(out, osc8), out)
	require.Contains(t, out, ";https://example.com/a\a")
	require.True(t, strings.HasSuffix(out, osc8+";\a"), "the link is closed")
	require.Equal(t, "docs", ansi.Strip(out))
	require.Equal(t, 4, Width(out), "only the text takes cells")
}

func TestLinkWithoutTextShowsTheURL(t *testing.T) {
	require.Equal(t, "https://example.com", ansi.Strip(Link("", "https://example.com", lipgloss.NewStyle())))
}

func TestLinkWithoutAURLIsTheStyledText(t *testing.T) {
	st := lipgloss.NewStyle().Bold(true)
	require.Equal(t, st.Render("docs"), Link("docs", "", st))
	require.Equal(t, st.Render("docs"), Link("docs", "\x1b\a", st), "a url of only controls is none")
}

// A url from outside must not end the link early and run what follows.
func TestLinkDropsControlsFromTheURL(t *testing.T) {
	out := Link("x", "https://e.com/\x1b]8;;\aevil\x1b[2J\u202e", lipgloss.NewStyle())
	require.Equal(t, 2, strings.Count(out, osc8), "one opening, one closing")
	require.Contains(t, out, ";https://e.com/]8;;evil[2J\a")
	require.NotContains(t, out, "\x1b[2J")
}

func TestLinkSanitizesTheText(t *testing.T) {
	out := Link("a\x1b[2Jb", "https://e.com", lipgloss.NewStyle())
	require.Equal(t, `a\e[2Jb`, ansi.Strip(out))
}

// Each line opens and closes its own link, with the same id on all of them.
func TestLinkOfManyLinesLinksEachLine(t *testing.T) {
	out := Link("one\ntwo", "https://e.com", lipgloss.NewStyle())
	lines := strings.Split(out, "\n")
	require.Len(t, lines, 2)
	id := ""
	for _, l := range lines {
		require.True(t, strings.HasPrefix(l, osc8+"id=bezel-"), l)
		require.True(t, strings.HasSuffix(l, osc8+";\a"), l)
		got := strings.SplitN(strings.TrimPrefix(l, osc8), ";", 2)[0]
		if id == "" {
			id = got
		}
		require.Equal(t, id, got)
	}
	other := Link("x", "https://other.com", lipgloss.NewStyle())
	require.NotContains(t, other, id)
}

// Cut short or covered by a dialog, a link still closes before what follows.
func TestALinkCutOrCoveredStaysClosed(t *testing.T) {
	out := Link("abcdef", "https://e.com", lipgloss.NewStyle())
	cut := Truncate(out, 3)
	require.Equal(t, "ab…", ansi.Strip(cut))
	require.Equal(t, 2, strings.Count(cut, osc8))

	over := Composite("XY", out, 2, 0)
	require.Equal(t, "abXYef", ansi.Strip(over))
	before, after, ok := strings.Cut(over, "XY")
	require.True(t, ok)
	require.True(t, strings.HasSuffix(before, osc8+";\a"), "closed before the dialog")
	require.True(t, strings.HasPrefix(after, osc8), "opened again after it")
}
