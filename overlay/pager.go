package overlay

import (
	"fmt"
	"strings"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
)

// Pager is a titled block of text taller than a dialog: a command's output, a
// report. The arrows, pgup/pgdown, home and end scroll it; esc, q or enter closes it.
type Pager struct {
	title  string
	lines  []string
	style  lipgloss.Style
	hint   legend.Style
	offset int
	// rows is the one thing copies share: View measures how many rows fit, and
	// Update needs it to stop at the last page. Atomic, as View may run anywhere.
	rows *atomic.Int64
}

// pagerChrome is the rows a Pager spends around its text: border and padding,
// the title and its gap, and the position line and its gap.
const pagerChrome = 2 + 2 + 2 + 2

// NewPager shows text from its first line. CRLF reads as LF and trailing
// blank lines are dropped.
func NewPager(title, text string, style lipgloss.Style, hint legend.Style) Pager {
	lines := strings.Split(strings.TrimRight(strings.ReplaceAll(text, "\r\n", "\n"), "\n"), "\n")
	rows := new(atomic.Int64)
	rows.Store(int64(len(lines)))
	return Pager{title: title, lines: lines, style: style, hint: hint, rows: rows}
}

// Offset is the index of the first line shown.
func (p Pager) Offset() int { return p.offset }

func (p Pager) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}
	page := p.page()
	switch k.String() {
	case "q", "enter":
		return p, Close()
	case "up", "k":
		p = p.scroll(-1)
	case "down", "j":
		p = p.scroll(1)
	case "pgup", "b":
		p = p.scroll(-page)
	case "pgdown", "f", "space":
		p = p.scroll(page)
	case "home", "g":
		p.offset = 0
	case "end", "G":
		p = p.scroll(len(p.lines))
	}
	return p, nil
}

// scroll starts from the offset View would draw, so a stale one past the end
// does not swallow the first keys back.
func (p Pager) scroll(by int) Pager {
	p.offset = min(max(0, min(p.offset, p.lastOffset())+by), p.lastOffset())
	return p
}

// page is the rows the last View showed, at least one.
func (p Pager) page() int { return max(1, int(p.rows.Load())) }

// lastOffset is the offset that shows the last page full.
func (p Pager) lastOffset() int { return max(0, len(p.lines)-p.page()) }

func (p Pager) Legend() []legend.Entry {
	return []legend.Entry{legend.New("↑/↓", "scroll"), legend.New("pgup/pgdn", "page"), legend.New("esc", "close")}
}

func (p Pager) View(body layout.Rect) string {
	rows := max(1, min(len(p.lines), body.H-pagerChrome))
	p.rows.Store(int64(rows))
	offset := min(p.offset, p.lastOffset())
	end := min(len(p.lines), offset+rows)

	out := []string{accent(p.style).Bold(true).Render(p.title), ""}
	out = append(out, p.lines[offset:end]...)
	// The keys ride on the position line: the screen's legend under a modal is
	// still the screen's, so nothing else says how to scroll or leave.
	pos := fmt.Sprintf("lines %d-%d of %d · ↑/↓ pgup/pgdn scroll · esc closes", offset+1, end, len(p.lines))
	return Box(append(out, "", p.hint.Text.Render(pos)), body, p.style.Padding(1, 3))
}
