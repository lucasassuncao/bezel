package logger

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

// logged is p after n records logged through its Handler reach Update.
func logged(p Pane, n int) Pane {
	log := slog.New(p.Handler().WithTimeFormat(""))
	for i := range n {
		log.Info(fmt.Sprintf("line %d", i))
	}
	p, _ = p.Update(p.Init()())
	return p
}

func shown(p Pane) []string {
	return strings.Split(ansi.Strip(p.View()), "\n")
}

func sized(p Pane) Pane { return p.SetWidth(20).SetHeight(3) }

func TestPaneFollowsTheEndAsLinesArrive(t *testing.T) {
	p := logged(sized(NewPane(paneTheme())), 5)
	require.Equal(t, 5, p.Len())
	require.True(t, p.Following())
	got := shown(p)
	require.Len(t, got, 3)
	require.Equal(t, "INFO  line 4", strings.TrimRight(got[2], " "), "the newest line is the last one shown")
}

func TestPaneHoldsStillWhileScrolledUpAndCountsWhatArrives(t *testing.T) {
	p := logged(sized(NewPane(paneTheme())), 5)
	p, r := p.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	require.True(t, r.Handled)
	require.False(t, p.Following())
	top := p.Scroll().Offset

	p = logged(p, 2)
	require.Equal(t, top, p.Scroll().Offset, "the reader's place holds")
	got := shown(p)
	require.Len(t, got, 3, "the count takes a viewport line")
	require.Equal(t, "↓ 2 new lines", strings.TrimRight(got[2], " "))

	p, _ = p.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	require.True(t, p.Following(), "the end follows again")
	require.Equal(t, "INFO  line 1", strings.TrimRight(shown(p)[2], " "), "the second batch counts from 0 again")
}

func TestPaneDropsTheOldestPastItsLimit(t *testing.T) {
	p := logged(sized(NewPane(paneTheme()).WithLimit(3)), 5)
	require.Equal(t, 3, p.Len())
	require.Contains(t, ansi.Strip(p.Lines()[0]), "line 2")
	require.Equal(t, 1, NewPane(paneTheme()).WithLimit(0).limit)
}

func TestPaneTakesOnlyItsOwnLines(t *testing.T) {
	a, b := NewPane(paneTheme()), NewPane(paneTheme())
	slog.New(a.Handler()).Info("hi")
	msg := a.Init()()
	_, r := b.Update(msg)
	require.Equal(t, bezel.Result{}, r)
	a, r = a.Update(msg)
	require.True(t, r.Handled)
	require.NotNil(t, r.Cmd, "it waits for the next lines")
	require.Equal(t, 1, a.Len())
}

func TestPaneLinesKeepTheLevelColour(t *testing.T) {
	th := paneTheme()
	p := logged(sized(NewPane(th)), 1)
	require.Contains(t, p.Lines()[0], DefaultStyles(th).Info.Render("INFO"))
	red := th.Danger
	p = logged(sized(NewPane(th).WithStyles(func(s PaneStyles) PaneStyles { s.Lines.Info = red; return s })), 1)
	require.Contains(t, p.Lines()[0], red.Render("INFO"))
}

func TestPaneAttributesBecomeLinesOfTheirOwn(t *testing.T) {
	p := sized(NewPane(paneTheme())).SetHeight(5)
	slog.New(p.Handler().WithTimeFormat("")).Info("saved", "file", "a.txt")
	p, _ = p.Update(p.Init()())
	require.Equal(t, 2, p.Len())
	require.Len(t, p.Legend(), 3, "scroll, page and filter: panning is left out while wrapping")
}

func TestPaneViewIsEmptyBeforeSized(t *testing.T) {
	require.Empty(t, logged(NewPane(paneTheme()), 2).View())
}

func paneTheme() theme.Resolved { return theme.Resolve(theme.Mint(), true) }

// Goroutines log at once; every line reaches the pane, none twice.
func TestPaneTakesLinesFromManyGoroutines(t *testing.T) {
	p := NewPane(paneTheme())
	log := slog.New(p.Handler())
	done := make(chan struct{})
	for g := range 4 {
		go func() {
			for i := range 50 {
				log.Info(fmt.Sprintf("g%d-%d", g, i))
			}
			done <- struct{}{}
		}()
	}
	for range 4 {
		<-done
	}
	for p.Len() < 200 {
		p, _ = p.Update(p.Init()())
	}
	require.Equal(t, 200, p.Len())
}

func TestPaneTakesAnyHandlerThroughItsWriter(t *testing.T) {
	p := sized(NewPane(paneTheme())).SetWidth(80)
	slog.New(slog.NewJSONHandler(p.Writer(), nil)).Info("saved", "file", "a.txt")
	p, _ = p.Update(p.Init()())
	require.Equal(t, 1, p.Len(), "one JSON record, one line")
	require.Contains(t, p.Lines()[0], `"msg":"saved","file":"a.txt"}`)
}

