package list

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Plain, true)

func rows() []Row[int] {
	return []Row[int]{
		{Section: true, Label: "ADDED"},
		{Label: "server", Mark: "●", Value: 1},
		{Label: "logging", Mark: "●", Value: 2},
		{Section: true},
		{Section: true, Label: "AVAILABLE"},
		{Label: "cache", Mark: "+", Value: 3},
		{Label: "metrics", Mark: "+", Value: 4},
	}
}

func press(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }
func typed(s string) tea.KeyPressMsg  { return tea.KeyPressMsg{Code: []rune(s)[0], Text: s} }

func TestCursorStartsOnAndSkipsOverHeadings(t *testing.T) {
	m := New(rows()).SetHeight(10)
	require.Equal(t, "server", m.Selected().Label)
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown)) // hops the spacer and the heading
	require.Equal(t, "cache", m.Selected().Label)
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown)) // clamped at the end
	require.Equal(t, "metrics", m.Selected().Label)
	require.Equal(t, 4, m.Selected().Value)
}

func TestEnterChoosesASelectableRow(t *testing.T) {
	m := New(rows()).SetHeight(10)
	_, act := m.Update(press(tea.KeyEnter))
	require.Equal(t, Chosen, act)
	require.Nil(t, New([]Row[int]{{Section: true, Label: "x"}}).SetHeight(5).Selected())
}

func TestFilterNarrowsAndEnterPicksTheMatch(t *testing.T) {
	m := New(rows()).SetHeight(10)
	m, _ = m.Update(typed("/"))
	require.True(t, m.Filtering())
	m, _ = m.Update(typed("m"))
	m, _ = m.Update(typed("e"))
	require.Equal(t, "me", m.Filter())
	require.Equal(t, "metrics", m.Selected().Label)
	require.Contains(t, ansi.Strip(m.View(th)), "/me▋")

	m, act := m.Update(press(tea.KeyEnter))
	require.Equal(t, Chosen, act)
	require.False(t, m.Filtering())
	require.Equal(t, "metrics", m.Selected().Label, "the pick becomes the cursor")

	m, _ = m.Update(typed("/"))
	m, _ = m.Update(typed("z"))
	require.Nil(t, m.Selected())
	m, _ = m.Update(press(tea.KeyBackspace))
	require.Equal(t, "", m.Filter())
	m, _ = m.Update(press(tea.KeyEscape))
	require.False(t, m.Filtering())
}

func TestSetRowsKeepsTheCursorByLabel(t *testing.T) {
	m := New(rows()).SetHeight(10)
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, "logging", m.Selected().Label)
	m = m.SetRows([]Row[int]{{Label: "cache"}, {Label: "logging"}})
	require.Equal(t, "logging", m.Selected().Label)
	m = m.SetRows([]Row[int]{{Label: "other"}})
	require.Equal(t, "other", m.Selected().Label)
}

func TestViewShowsHeadingsMarksAndTheOverflowIndicator(t *testing.T) {
	m := New(rows()).SetHeight(4)
	got := ansi.Strip(m.View(th))
	lines := strings.Split(got, "\n")
	require.Len(t, lines, 4)
	require.Equal(t, " ADDED", lines[0])
	require.Equal(t, "▶ ●  server", lines[1])
	require.Equal(t, "  ●  logging", lines[2])
	require.Contains(t, lines[3], "↓ 4 more")

	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown))
	require.Contains(t, ansi.Strip(m.View(th)), "▶ +  cache", "scrolled to keep the cursor visible")
}

func TestScrollIsWhereTheWindowSits(t *testing.T) {
	m := New(rows()).SetHeight(4)
	require.Equal(t, layout.Scroll{Visible: 3, Total: 7}, m.Scroll(), "the more row is not a row shown")
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, layout.Scroll{Offset: 3, Visible: 4, Total: 7}, m.Scroll())

	f, _ := New(numbered(20)).SetHeight(6).Update(typed("/"))
	require.Equal(t, layout.Scroll{Visible: 4, Total: 20}, f.Scroll(), "filtering: the matches' window")
}

func TestViewIsEmptyBeforeSized(t *testing.T) {
	require.Empty(t, New(rows()).View(th))
}

func TestOneRowListShowsTheCursorNotTheIndicator(t *testing.T) {
	m := New(rows()).SetHeight(1)
	m, _ = m.Update(press(tea.KeyDown))
	require.Contains(t, ansi.Strip(m.View(th)), "logging")
}

