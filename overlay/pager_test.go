package overlay

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

func numbered(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	return b.String()
}

func TestPagerShowsOnePageAndScrolls(t *testing.T) {
	body := layout.Rect{W: 60, H: 18} // 10 rows of text once the chrome is paid
	p := NewPager("Output", numbered(40), theme.Resolved{}).Resize(body)

	view := ansi.Strip(p.View(body))
	require.Contains(t, view, "line 1")
	require.NotContains(t, view, "line 11")
	require.Contains(t, view, "lines 1-10 of 40")
	require.Contains(t, view, "esc closes", "the pager says how to leave it")

	o, _ := p.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, 1, o.(Pager).Offset())
	o, _ = o.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	require.Equal(t, 11, o.(Pager).Offset())
}

func TestPagerStopsAtTheLastPage(t *testing.T) {
	body := layout.Rect{W: 60, H: 18}
	p := NewPager("Output", numbered(40), theme.Resolved{}).Resize(body)

	o, _ := p.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	require.Equal(t, 30, o.(Pager).Offset(), "end shows the last page full, not one line")
	require.Contains(t, ansi.Strip(o.View(body)), "line 40")
	o, _ = o.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, 30, o.(Pager).Offset(), "scrolling past the end does nothing")
	o, _ = o.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	require.Equal(t, 29, o.(Pager).Offset(), "the first key back up moves at once")
}

func TestPagerCloses(t *testing.T) {
	p := NewPager("Output", "one\r\ntwo\r\n", theme.Resolved{})
	require.NotContains(t, p.View(layout.Rect{W: 60, H: 18}), "\r")

	for _, k := range []string{"q", "enter"} {
		key := press(k)
		if k == "enter" {
			key = tea.KeyPressMsg{Code: tea.KeyEnter}
		}
		_, cmd := p.Update(key)
		require.NotNil(t, cmd, "%s should close", k)
		require.Equal(t, CloseMsg{}, cmd())
	}
}

// Scrolling is a copy's own: an earlier Pager keeps its place.
func TestPagerCopiesScrollApart(t *testing.T) {
	p := NewPager("Output", numbered(40), theme.Resolved{}).Resize(layout.Rect{W: 60, H: 18}).(Pager)
	o, _ := p.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	require.Equal(t, 1, o.(Pager).Offset())
	require.Equal(t, 0, p.Offset(), "the earlier copy scrolled too")
}

// A dialog keeps the lines it was given, whatever the caller does after.
func TestDialogsKeepTheirLines(t *testing.T) {
	lines := []string{"original"}
	text := NewText("T", lines, theme.Resolved{})
	pick := NewPick("P", lines, nil, theme.Resolved{}).WithNotes(lines...)
	choice := NewChoice("C", lines, []Option{Opt("y", "yes", nil)}, theme.Resolved{})
	lines[0] = "changed"

	body := layout.Rect{W: 60, H: 18}
	for _, o := range []Overlay{text, pick, choice} {
		require.Contains(t, o.View(body), "original")
	}
}

func TestPagerScrollsOnTheKeysItWasGiven(t *testing.T) {
	keys := DefaultPagerKeys()
	keys.Down = key.NewBinding(key.WithKeys("n"))
	p := NewPager("t", "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12", theme.Resolved{}).WithKeys(keys)
	o := p.Resize(layout.Rect{W: 60, H: 12})

	o, _ = o.Update(press("j"))
	require.Equal(t, 0, o.(Pager).Offset(), "j is no longer bound")
	o, _ = o.Update(press("n"))
	require.Equal(t, 1, o.(Pager).Offset())
}

func TestPagerPansALongLine(t *testing.T) {
	body := layout.Rect{W: 60, H: 18}
	o := NewPager("t", strings.Repeat("x", 100)+"END", theme.Resolved{}).Resize(body)
	require.NotContains(t, ansi.Strip(o.View(body)), "END")
	for range 20 {
		o, _ = o.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	}
	require.Contains(t, ansi.Strip(o.View(body)), "END")
}

func TestPagerScrollsOnTheWheel(t *testing.T) {
	body := layout.Rect{W: 60, H: 18}
	o, _ := NewPager("Output", numbered(40), theme.Resolved{}).Resize(body).Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	require.Equal(t, 3, o.(Pager).Offset())
}
