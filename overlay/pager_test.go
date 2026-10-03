package overlay

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
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
	p := NewPager("Output", numbered(40), lipgloss.NewStyle(), legend.Style{})

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
	p := NewPager("Output", numbered(40), lipgloss.NewStyle(), legend.Style{})
	_ = p.View(body)

	o, _ := p.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	require.Equal(t, 30, o.(Pager).Offset(), "end shows the last page full, not one line")
	require.Contains(t, ansi.Strip(o.View(body)), "line 40")
	o, _ = o.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, 30, o.(Pager).Offset(), "scrolling past the end does nothing")
	o, _ = o.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	require.Equal(t, 29, o.(Pager).Offset(), "the first key back up moves at once")
}

func TestPagerCloses(t *testing.T) {
	p := NewPager("Output", "one\r\ntwo\r\n", lipgloss.NewStyle(), legend.Style{})
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
