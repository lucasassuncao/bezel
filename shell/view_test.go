package shell

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/theme"
)

func demoPanes() map[string]Pane {
	return map[string]Pane{
		"list":   {Title: "List", Body: func(r layout.Rect) string { return "item one\nitem two" }},
		"detail": {Title: "Detail", Body: func(r layout.Rect) string { return fmt.Sprintf("%dx%d", r.W, r.H) }},
	}
}

func lastRow(s Shell) string {
	rows := strings.Split(s.View(map[string]Pane{"list": {}, "detail": {}}), "\n")
	return ansi.Strip(rows[len(rows)-1])
}

func TestBodyIsTheTerminalMinusTheChrome(t *testing.T) {
	s := newShell(stubTab{name: "a", actions: []Action{Help()}})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	body := s.Body()
	require.Equal(t, 100, body.W)
	// header 1 + tabs 1 + blank 1 above; status 1 + legend 1 below.
	require.Equal(t, 3, body.Y)
	require.Equal(t, 30-3-2, body.H)

	require.Equal(t, 33, s.Rect("list").W)
	require.Equal(t, body.H, s.Rect("list").H)
	require.Equal(t, 100, s.Rect("list").W+s.Rect("detail").W)
}

func TestNoTabsMeansNoTabRow(t *testing.T) {
	s := newShell()
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	require.Equal(t, 2, s.Body().Y) // the header, then the blank row above the panes
	plain := strings.Split(ansi.Strip(s.View(demoPanes())), "\n")
	require.Empty(t, strings.TrimSpace(plain[1]))
}

func TestViewIsExactlyTheTerminal(t *testing.T) {
	s := newShell(stubTab{name: "a", actions: []Action{Help(), Quit()}})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	out := s.View(demoPanes())
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
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	out := ansi.Strip(s.View(demoPanes()))
	inner := draw.InnerRect(s.Rect("detail"))
	require.Contains(t, out, fmt.Sprintf("%dx%d", inner.W, inner.H))
}

func TestAMissingPaneVanishesAndItsSpaceIsReused(t *testing.T) {
	s := newShell()
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	panes := demoPanes()
	delete(panes, "detail")
	s2 := s.withPresent(panes)
	require.Equal(t, 80, s2.Rect("list").W)
	require.Equal(t, layout.Rect{}, s2.Rect("detail"))
	require.NotContains(t, ansi.Strip(s.View(panes)), "Detail")
}

func TestOverlayFloatsOverThePanes(t *testing.T) {
	s := newShell()
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	s = s.Push(overlay.NewAlert(overlay.Info, "Saved", "written", lipgloss.NewStyle().Border(lipgloss.NormalBorder()), legend.Style{}))
	out := ansi.Strip(s.View(demoPanes()))
	require.Contains(t, out, "written")
	require.Contains(t, out, "item one") // background still there
	require.Contains(t, out, "OK", "the alert shows how it closes, in its box")
}

func TestStatusRowShowsInputThenBannerThenTabStatus(t *testing.T) {
	s := newShell(stubTab{name: "a"})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	require.Contains(t, ansi.Strip(s.View(demoPanes())), "status of a")

	s, _ = s.SetStatus("saved!", OK, 0)
	require.Contains(t, ansi.Strip(s.View(demoPanes())), "saved!")

	in := textinput.New()
	in.Prompt = ":"
	in.SetValue("cmd")
	s = s.SetInput(&in)
	plain := ansi.Strip(s.View(demoPanes()))
	require.Contains(t, plain, ":cmd")
	require.NotContains(t, plain, "saved!")
}

func TestViewSurvivesATinyTerminal(t *testing.T) {
	s := newShell(stubTab{name: "a"})
	for _, size := range [][2]int{{0, 0}, {1, 1}, {5, 3}, {20, 4}, {200, 2}} {
		s, _, _ = s.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		require.NotPanics(t, func() { s.View(demoPanes()) }, "%v", size)
	}
}

