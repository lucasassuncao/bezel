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
	p = logged(sized(NewPane(th).WithStyles(func(s Styles) Styles { s.Info = red; return s })), 1)
	require.Contains(t, p.Lines()[0], red.Render("INFO"))
}

func TestPaneAttributesBecomeLinesOfTheirOwn(t *testing.T) {
	p := sized(NewPane(paneTheme())).SetHeight(5)
	slog.New(p.Handler().WithTimeFormat("")).Info("saved", "file", "a.txt")
	p, _ = p.Update(p.Init()())
	require.Equal(t, 2, p.Len())
	require.Len(t, p.Legend(), 2, "scroll and page: panning is left out while wrapping")
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
