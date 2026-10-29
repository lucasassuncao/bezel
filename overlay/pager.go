package overlay

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/component/viewport"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Pager is a titled block of text taller than a dialog: a command's output, a
// report. The arrows, pgup/pgdown, home and end scroll it, ← and → pan a long
// line, and the wheel scrolls; esc, q or enter closes it.
type Pager struct {
	look
	title  string
	vp     viewport.Model
	widest int // the widest line's cells, so a short text keeps a narrow box
	keys   PagerKeys
}

// PagerKeys are the viewport's scrolling keys, and Close besides esc.
type PagerKeys struct {
	viewport.Keys
	Close key.Binding
}

// DefaultPagerKeys are the viewport's own, and q or enter to close.
func DefaultPagerKeys() PagerKeys {
	return PagerKeys{Keys: viewport.DefaultKeys(), Close: key.New("q", "enter")}
}

// VimPagerKeys scroll with viewport.VimKeys and close as DefaultPagerKeys do.
func VimPagerKeys() PagerKeys {
	k := DefaultPagerKeys()
	k.Keys = viewport.VimKeys()
	return k
}

// pagerChrome is the rows a Pager spends around its text: border and padding,
// the title and its gap, and the position line and its gap.
const pagerChrome = 2 + 2 + 2 + 2

// NewPager shows text from its first line. CRLF reads as LF and trailing
// blank lines are dropped.
func NewPager(title, text string, th theme.Resolved) Pager {
	keys := DefaultPagerKeys()
	vp := viewport.New().WithKeys(keys.Keys).SetContent(strings.TrimRight(strings.ReplaceAll(text, "\r\n", "\n"), "\n"))
	widest, _ := draw.BlockSize(strings.Join(vp.Lines(), "\n"))
	return Pager{title: title, vp: vp, widest: widest, look: lookOf(th), keys: keys}
}

// WithKeys rebinds the pager; its legend and position line follow the bindings' help.
func (p Pager) WithKeys(k PagerKeys) Pager {
	p.keys = k
	p.vp = p.vp.WithKeys(k.Keys)
	return p
}

// Resize sizes the viewport to what fits in body.
func (p Pager) Resize(body layout.Rect) Overlay {
	p.vp = p.sized(body)
	return p
}

// sized is the viewport as wide as the text, within the box, and as many
// lines tall as fit, at least one.
func (p Pager) sized(body layout.Rect) viewport.Model {
	rows := max(1, min(p.vp.TotalRows(), body.H-pagerChrome))
	return p.vp.SetWidth(min(Inner(body), max(1, p.widest))).SetHeight(rows)
}

// Offset is the index of the first line shown.
func (p Pager) Offset() int { return p.vp.YOffset() }

func (p Pager) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok && key.Matches(k, p.keys.Close) {
		return p, Close()
	}
	p.vp, _ = p.vp.Update(msg)
	return p, nil
}

func (p Pager) Legend() []legend.Entry {
	keys := append(p.vp.Legend(), legend.FromBindings([]key.Binding{p.keys.Close})...)
	return append(keys, legend.New("esc", "close"))
}

func (p Pager) View(body layout.Rect) string {
	vp := p.sized(body)
	s := vp.Scroll()
	out := strings.Split(vp.View(*p.styles()), "\n")
	// The keys ride on the position line: the screen's legend under a modal is
	// still the screen's, so nothing else says how to scroll or leave.
	scroll := strings.TrimSpace(p.keys.Up.Help().Key + " " + p.keys.PageUp.Help().Key)
	pos := fmt.Sprintf("lines %d-%d of %d · %s scroll · esc closes", s.Offset+1, s.Offset+s.Visible, s.Total, scroll)
	return dialog(p.title, append(out, "", p.hint().Text.Render(pos)), "", p.style(), body)
}