// typed sends each rune of s as a key press.
func typed(p Pane, s string) Pane {
	for _, r := range s {
		p, _ = p.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return p
}

// deployed is a pane with three records, one of them two lines long.
func deployed() Pane {
	p := sized(NewPane(paneTheme())).SetWidth(40).SetHeight(6)
	log := slog.New(p.Handler().WithTimeFormat(""))
	log.Info("pulling image")
	log.Error("refused", "host", "db-primary")
	log.Info("replica ready")
	p, _ = p.Update(p.Init()())
	return p
}

func TestPaneFilterShowsTheRecordsHoldingTheText(t *testing.T) {
	p, r := deployed().Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	require.True(t, r.Handled)
	require.True(t, p.Filtering())

	p = typed(p, "DB")
	require.Equal(t, "DB", p.Filter())
	require.Equal(t, 2, p.Shown(), "the whole record, its attribute line too")
	require.Equal(t, 4, p.Len())
	got := shown(p)
	require.Contains(t, got[0], "refused", "a match in an attribute keeps the record's head")
	require.Equal(t, "/DB▋", strings.TrimRight(got[len(got)-1], " "))
	require.Len(t, p.Legend(), 2, "enter and esc while typing")

	p, _ = p.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	require.Equal(t, "D", p.Filter())
}

func TestPaneFilterKeptByEnterTakesNewLinesAndEscClearsIt(t *testing.T) {
	p := typed(deployed(), "/ready")
	p, _ = p.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.False(t, p.Filtering())
	require.Equal(t, "ready", p.Filter())
	require.Equal(t, "/ready", strings.TrimRight(shown(p)[len(shown(p))-1], " "))

	log := slog.New(p.Handler().WithTimeFormat(""))
	log.Info("cache ready")
	log.Info("pulling image")
	p, _ = p.Update(p.Init()())
	require.Equal(t, 2, p.Shown(), "only the new line that matches shows")
	require.Equal(t, 6, p.Len())

	p, r := p.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.True(t, r.Handled)
	require.Empty(t, p.Filter())
	require.Equal(t, 6, p.Shown())
}

func TestPaneEscWithoutAFilterIsNotItsOwn(t *testing.T) {
	_, r := deployed().Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.False(t, r.Handled)
}

func TestPaneFilterPaintsItsMatches(t *testing.T) {
	p := typed(deployed(), "/image")
	require.Contains(t, p.View(), paneTheme().Match.Render("image"), "no match is current")
	require.Equal(t, [][]int{{0, 2}, {3, 5}}, matches("Ab ab", "AB"))
	require.Nil(t, matches("İ", "i"), "lower case changes the length")
}

func TestPaneLimitCutsIntoTheOldestRecord(t *testing.T) {
	p := deployed().WithLimit(2)
	require.Equal(t, 2, p.Len())
	require.Contains(t, ansi.Strip(p.Lines()[0]), "host=db-primary", "the record's head went first")
}

func TestPaneStylesPaintItsOwnParts(t *testing.T) {
	th := paneTheme()
	red := th.Danger
	p := typed(deployed().WithStyles(func(s PaneStyles) PaneStyles { s.Prompt, s.Match = red, red; return s }), "/image")
	require.Contains(t, p.View(), red.Render("image"))
	require.Contains(t, p.View(), red.Render(draw.Fit("/image▋", 40)))
}

func TestPaneWithMatchFiltersByItsOwnRule(t *testing.T) {
	exact := func(q, text string) bool { return strings.Contains(text, q) }
	p := typed(deployed().WithMatch(exact), "/IMAGE")
	require.Zero(t, p.Shown(), "the case counts under this match")
	p = typed(deployed().WithMatch(exact), "/image")
	require.Equal(t, 1, p.Shown())
	require.NotContains(t, p.View(), paneTheme().Match.Render("image"), "only Contains paints")
}

// A held pane keeps its line when the limit drops lines that wrap.
func TestPaneHeldKeepsItsLineWhileWrappedLinesDrop(t *testing.T) {
	p := NewPane(paneTheme()).WithLimit(6).SetWidth(10).SetHeight(3)
	w := p.Writer()
	for i := range 6 {
		fmt.Fprintf(w, "line %d %s\n", i, strings.Repeat("x", 12)) // two rows each
	}
	p, _ = p.Update(p.Init()())
	p, _ = p.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
	require.False(t, p.Following())
	top := strings.TrimRight(shown(p)[0], " ")

	fmt.Fprintf(w, "line 6 %s\n", strings.Repeat("x", 12))
	p, _ = p.Update(p.Init()())
	require.Equal(t, top, strings.TrimRight(shown(p)[0], " "), "the same row stays on top")
}

func TestPaneSetThemeRepaintsWhatComesNext(t *testing.T) {
	mint, dracula := theme.Resolve(theme.Mint(), true), theme.Resolve(mustTheme(t, "dracula"), true)
	p := NewPane(mint).WithStyles(func(s PaneStyles) PaneStyles { s.Filter = s.Filter.Bold(true); return s })
	require.NotEqual(t, DefaultPaneStyles(dracula).Match.GetForeground(), p.styles().Match.GetForeground())
	p = p.SetTheme(dracula)
	require.Equal(t, DefaultPaneStyles(dracula).Match.GetForeground(), p.styles().Match.GetForeground())
	require.True(t, p.styles().Filter.GetBold(), "the edit is kept")
	p = logged(sized(p), 1)
	require.Contains(t, p.Lines()[0], DefaultStyles(dracula).Info.Render("INFO"))
}

func mustTheme(t *testing.T, name string) theme.Theme {
	t.Helper()
	th, err := theme.Lookup(name)
	require.NoError(t, err)
	return th
}
