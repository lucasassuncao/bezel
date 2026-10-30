package shell

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/input"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/theme"
)

func demoPanes() map[string]Pane {
	return map[string]Pane{
		"list":   {Title: "List", Body: func(r layout.Rect) string { return "item one\nitem two" }},
		"detail": {Title: "Detail", Body: func(r layout.Rect) string { return fmt.Sprintf("%dx%d", r.W, r.H) }},
	}
}

// legendRow is the legend's first row, right under the status row.
func legendRow(s Shell) string {
	rows := strings.Split(s.render(map[string]Pane{"list": {}, "detail": {}}), "\n")
	return ansi.Strip(rows[statusAt(s)+1])
}

// statusAt is the index of the status row, the first under the panes.
func statusAt(s Shell) int { return s.Body().Y + s.Body().H }

func TestBodyIsTheTerminalMinusTheChrome(t *testing.T) {
	s := newShell(stubTab{name: "a", actions: []Action{Help()}})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	body := s.Body()
	require.Equal(t, 100, body.W)
	// header 1 + tabs 1 + blank 1 above; status 1 + legend 2 below.
	require.Equal(t, 3, body.Y)
	require.Equal(t, 30-3-3, body.H)

	require.Equal(t, 33, s.Rect("list").W)
	require.Equal(t, body.H, s.Rect("list").H)
	require.Equal(t, 100, s.Rect("list").W+s.Rect("detail").W)
}

func TestNoTabsMeansNoTabRow(t *testing.T) {
	s := newShell()
	s, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	require.Equal(t, 2, s.Body().Y) // the header, then the blank row above the panes
	plain := strings.Split(ansi.Strip(s.render(demoPanes())), "\n")
	require.Empty(t, strings.TrimSpace(plain[1]))
}

func TestViewIsExactlyTheTerminal(t *testing.T) {
	s := newShell(stubTab{name: "a", actions: []Action{Help(), Quit()}})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	out := s.render(demoPanes())
	lines := strings.Split(out, "\n")
	require.Len(t, lines, 24)
	for i, l := range lines {
		require.Equal(t, 80, ansi.StringWidth(l), "row %d: %q", i, l)
	}
	plain := ansi.Strip(out)
	require.Contains(t, plain, "demo")     // header
	require.Contains(t, plain, " a ")      // tab
	require.Contains(t, plain, "List")     // panel title
	require.Contains(t, plain, "item one") // pane body
	require.Contains(t, plain, "status of a in list")
	require.Contains(t, plain, "[q] quit") // legend
}

func TestPaneBodyGetsTheInnerRect(t *testing.T) {
	s := newShell()
	s, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	out := ansi.Strip(s.render(demoPanes()))
	inner := draw.InnerRect(s.Rect("detail"))
	require.Contains(t, out, fmt.Sprintf("%dx%d", inner.W, inner.H))
}

func TestAMissingPaneVanishesAndItsSpaceIsReused(t *testing.T) {
	s := newShell()
	s, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	panes := demoPanes()
	delete(panes, "detail")
	s2 := s.withPresent(panes)
	require.Equal(t, 80, s2.Rect("list").W)
	require.Equal(t, layout.Rect{}, s2.Rect("detail"))
	require.NotContains(t, ansi.Strip(s.render(panes)), "Detail")
}

func TestOverlayFloatsOverThePanes(t *testing.T) {
	s := newShell()
	s, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	s = s.Push(overlay.NewAlert(draw.Info, "Saved", "written", theme.Resolved{Modal: lipgloss.NewStyle().Border(lipgloss.NormalBorder())}))
	out := ansi.Strip(s.render(demoPanes()))
	require.Contains(t, out, "written")
	require.Contains(t, out, "item one") // background still there
	require.Contains(t, out, "OK", "the alert shows how it closes, in its box")
}

func TestStatusRowShowsInputThenBannerThenTabStatus(t *testing.T) {
	s := newShell(stubTab{name: "a"})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	require.Contains(t, ansi.Strip(s.render(demoPanes())), "status of a")

	s, _ = s.SetStatus("saved!", OK, 0)
	require.Contains(t, ansi.Strip(s.render(demoPanes())), "saved!")

	in := input.New(theme.Resolved{}).WithPrompt(":").SetValue("cmd")
	s = s.SetInput(in)
	plain := ansi.Strip(s.render(demoPanes()))
	require.Contains(t, plain, ":cmd")
	require.NotContains(t, plain, "saved!")
}

func TestViewSurvivesATinyTerminal(t *testing.T) {
	s := newShell(stubTab{name: "a"})
	for _, size := range [][2]int{{0, 0}, {1, 1}, {5, 3}, {20, 4}, {200, 2}} {
		s, _ = s.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		require.NotPanics(t, func() { s.render(demoPanes()) }, "%v", size)
	}
}

