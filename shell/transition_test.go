package shell

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/effect"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

// sliding is a two-tab shell, sized, that slides between its tabs.
func sliding(t *testing.T) Shell {
	t.Helper()
	t.Setenv("BEZEL_NO_MOTION", "")
	t.Setenv("TERM", "xterm-256color")
	s := New(Config{
		Layout: layout.Rows(layout.Fill("main")), Theme: theme.Resolve(theme.Plain, true),
		Tabs:       []Tab{stubTab{name: "one"}, stubTab{name: "two"}},
		Actions:    []Action{ChangeTab()},
		Transition: SlideTabs(100 * time.Millisecond),
	})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	return s
}

// panes draws the active tab's name in the one pane.
func panes(s Shell) map[string]Pane {
	name := s.cfg.Tabs[s.ActiveTab()].Name()
	return map[string]Pane{"main": {Title: "main", Body: func(layout.Rect) string { return "content of " + name }}}
}

func TestChangingTabPlaysTheTransition(t *testing.T) {
	s := sliding(t)
	s, handled, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.True(t, handled)
	require.NotNil(t, cmd, "the transition asks for its first frame")
	require.True(t, s.Transition().Active())
	require.NotContains(t, ansi.Strip(s.View(panes(s))), "content of two", "before its first frame the new tab is off screen")
	require.Contains(t, ansi.Strip(s.View(panes(s))), "two", "the tab strip itself is never moved")

	at := time.Unix(1000, 0)
	s, handled, _ = s.Update(s.Transition().FrameAt(at))
	require.True(t, handled, "the shell takes its own frames")
	s, _, cmd = s.Update(s.Transition().FrameAt(at.Add(time.Second)))
	require.Nil(t, cmd)
	require.False(t, s.Transition().Active())
	require.Contains(t, ansi.Strip(s.View(panes(s))), "content of two")
}

// Half way, the new tab has come in part of the way, its left edge not home yet.
func TestTheTransitionMovesThePanesHalfWay(t *testing.T) {
	s := sliding(t)
	s, _, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	at := time.Unix(1000, 0)
	s, _, _ = s.Update(s.Transition().FrameAt(at))
	s, _, _ = s.Update(s.Transition().FrameAt(at.Add(50 * time.Millisecond)))
	lines := strings.Split(ansi.Strip(s.View(panes(s))), "\n")
	require.True(t, strings.HasPrefix(lines[3], " "), "the pane has not reached its place: %q", lines[3])
	require.Contains(t, lines[3], "┏", "but it is on its way in")
	for _, l := range strings.Split(s.View(panes(s)), "\n") {
		require.Equal(t, 40, ansi.StringWidth(l))
	}
}

func TestNoTransitionWithoutOneOrFromSetTab(t *testing.T) {
	s := sliding(t).SetTab(1)
	require.False(t, s.Transition().Active(), "an app turning the tab itself gets no slide")

	plain := newShell(stubTab{name: "a"}, stubTab{name: "b"})
	plain, _, _ = plain.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	_, _, cmd := plain.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.Nil(t, cmd)
}

// A frame that is not the transition's is left for the app.
func TestOtherFramesAreNotTheShells(t *testing.T) {
	s := sliding(t)
	s, _, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	other := animation.NewFloat(0, 1, time.Second)
	_, handled, _ := s.Update(other.FrameAt(time.Unix(1000, 0)))
	require.False(t, handled)
}

func TestSlideTabsComesFromTheSideTheTabIsOn(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "")
	t.Setenv("TERM", "xterm-256color")
	slide := SlideTabs(100 * time.Millisecond)
	half := func(e effect.Effect) string {
		at := time.Unix(1000, 0)
		e, _ = e.Update(e.FrameAt(at))
		e, _ = e.Update(e.FrameAt(at.Add(50 * time.Millisecond)))
		return e.Apply("abcd")
	}
	require.Equal(t, "  ab", half(slide(0, 1).WithEase(animation.Linear)), "a later tab comes in from the right")
	require.Equal(t, "cd  ", half(slide(1, 0).WithEase(animation.Linear)), "an earlier one from the left")
}
