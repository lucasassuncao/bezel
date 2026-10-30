package doccheck

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Every Go snippet in the README and the guides names only API bezel has.
func TestTheGuidesMatchTheAPI(t *testing.T) {
	c, err := NewChecker()
	require.NoError(t, err)
	files, err := filepath.Glob("../../docs/*.md")
	require.NoError(t, err)
	for _, f := range append(files, "../../README.md") {
		problems, err := c.ReadFile(f)
		require.NoError(t, err)
		for _, p := range problems {
			t.Errorf("%s:%d: %s", filepath.Base(f), p.Line, p.Msg)
		}
	}
}

// A snippet that names something bezel does not have is caught; one that
// only leaves its own names undeclared is not.
func TestCheckTellsTheAPIFromTheSnippetsOwnNames(t *testing.T) {
	c, err := NewChecker()
	require.NoError(t, err)
	stale := "```go\nth := theme.Resolve(theme.Gruvbox(), true)\n```\n"
	require.Len(t, c.Check(stale), 1)
	own := "```go\nfunc (m model) View() tea.View { return m.sh.View(nil) }\n```\n"
	require.Empty(t, c.Check(own))
}
