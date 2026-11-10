package tree

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

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