func twins() []Row[int] {
	return []Row[int]{{Label: "dup", Value: 1}, {Label: "dup", Value: 2}}
}

func TestFilterPicksTheRowItShowsAmongEqualLabels(t *testing.T) {
	m := New(twins()).SetHeight(5)
	m, _ = m.Update(typed("/"))
	m, _ = m.Update(press(tea.KeyDown))
	m, act := m.Update(press(tea.KeyEnter))
	require.Equal(t, Chosen, act)
	require.Equal(t, 2, m.Selected().Value)
}

func TestSetRowsKeepsTheSameRowAmongEqualLabels(t *testing.T) {
	m := New(twins()).SetHeight(5)
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, 2, m.SetRows(twins()).Selected().Value)
}

// The list is a value: the caller's slice must not reach into it.
func TestNewAndSetRowsCopyTheSlice(t *testing.T) {
	xs := rows()
	m := New(xs).SetHeight(10)
	xs[1].Label = "changed"
	require.NotContains(t, ansi.Strip(m.View(th)), "changed")

	ys := rows()
	m = m.SetRows(ys)
	ys[1].Label = "changed"
	require.NotContains(t, ansi.Strip(m.View(th)), "changed")
}

func numbered(n int) []Row[int] {
	rs := make([]Row[int], n)
	for i := range rs {
		rs[i] = Row[int]{Label: fmt.Sprintf("item-%02d", i), Value: i}
	}
	return rs
}

// The matches scroll the way the list does: a window that keeps the cursor
// in view and counts the rows under it, the prompt on the last row.
func TestFilterScrollsItsMatchesAndCountsTheRest(t *testing.T) {
	m := New(numbered(20)).SetHeight(6)
	m, _ = m.Update(typed("/"))
	lines := strings.Split(ansi.Strip(m.View(th)), "\n")
	require.Len(t, lines, 6)
	require.Contains(t, lines[4], "↓ 16 more")
	require.Equal(t, "/▋", lines[5])

	for range 10 {
		m, _ = m.Update(press(tea.KeyDown))
	}
	require.Equal(t, "item-10", m.Selected().Label)
	require.Contains(t, ansi.Strip(m.View(th)), "▶ item-10")
}

// The payload keeps its type: nothing to assert between the row and the caller.
func TestSelectedValueKeepsItsType(t *testing.T) {
	m := New([]Row[string]{{Label: "a", Value: "payload"}}).SetHeight(5)
	// strings.ToUpper compiles only while Value is a string, not an any.
	require.Equal(t, "PAYLOAD", strings.ToUpper(m.Selected().Value))
}

func TestWithMatchDecidesWhatTheFilterKeeps(t *testing.T) {
	filter := func(m Model[int], text string) Model[int] {
		m, _ = m.Update(typed("/"))
		for _, r := range text {
			m, _ = m.Update(typed(string(r)))
		}
		return m
	}
	require.Equal(t, "server", filter(New(rows()).SetHeight(10), "e").Selected().Label, "the default keeps any label holding the text")

	prefix := New(rows()).SetHeight(10).WithMatch(Prefix)
	require.Nil(t, filter(prefix, "e").Selected(), "no label starts with e")
	require.Equal(t, "cache", filter(prefix, "c").Selected().Label)
	require.Equal(t, "server", filter(prefix.WithMatch(nil), "e").Selected().Label, "nil is the default again")
}

func TestTheMatchersIgnoreCase(t *testing.T) {
	require.True(t, Contains("RV", "server"))
	require.True(t, Prefix("SER", "server"))
	require.False(t, Prefix("rv", "server"))
}

func bigList(n int) Model[int] {
	rows := make([]Row[int], n)
	for i := range rows {
		rows[i] = Row[int]{Label: fmt.Sprintf("package-%05d", i), Value: i}
	}
	return New(rows).SetHeight(40)
}

func BenchmarkUpdateDown(b *testing.B) {
	down := tea.KeyPressMsg{Code: tea.KeyDown}
	m := bigList(100_000)
	for b.Loop() {
		m, _ = m.Update(down)
	}
}

// The filter re-matches every row on each keystroke and each frame.
func BenchmarkFilter(b *testing.B) {
	th := theme.Resolve(theme.Plain, true)
	m := bigList(100_000)
	m, _ = m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m, _ = m.Update(tea.KeyPressMsg{Code: '7', Text: "7"})
	down := tea.KeyPressMsg{Code: tea.KeyDown}
	b.Run("down", func(b *testing.B) {
		for b.Loop() {
			_, _ = m.Update(down)
		}
	})
	b.Run("view", func(b *testing.B) {
		for b.Loop() {
			_ = m.View(th)
		}
	})
}