func TestInfoStatusUsesTheNeutralStyleAndHelpHasNoEmptySection(t *testing.T) {
	s := newShell()
	s, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	s, _ = s.SetStatus("just saying", Info, 0)
	require.Contains(t, ansi.Strip(s.render(demoPanes())), "just saying")

	s = New(Config{Layout: twoPane(), Theme: theme.Resolve(theme.Plain(), true)})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	s = s.ShowHelp()
	require.NotContains(t, ansi.Strip(s.render(demoPanes())), "Keys")
}

// "?" lists only what has a key; a keyless command lives in the palette alone.
func TestHelpListsOnlyActionsWithAKey(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Help(), Custom("r", "reveal", nil), Command("goto", "jump to a path", nil)))
	s, _ = s.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
	screen := ansi.Strip(s.render(map[string]Pane{"list": {}, "detail": {}}))
	require.Contains(t, screen, "reveal")
	require.NotContains(t, screen, "jump to a path")
}

func TestLegendPinsTheShellHelpKey(t *testing.T) {
	s := newShell(stubTab{name: "a", actions: []Action{Custom("x", "do x", nil), Help()}})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	require.Contains(t, legendRow(s), "[?] help")
}

// No overlay changes the legend: it is always the screen's. A modal draws its
// own keys inside its box.
func TestNoOverlayChangesTheLegend(t *testing.T) {
	st := lipgloss.NewStyle()
	for name, o := range map[string]overlay.Overlay{
		"alert":   overlay.NewAlert(draw.Info, "t", "m", theme.Resolved{Modal: st}),
		"text":    overlay.NewText("t", "body", theme.Resolved{Modal: st}),
		"confirm": overlay.NewConfirm("t", "m", nil, theme.Resolved{Modal: st}),
		"pick":    overlay.NewPick("t", []string{"a", "b"}, nil, theme.Resolved{Modal: st}),
		"prompt":  overlay.NewPrompt("t", nil, theme.Resolved{Modal: st}),
	} {
		s := newShell(stubTab{name: "a", actions: []Action{Custom("x", "do x", nil), Help()}})
		s, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		before := legendRow(s)
		require.Equal(t, before, legendRow(s.Push(o)), name)
	}
	s := newShell(stubTab{name: "a", actions: []Action{Custom("x", "do x", nil), Help()}})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	require.Equal(t, legendRow(s), legendRow(s.ShowHelp()), "help")
	require.Equal(t, legendRow(s), legendRow(s.OpenCommands()), "command palette")
}

// An app that opens its own help still pins "?" through a display-only Help.
func TestADisplayOnlyHelpIsPinnedButNotHandled(t *testing.T) {
	s := newShell(stubTab{name: "a", actions: []Action{Help(DisplayOnly())}})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	require.Contains(t, legendRow(s), "[?] help")
	_, r := s.Update(press("?"))
	require.False(t, r.Handled)
}

func TestLegendDropsHelpWhileAnInputHasTheKeys(t *testing.T) {
	s := newShell(stubTab{name: "a", actions: []Action{Help()}})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	in := input.New(theme.Resolved{})
	s = s.SetInput(in)
	require.NotContains(t, legendRow(s), "[?] help")
}

// A blank row separates the tab strip from the panes below it.
func TestABlankRowSitsBetweenTheTabsAndThePanes(t *testing.T) {
	s := newShell(stubTab{name: "a"}, stubTab{name: "b"})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	rows := strings.Split(ansi.Strip(s.render(map[string]Pane{"list": {Title: "L"}, "detail": {Title: "D"}})), "\n")
	require.Contains(t, rows[1], "a")
	require.Empty(t, strings.TrimSpace(rows[2]))
	require.Contains(t, rows[3], "L")
}

// The status row starts at the left edge, in line with the legend under it.
func TestTheStatusRowIsFlushLeft(t *testing.T) {
	s := newShell(stubTab{name: "a", actions: []Action{Help()}})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	for _, lvl := range []Level{Info, OK, Error} {
		s, _ = s.SetStatus("copied: kv/app", lvl, 0)
		rows := strings.Split(ansi.Strip(s.render(map[string]Pane{"list": {}, "detail": {}})), "\n")
		require.True(t, strings.HasPrefix(rows[statusAt(s)], "copied: kv/app"), "level %d: %q", lvl, rows[statusAt(s)])
	}
}

// openSGR is the sequence that starts text in st, to find where it was used.
func openSGR(st lipgloss.Style) string { return strings.Split(st.Render("x"), "x")[0] }

