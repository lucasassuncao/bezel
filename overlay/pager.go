package overlay

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
)

// Pager is a titled block of text taller than a dialog: a command's output, a
// report. The arrows, pgup/pgdown, home and end scroll it; esc, q or enter closes it.
type Pager struct {
	title string
	lines []string
	style lipgloss.Style
	hint  legend.Style
	// view is shared by every copy: View learns how many rows fit, and Update
	// needs that to stop scrolling at the last page instead of past it.
	view *pagerView
}

type pagerView struct{ offset, rows int }

// pagerChrome is the rows a Pager spends around its text: border and padding,
// the title and its gap, and the position line and its gap.
const pagerChrome = 2 + 2 + 2 + 2

func NewPager(title, text string, style lipgloss.Style, hint legend.Style) Pager {
	lines := strings.Split(strings.TrimRight(strings.ReplaceAll(text, "\r\n", "\n"), "\n"), "\n")
	return Pager{title: title, lines: lines, style: style, hint: hint, view: &pagerView{rows: len(lines)}}
}

// Offset is the index of the first line shown.
func (p Pager) Offset() int { return p.view.offset }

func (p Pager) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}
	page := max(1, p.view.rows)
	switch k.String() {
	case "q", "enter":
		return p, Close()
	case "up", "k":
		p.scroll(-1)
	case "down", "j":
		p.scroll(1)
	case "pgup", "b":
		p.scroll(-page)
	case "pgdown", "f", "space":
		p.scroll(page)
	case "home", "g":
		p.view.offset = 0
	case "end", "G":
		p.scroll(len(p.lines))
	}
	return p, nil
}

func (p Pager) scroll(by int) {
	p.view.offset = min(max(0, p.view.offset+by), p.lastOffset())
}

// lastOffset is the offset that shows the last page full.
func (p Pager) lastOffset() int { return max(0, len(p.lines)-max(1, p.view.rows)) }

func (p Pager) Legend() []legend.Entry {
	return []legend.Entry{legend.New("↑/↓", "scroll"), legend.New("pgup/pgdn", "page"), legend.New("esc", "close")}
}

func (p Pager) View(body layout.Rect) string {
	p.view.rows = max(1, min(len(p.lines), body.H-pagerChrome))
	p.view.offset = min(p.view.offset, p.lastOffset())
	end := min(len(p.lines), p.view.offset+p.view.rows)

	out := []string{accent(p.style).Bold(true).Render(p.title), ""}
	out = append(out, p.lines[p.view.offset:end]...)
	// The keys ride on the position line: the screen's legend under a modal is
	// still the screen's, so nothing else says how to scroll or leave.
	pos := fmt.Sprintf("lines %d-%d of %d · ↑/↓ pgup/pgdn scroll · esc closes", p.view.offset+1, end, len(p.lines))
	return Box(append(out, "", p.hint.Text.Render(pos)), body, p.style.Padding(1, 3))
}
