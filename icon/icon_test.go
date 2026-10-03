package icon

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

func TestEveryStateHasAMarkInBothSets(t *testing.T) {
	for _, set := range []Set{Unicode, ASCII} {
		seen := map[string]State{}
		for st := OK; st <= Skipped; st++ {
			mark := set.Mark(st)
			require.NotEmpty(t, mark, "state %d has no mark", st)
			if prev, dup := seen[mark]; dup {
				t.Errorf("states %d and %d share %q", prev, st, mark)
			}
			seen[mark] = st
		}
	}
}

func TestASCIIStaysASCII(t *testing.T) {
	for st := OK; st <= Skipped; st++ {
		for _, r := range ASCII.Mark(st) {
			require.Less(t, r, rune(128), "state %d", st)
		}
	}
}

func TestDefaultFallsBackToASCII(t *testing.T) {
	t.Setenv("BEZEL_ASCII", "")
	t.Setenv("TERM", "xterm-256color")
	require.Equal(t, Unicode, Default())

	t.Setenv("TERM", "dumb")
	require.Equal(t, ASCII, Default())

	t.Setenv("TERM", "xterm-256color")
	t.Setenv("BEZEL_ASCII", "1")
	require.Equal(t, ASCII, Default())
}

func TestRenderKeepsTheMark(t *testing.T) {
	th := theme.Resolve(theme.ThemePlain, true)
	require.Equal(t, "✓", ansi.Strip(Unicode.Render(OK, th)))
	require.Equal(t, "x", ansi.Strip(ASCII.Render(Fail, th)))
}

func TestConnectorsDrawATree(t *testing.T) {
	// A parent with two children, the second holding one child of its own.
	tree := func(s Set) string {
		return strings.Join([]string{
			"src",
			s.Connector(false) + "cmd",
			s.Connector(true) + "internal",
			s.Indent(true) + s.Connector(true) + "tui",
		}, "\n")
	}
	require.Equal(t, "src\n├─ cmd\n└─ internal\n   └─ tui", tree(Unicode))
	require.Equal(t, "src\n|- cmd\n`- internal\n   `- tui", tree(ASCII))
	require.Equal(t, "│  ", Unicode.Indent(false))
	for _, s := range []Set{Unicode, ASCII} {
		for _, g := range []string{s.Branch, s.LastBranch, s.Pipe, s.Indent(true)} {
			require.Equal(t, 3, ansi.StringWidth(g), "%q keeps rows aligned", g)
		}
	}
}
