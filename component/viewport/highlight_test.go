package viewport

import (
	"regexp"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestHighlightsPaintEveryMatchAndTheCurrentOne(t *testing.T) {
	m := sized("an ant\nbat", 6, 2).SetHighlights([][]int{{0, 2}, {3, 5}})
	require.Equal(t, 2, m.Highlights())
	require.Equal(t, 0, m.Highlight())
	require.Equal(t, th.MatchCurrent.Render("an")+" "+th.Match.Render("an")+"t", strings.Split(m.View(th), "\n")[0])
}

func TestNextAndPrevHighlightWrapAroundAndScrollToIt(t *testing.T) {
	content := numbered(30)
	matches := regexp.MustCompile(`(?m)^line (1|21|30)$`).FindAllStringIndex(content, -1)
	m := sized(content, 20, 5).SetHighlights(matches).NextHighlight()
	require.Equal(t, 1, m.Highlight())
	require.Equal(t, 20, m.YOffset(), "the match scrolls into view")
	m = m.NextHighlight().NextHighlight()
	require.Equal(t, 0, m.Highlight(), "past the last it wraps to the first")
	require.Equal(t, 2, m.PrevHighlight().Highlight())
}

func TestScrollingMovesTheCurrentMatchToTheFirstBelowTheTop(t *testing.T) {
	content := numbered(30)
	matches := regexp.MustCompile(`(?m)^line (1|21|30)$`).FindAllStringIndex(content, -1)
	require.Equal(t, 1, sized(content, 20, 5).SetHighlights(matches).ScrollDown(10).Highlight())
}

func TestAMatchAcrossLinesPaintsEachLine(t *testing.T) {
	r := strings.Split(sized("ab\ncd", 2, 2).SetHighlights([][]int{{1, 4}}).View(th), "\n")
	require.Equal(t, "a"+th.MatchCurrent.Render("b"), r[0])
	require.Equal(t, th.MatchCurrent.Render("c")+"d", r[1])
}

func TestHighlightsCountCellsNotBytes(t *testing.T) {
	m := sized("é🌍x", 4, 1).SetHighlights([][]int{{len("é🌍"), len("é🌍x")}})
	require.Equal(t, "é🌍"+th.MatchCurrent.Render("x"), m.View(th))
}

func TestHighlightsMatchThePlainTextOfStyledContent(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("ab")
	m := sized(red+"cd", 4, 1).SetHighlights([][]int{{2, 3}})
	require.Equal(t, "abcd", ansi.Strip(m.View(th)))
	require.Contains(t, m.View(th), th.MatchCurrent.Render("c"))
}

func TestSetContentAndClearDropTheHighlights(t *testing.T) {
	m := sized("ab", 2, 1).SetHighlights([][]int{{0, 1}})
	require.Equal(t, 0, m.SetContent("cd").Highlights())
	require.Equal(t, -1, m.ClearHighlights().Highlight())
	require.Equal(t, -1, m.SetHighlights(nil).Highlight())
}