// A frame of a list that fills a 40-row pane.
func BenchmarkListFrame(b *testing.B) {
	m := New(numbered(500)).SetHeight(40)
	for b.Loop() {
		_ = m.View(th)
	}
}

func filtered(m Model[int], text string) Model[int] {
	m, _ = m.Update(typed("/"))
	for _, r := range text {
		m, _ = m.Update(typed(string(r)))
	}
	return m
}

func TestFuzzyRanksTheBestMatchFirstAndMarksIt(t *testing.T) {
	m := New([]Row[int]{{Label: "carousel"}, {Label: "cherry-pick"}, {Label: "checkout"}}).SetHeight(10).WithFuzzy(true)
	m = filtered(m, "chk")
	require.Equal(t, "checkout", m.Selected().Label, "a run beats scattered letters")
	lines := strings.Split(ansi.Strip(m.View(th)), "\n")
	require.Equal(t, "▶ checkout", lines[0])
	require.Equal(t, "  cherry-pick", lines[1])
	require.NotContains(t, ansi.Strip(m.View(th)), "carousel")
	require.Contains(t, m.View(th), th.Cursor.Underline(true).Bold(true).Render("ch"))
	plain := New([]Row[int]{{Label: "carousel"}, {Label: "cherry-pick"}}).SetHeight(10).WithFuzzy(true).WithMatch(nil)
	require.Equal(t, "carousel", filtered(plain, "ar").Selected().Label, "WithMatch ends it: the cherry-pick letters are not a run")
}

func TestTitleAndStatusFrameTheRows(t *testing.T) {
	m := New(numbered(5)).SetHeight(5).WithTitle("Items").WithStatusBar(true)
	lines := strings.Split(ansi.Strip(m.View(th)), "\n")
	require.Len(t, lines, 5)
	require.Equal(t, "Items", lines[0])
	require.Equal(t, "  5 items", lines[4])
	require.Contains(t, lines[3], "↓ 3 more", "three lines left for rows")

	m = filtered(m, "3")
	require.Equal(t, "  1 of 5", strings.Split(ansi.Strip(m.View(th)), "\n")[4])
	m = filtered(New(numbered(5)).SetHeight(5).WithStatusBar(true), "zz")
	require.Contains(t, ansi.Strip(m.View(th)), "nothing matched")
}

func TestPagesShowARunOfRowsAndTheirDots(t *testing.T) {
	m := New(numbered(7)).SetHeight(4).WithPages(true)
	lines := strings.Split(ansi.Strip(m.View(th)), "\n")
	require.Len(t, lines, 4)
	require.Equal(t, "▶ item-00", lines[0])
	require.NotContains(t, ansi.Strip(m.View(th)), "more", "a page has no overflow row")
	m, _ = m.Update(press(tea.KeyPgDown))
	lines = strings.Split(ansi.Strip(m.View(th)), "\n")
	require.Equal(t, "▶ item-03", lines[0], "the next page")
	require.NotEqual(t, strings.Split(ansi.Strip(New(numbered(7)).SetHeight(4).WithPages(true).View(th)), "\n")[3], lines[3], "the dots moved")
}

func TestDescriptionsGiveEveryRowTwoLines(t *testing.T) {
	rows := []Row[int]{{Label: "a", Description: "first"}, {Label: "b"}, {Label: "c", Description: "third"}}
	m := New(rows).SetHeight(6)
	lines := strings.Split(ansi.Strip(m.View(th)), "\n")
	require.Equal(t, []string{"▶ a", "  first", "  b", "  ", "  c", "  third"}, lines)
	require.Contains(t, ansi.Strip(New(rows).SetHeight(4).View(th)), "↓ 2 more", "four lines hold one row and the count")
	m, _ = m.Update(press(tea.KeyDown))
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, "c", m.Selected().Label)
	require.Contains(t, ansi.Strip(m.View(th)), "  third")
}

func TestRowsGoInComeOutAndChange(t *testing.T) {
	m := New(numbered(3)).SetHeight(10)
	m, _ = m.Update(press(tea.KeyDown))
	m = m.InsertRow(0, Row[int]{Label: "new"})
	require.Equal(t, "item-01", m.Selected().Label, "the cursor stays on its row")
	require.Equal(t, 4, m.Len())
	m = m.RemoveRow(0).RemoveRow(99)
	require.Equal(t, 3, m.Len())
	m = m.SetRow(1, Row[int]{Label: "renamed"})
	require.Equal(t, "renamed", m.Rows()[1].Label)
	require.Equal(t, 3, m.SetRow(-1, Row[int]{}).Len())
}

