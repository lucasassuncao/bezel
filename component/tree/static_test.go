package tree

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

func TestStaticHangsChildrenFromGuides(t *testing.T) {
	got := Static([]Branch{{Text: "app", Children: []Branch{
		{Text: "cmd", Children: []Branch{{Text: "main.go"}}},
		{Text: "go.mod"},
	}}}, theme.Resolve(theme.Mint(), true))
	require.Equal(t, "app\n├─ cmd\n│  └─ main.go\n└─ go.mod", ansi.Strip(got))
}

func TestStaticPrintsEachTopBranchOnItsOwnLine(t *testing.T) {
	got := Static([]Branch{{Text: "a"}, {Text: "b"}}, theme.Resolve(theme.Plain(), true))
	require.Equal(t, "a\nb", got, "no guide, so no escape codes either")
	require.Empty(t, Static(nil, theme.Resolved{}))
}

func TestStaticExpandsATabInARow(t *testing.T) {
	got := Static([]Branch{{Text: "a\tb", Children: []Branch{{Text: "c\td"}}}}, theme.Resolve(theme.Plain(), true))
	require.Equal(t, "a   b\n└─ c    d", ansi.Strip(got))
}

// Static is printed straight to the terminal: a branch from data must not be
// able to set the clipboard, clear the screen or rewrite its line.
func TestStaticPrintsNoCommandsFromItsBranches(t *testing.T) {
	evil := "pod \x1b]52;c;aGk=\x07 \x1b[2J\r x"
	out := Static([]Branch{{Text: evil, Children: []Branch{{Text: evil + "\nnext"}}}}, theme.Resolve(theme.Plain(), true))
	require.Equal(t, draw.Clean(out), out)
	require.Len(t, strings.Split(out, "\n"), 2, "a branch is one row")
}
