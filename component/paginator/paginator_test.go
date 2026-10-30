package paginator

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Mint(), true)

func TestNewHasOnePageUntilItKnowsTheItems(t *testing.T) {
	p := New(10)
	require.Equal(t, 1, p.Pages())
	require.Equal(t, 0, p.Page())
	require.True(t, p.OnFirst() && p.OnLast())
}

func TestSetItemsRoundsThePagesUp(t *testing.T) {
	require.Equal(t, 3, New(10).SetItems(25).Pages())
	require.Equal(t, 3, New(10).SetItems(30).Pages())
	require.Equal(t, 1, New(10).SetItems(0).Pages())
}

func TestAPerPageBelowOneIsOne(t *testing.T) {
	require.Equal(t, 1, New(0).PerPage())
	require.Equal(t, 1, New(-3).PerPage())
}

func TestBoundsCutTheCurrentPage(t *testing.T) {
	p := New(10).SetItems(25).SetPage(2)
	start, end := p.Bounds(25)
	require.Equal(t, []int{20, 25}, []int{start, end})
	require.Equal(t, 5, p.ItemsOnPage(25))
}

func TestBoundsNeverRunPastTheItems(t *testing.T) {
	start, end := New(10).SetItems(25).SetPage(2).Bounds(5)
	require.Equal(t, []int{5, 5}, []int{start, end})
}

func TestSetItemsKeepsThePageInRange(t *testing.T) {
	require.Equal(t, 1, New(10).SetItems(50).SetPage(4).SetItems(15).Page())
}

func TestPrevAndNextStopAtTheEnds(t *testing.T) {
	p := New(10).SetItems(20)
	require.Equal(t, 0, p.Prev().Page())
	require.Equal(t, 1, p.Next().Next().Page())
}

func TestTheKeysTurnThePage(t *testing.T) {
	p, r := New(10).SetItems(30).Update(bezeltest.Key("right"))
	require.Equal(t, 1, p.Page())
	require.Equal(t, Paged{Page: 1}, r.Event)
	p, _ = p.Update(bezeltest.Key("pgdown"))
	p, r = p.Update(bezeltest.Key("pgdown"))
	require.Equal(t, 2, p.Page())
	require.True(t, r.Handled, "the key is the paginator's even on the last page")
	require.Nil(t, r.Event, "already on the last page")
	p, r = p.Update(bezeltest.Key("left"))
	require.Equal(t, 1, p.Page())
	require.Equal(t, Paged{Page: 1}, r.Event)
}

func TestNextReportsPaged(t *testing.T) {
	_, r := New(10).SetItems(30).Update(bezeltest.Key("pgdown"))
	require.Equal(t, bezel.Result{Handled: true, Event: Paged{Page: 1}}, r)
}

func TestForeignKeyIsNotHandled(t *testing.T) {
	_, r := New(10).SetItems(30).Update(bezeltest.Key("x"))
	require.Equal(t, bezel.Result{}, r)
	_, r = New(10).SetItems(30).Update(tea.WindowSizeMsg{Width: 1, Height: 1})
	require.Equal(t, bezel.Result{}, r)
}

func TestNumbersReadPageOfPages(t *testing.T) {
	p := New(10).SetItems(30).SetPage(1)
	require.Equal(t, "2/3", ansi.Strip(p.View(th)))
	require.Equal(t, "page 2 of 3", ansi.Strip(p.WithFormat("page %d of %d").View(th)))
}

func TestDotsMarkTheCurrentPage(t *testing.T) {
	p := New(10).SetItems(30).SetPage(1).WithDisplay(Dots)
	require.Equal(t, "○●○", ansi.Strip(p.WithIcons(icon.Unicode()).View(th)))
	require.Equal(t, "o*o", ansi.Strip(p.WithIcons(icon.ASCII()).View(th)))
}

func TestDotsAreStyledFromTheTheme(t *testing.T) {
	p := New(10).SetItems(20).WithDisplay(Dots).WithIcons(icon.Unicode())
	require.Equal(t, th.Accent.Render("●")+th.Muted.Render("○"), p.View(th))
}

func TestWithStylesHoldsUnderAnotherTheme(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	p := New(10).WithStyles(func(s Styles) Styles { s.Numbers = red; return s })
	require.Equal(t, red.Render("1/1"), p.View(theme.Resolve(theme.Plain(), true)))
}

func TestLegendNamesThePageKeysOnce(t *testing.T) {
	es := New(1).Legend()
	require.Len(t, es, 1)
	require.Equal(t, key.Help{Key: "←/→", Desc: "page"}, es[0].Help())
}

func TestAZeroPaginatorIsOnePageOfOne(t *testing.T) {
	var p Model
	require.Equal(t, 1, p.Pages())
	require.Equal(t, 1, p.PerPage())
	require.NotPanics(t, func() { p = p.SetItems(10) })
	require.Equal(t, 10, p.Pages())
	require.Equal(t, "1/10", ansi.Strip(p.View(th)))
}

func TestSetPerPageKeepsTheFirstItemInView(t *testing.T) {
	p := New(10).SetItems(100).SetPage(3).SetPerPage(25)
	require.Equal(t, 4, p.Pages())
	require.Equal(t, 1, p.Page(), "item 30 sits on the second page of 25")
	require.Equal(t, 1, New(10).SetPerPage(0).PerPage())
}

func TestTheDefaultDotsFollowTheEnvironment(t *testing.T) {
	t.Setenv("BEZEL_ASCII", "1")
	require.Equal(t, "*oo", ansi.Strip(New(10).SetItems(30).WithDisplay(Dots).View(th)))
}