func TestInfoStatusUsesTheNeutralStyleAndHelpHasNoEmptySection(t *testing.T) {
	s := newShell()
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	s, _ = s.SetStatus("just saying", Info, 0)
	require.Contains(t, ansi.Strip(s.View(demoPanes())), "just saying")

	s = New(Config{Layout: twoPane(), Theme: theme.Resolve(theme.ThemePlain, true)})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	s = s.ShowHelp()
	require.NotContains(t, ansi.Strip(s.View(demoPanes())), "Keys")
}

// "?" lists only what has a key; a keyless command lives in the palette alone.
func TestHelpListsOnlyActionsWithAKey(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Help(), Custom("r", "reveal", nil), Command("goto", "jump to a path", nil)))
	s, _, _ = s.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
	screen := ansi.Strip(s.View(map[string]Pane{"list": {}, "detail": {}}))
	require.Contains(t, screen, "reveal")
	require.NotContains(t, screen, "jump to a path")
}

func TestLegendPinsTheShellHelpKey(t *testing.T) {
	s := newShell(stubTab{name: "a", actions: []Action{Custom("x", "do x", nil), Help()}})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	require.Contains(t, lastRow(s), "[?] help")
}

// No overlay changes the legend: it is always the screen's. A modal draws its
// own keys inside its box.
func TestNoOverlayChangesTheLegend(t *testing.T) {
	st := lipgloss.NewStyle()
	for name, o := range map[string]overlay.Overlay{
		"alert":   overlay.NewAlert(overlay.Info, "t", "m", st, legend.Style{}),
		"text":    overlay.NewText("t", []string{"body"}, st, legend.Style{}),
		"confirm": overlay.NewConfirm("t", "m", nil, st, legend.Style{}),
		"pick":    overlay.NewPick("t", []string{"a", "b"}, nil, nil, st, legend.Style{}),
		"prompt":  overlay.NewPrompt("t", nil, st, legend.Style{}),
	} {
		s := newShell(stubTab{name: "a", actions: []Action{Custom("x", "do x", nil), Help()}})
		s, _, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		before := lastRow(s)
		require.Equal(t, before, lastRow(s.Push(o)), name)
	}
	s := newShell(stubTab{name: "a", actions: []Action{Custom("x", "do x", nil), Help()}})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	require.Equal(t, lastRow(s), lastRow(s.ShowHelp()), "help")
	require.Equal(t, lastRow(s), lastRow(s.OpenCommands()), "command palette")
}

// An app that opens its own help still pins "?" through a display-only Help.
func TestADisplayOnlyHelpIsPinnedButNotHandled(t *testing.T) {
	s := newShell(stubTab{name: "a", actions: []Action{Help(DisplayOnly())}})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	require.Contains(t, lastRow(s), "[?] help")
	_, handled, _ := s.Update(press("?"))
	require.False(t, handled)
}

func TestLegendDropsHelpWhileAnInputHasTheKeys(t *testing.T) {
	s := newShell(stubTab{name: "a", actions: []Action{Help()}})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	in := textinput.New()
	s = s.SetInput(&in)
	require.NotContains(t, lastRow(s), "[?] help")
}

// A blank row separates the tab strip from the panes below it.
func TestABlankRowSitsBetweenTheTabsAndThePanes(t *testing.T) {
	s := newShell(stubTab{name: "a"}, stubTab{name: "b"})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	rows := strings.Split(ansi.Strip(s.View(map[string]Pane{"list": {Title: "L"}, "detail": {Title: "D"}})), "\n")
	require.Contains(t, rows[1], "a")
	require.Empty(t, strings.TrimSpace(rows[2]))
	require.Contains(t, rows[3], "L")
}

// The status row starts at the left edge, in line with the legend under it.
func TestTheStatusRowIsFlushLeft(t *testing.T) {
	s := newShell(stubTab{name: "a", actions: []Action{Help()}})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	for _, lvl := range []Level{Info, OK, Error} {
		s, _ = s.SetStatus("copied: kv/app", lvl, 0)
		rows := strings.Split(ansi.Strip(s.View(map[string]Pane{"list": {}, "detail": {}})), "\n")
		require.True(t, strings.HasPrefix(rows[len(rows)-2], "copied: kv/app"), "level %d: %q", lvl, rows[len(rows)-2])
	}
}
