package overlay

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Pager is a titled block of text taller than a dialog: a command's output, a
// report. The arrows, pgup/pgdown, home and end scroll it; esc, q or enter closes it.
type Pager struct {
	look
	title  string
	lines  []string
	offset int
	rows   int // the lines that fit at the last Resize, every line before one
	keys   PagerKeys
}

// PagerKeys are the bindings a Pager scrolls and closes on besides esc. Down
// and PageDown have no help of their own: Up's and PageUp's name both.
type PagerKeys struct {
	Up, Down, PageUp, PageDown, Top, Bottom, Close key.Binding
}

// DefaultPagerKeys are the arrows and j/k, pgup/pgdown with b, f and space,
// home/end with g/G, and q or enter to close.
func DefaultPagerKeys() PagerKeys {
	return PagerKeys{
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/↓", "scroll")),
		Down:     key.NewBinding(key.WithKeys("down", "j")),
		PageUp:   key.NewBinding(key.WithKeys("pgup", "b"), key.WithHelp("pgup/pgdn", "page")),
		PageDown: key.NewBinding(key.WithKeys("pgdown", "f", "space")),
		Top:      key.NewBinding(key.WithKeys("home", "g")),
		Bottom:   key.NewBinding(key.WithKeys("end", "G")),
		Close:    key.NewBinding(key.WithKeys("q", "enter")),
	}
}

// pagerChrome is the rows a Pager spends around its text: border and padding,
// the title and its gap, and the position line and its gap.
const pagerChrome = 2 + 2 + 2 + 2

// NewPager shows text from its first line. CRLF reads as LF and trailing
// blank lines are dropped.
func NewPager(title, text string, th theme.Resolved) Pager {
	lines := strings.Split(strings.TrimRight(strings.ReplaceAll(text, "\r\n", "\n"), "\n"), "\n")
	return Pager{title: title, lines: lines, look: lookOf(th), rows: len(lines), keys: DefaultPagerKeys()}
}

// WithKeys rebinds the pager; its legend and position line follow the bindings' help.
func (p Pager) WithKeys(k PagerKeys) Pager {
	p.keys = k
	return p
}

// Resize makes a page as many lines as fit in body.
func (p Pager) Resize(body layout.Rect) Overlay {
	p.rows = p.fit(body)
	return p
}

// fit is how many lines body leaves for the text, at least one.
func (p Pager) fit(body layout.Rect) int { return max(1, min(len(p.lines), body.H-pagerChrome)) }

// Offset is the index of the first line shown.
func (p Pager) Offset() int { return p.offset }

func (p Pager) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}
	page, km := p.page(), p.keys
	switch {
	case key.Matches(k, km.Close):
		return p, Close()
	case key.Matches(k, km.Up):
		p = p.scroll(-1)
	case key.Matches(k, km.Down):
		p = p.scroll(1)
	case key.Matches(k, km.PageUp):
		p = p.scroll(-page)
	case key.Matches(k, km.PageDown):
		p = p.scroll(page)
	case key.Matches(k, km.Top):
		p.offset = 0
	case key.Matches(k, km.Bottom):
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

// page is the lines a page scrolls, at least one.
func (p Pager) page() int { return max(1, p.rows) }

// lastOffset is the offset that shows the last page full.
func (p Pager) lastOffset() int { return max(0, len(p.lines)-p.page()) }

func (p Pager) Legend() []legend.Entry {
	km := p.keys
	keys := legend.FromBindings([]key.Binding{km.Up, km.Down, km.PageUp, km.PageDown, km.Top, km.Bottom, km.Close})
	return append(keys, legend.New("esc", "close"))
}

func (p Pager) View(body layout.Rect) string {
	rows := p.fit(body)
	offset := min(p.offset, max(0, len(p.lines)-rows))
	end := min(len(p.lines), offset+rows)

	out := []string{accent(p.style()).Bold(true).Render(p.title), ""}
	out = append(out, p.lines[offset:end]...)
	// The keys ride on the position line: the screen's legend under a modal is
	// still the screen's, so nothing else says how to scroll or leave.
	scroll := strings.TrimSpace(p.keys.Up.Help().Key + " " + p.keys.PageUp.Help().Key)
	pos := fmt.Sprintf("lines %d-%d of %d · %s scroll · esc closes", offset+1, end, len(p.lines), scroll)
	return Box(append(out, "", p.hint().Text.Render(pos)), body, dialogFrame(p.style()))
}