func TestAPaneCarriesRightTextAndItsOwnBorderColour(t *testing.T) {
	s := newShell()
	s, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	red := lipgloss.Color("#FF0000")

	plain := strings.Split(s.render(demoPanes()), "\n")[s.Body().Y]
	out := strings.Split(s.render(map[string]Pane{
		"list":   {Title: "List"},
		"detail": {Title: "Detail", Right: "3 items", Accent: red},
	}), "\n")[s.Body().Y]

	require.Contains(t, ansi.Strip(out), "3 items")
	require.Equal(t, 100, ansi.StringWidth(out))
	require.Contains(t, out, openSGR(lipgloss.NewStyle().Foreground(red)))
	require.NotContains(t, plain, openSGR(lipgloss.NewStyle().Foreground(red)))
}

func TestAPaneCarriesABottomTextAndAScrollbar(t *testing.T) {
	s := newShell()
	s, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	r := s.Rect("detail")
	lines := strings.Split(ansi.Strip(s.render(map[string]Pane{
		"list":   {Title: "List"},
		"detail": {Title: "Detail", Bottom: "1/40", Scroll: layout.Scroll{Visible: 5, Total: 40}},
	})), "\n")
	cell := func(row int) string { return ansi.TruncateLeft(ansi.Truncate(lines[row], r.X+r.W, ""), r.X+r.W-1, "") }

	require.True(t, strings.HasSuffix(ansi.Truncate(lines[r.Y+r.H-1], r.X+r.W, ""), " 1/40 ─╯"))
	require.Equal(t, "█", cell(r.Y+1), "the thumb at the top of the track")
	require.Equal(t, "│", cell(r.Y+r.H-2), "the track below it")
}

// quietTab has nothing to say on the status row.
type quietTab struct{ stubTab }

func (quietTab) Status(Context) string { return "" }

// The Toast style gives the row back to the panes until something is typed
// there, Busy runs or the tab has a status, and shows a banner as a toast.
func TestTheToastStyleShowsTheRowOnlyWhenItHasALine(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1") // a toast shows at once, without frames
	th := theme.Resolve(theme.Plain(), true)
	size := tea.WindowSizeMsg{Width: 120, Height: 30}
	s, _ := New(Config{Layout: twoPane(), Tabs: []Tab{quietTab{stubTab{name: "a"}}}, Theme: th, StatusStyle: Toast}).Update(size)
	bar, _ := New(Config{Layout: twoPane(), Tabs: []Tab{quietTab{stubTab{name: "a"}}}, Theme: th}).Update(size)
	require.Equal(t, bar.Body().H+1, s.Body().H, "no row: the panes take it")

	s, cmd := s.SetStatus("saved notes.md", OK, time.Minute)
	require.NotNil(t, cmd)
	rows := strings.Split(ansi.Strip(s.render(demoPanes())), "\n")
	at := slices.IndexFunc(rows, func(r string) bool { return strings.Contains(r, "saved notes.md") })
	require.Positive(t, at, "the banner is a toast")
	require.Less(t, at, 5, "at the top")
	require.Greater(t, strings.Index(rows[at], "saved"), 60, "on the right")
	require.Equal(t, "saved notes.md", s.Status(), "Status still reads it")

	typing := s.SetInput(input.New(th).SetValue("kv/"))
	require.Equal(t, bar.Body().H, typing.Body().H, "typing brings the row back")
	busy, _ := s.Busy("indexing")
	require.Equal(t, bar.Body().H, busy.Body().H, "and so does Busy")
	talking, _ := New(Config{Layout: twoPane(), Tabs: []Tab{stubTab{name: "a"}}, Theme: th, StatusStyle: Toast}).Update(size)
	require.Equal(t, bar.Body().H, talking.Body().H, "and the tab's status")
}

// A toast's frames and expiry are the shell's, and a harness drops them like
// the status timer.
func TestAToastsMessagesAreTheShells(t *testing.T) {
	s, _ := New(Config{Layout: twoPane(), Theme: theme.Resolve(theme.Plain(), true), StatusStyle: Toast}).
		Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	s, _ = s.SetStatus("x", Info, time.Minute)
	frame := toastMsg{s.toasts.FrameAt(time.Now())}
	require.True(t, IsStatusExpiry(frame))
	_, r := s.Update(frame)
	require.True(t, r.Handled)
}

// The program's view paints the theme's own background and text colours.
func TestViewCarriesTheThemesColours(t *testing.T) {
	th := theme.Resolve(theme.Mint(), true)
	v := New(Config{Layout: twoPane(), Theme: th}).View(demoPanes())
	require.Equal(t, th.Background, v.BackgroundColor)
	require.Equal(t, th.Foreground, v.ForegroundColor)
}