func TestWrapGoesRoundTheEnds(t *testing.T) {
	m := New(rows()).SetHeight(10).WithWrap(true)
	m, _ = m.Update(press(tea.KeyUp))
	require.Equal(t, "metrics", m.Selected().Label, "up from the first goes to the last")
	m, _ = m.Update(press(tea.KeyDown))
	require.Equal(t, "server", m.Selected().Label, "down from the last, over the heading, to the first")
	n, _ := New(rows()).SetHeight(10).Update(press(tea.KeyUp))
	require.Equal(t, "server", n.Selected().Label, "without wrap it stops")
}

func TestPageAndEndKeysJump(t *testing.T) {
	m := New(numbered(20)).SetHeight(5)
	m, _ = m.Update(press(tea.KeyPgDown))
	require.Equal(t, "item-05", m.Selected().Label)
	m, _ = m.Update(press(tea.KeyEnd))
	require.Equal(t, "item-19", m.Selected().Label)
	m, _ = m.Update(press(tea.KeyHome))
	require.Equal(t, "item-00", m.Selected().Label)
}

func TestLegendListsTheKeysOfTheMode(t *testing.T) {
	keys := func(m Model[int]) []string {
		var out []string
		for _, e := range m.Legend() {
			out = append(out, e.Help().Key)
		}
		return out
	}
	require.Equal(t, []string{"↑/↓", "pgup/pgdn", "home/end", "enter", "/"}, keys(New(rows()).SetHeight(5)))
	require.Equal(t, []string{"↑/↓", "enter", "esc"}, keys(filtered(New(rows()).SetHeight(5), "x")))
}

// Fuzzy ranking over many rows, as each keystroke reruns it.
func BenchmarkFuzzyFilter(b *testing.B) {
	m := bigList(10_000).WithFuzzy(true)
	m, _ = m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m, _ = m.Update(tea.KeyPressMsg{Code: '7', Text: "7"})
	for b.Loop() {
		_ = m.View(th)
	}
}

func TestAClickMovesTheCursorAndASecondChooses(t *testing.T) {
	m := New(rows()).SetHeight(10) // ADDED, server, logging, spacer, AVAILABLE, cache, metrics
	m, act := m.Click(5)
	require.Equal(t, None, act)
	require.Equal(t, "cache", m.Selected().Label)
	_, act = m.Click(5)
	require.Equal(t, Chosen, act, "the row under the cursor is chosen")
	same, act := m.Click(0)
	require.Equal(t, None, act, "a heading is not a row to land on")
	require.Equal(t, "cache", same.Selected().Label)
	_, act = m.Click(40)
	require.Equal(t, None, act)

	titled := New(rows()).SetHeight(10).WithTitle("Plugins")
	titled, _ = titled.Click(3)
	require.Equal(t, "logging", titled.Selected().Label, "the title takes the first line")
}

func TestAClickPicksAMatchWhileFiltering(t *testing.T) {
	m := filtered(New(numbered(5)).SetHeight(10), "item")
	m, _ = m.Click(2)
	require.Equal(t, "item-02", m.Selected().Label)
	m, act := m.Click(2)
	require.Equal(t, Chosen, act)
	require.False(t, m.Filtering())
	require.Equal(t, "item-02", m.Selected().Label)
}

func TestWithStylesRecoloursOneList(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m := New(rows()).SetHeight(10).WithTitle("Plugins").WithStyles(func(s Styles) Styles { s.Title = red; return s })
	for _, under := range []theme.Resolved{th, theme.Resolve(theme.Mint, true)} {
		require.Equal(t, red.Render("Plugins"), strings.Split(m.View(under), "\n")[0], "the edit holds under any theme")
	}
}

func TestVimKeysMoveOnLettersButFilterTakesThemAsText(t *testing.T) {
	m := New(numbered(5)).SetHeight(5)
	n, _ := m.Update(typed("j"))
	require.Equal(t, m.Selected().Label, n.Selected().Label, "no letters by default")
	m = m.WithKeys(VimKeys())
	n, _ = m.Update(typed("j"))
	require.Equal(t, "item-01", n.Selected().Label)
	f := filtered(m, "j")
	require.Equal(t, "j", f.Filter(), "the prompt types the letter")
}
