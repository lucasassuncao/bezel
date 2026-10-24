// Package toast shows short notices in a corner of the screen, stacked,
// sliding in and out on their own: saved, copied, failed to connect.
package toast

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/theme"
)

// Kind is what a toast reports; it picks the colour of its edge.
type Kind int

const (
	Info Kind = iota
	Success
	Warning
	Error
)

// slide is how long a toast takes to come in, and again to go out.
const slide = 250 * time.Millisecond

// item is one toast: what it says, and the frame it arrived on.
type item struct {
	text string
	kind Kind
	born time.Time // zero until the first frame after Push
	key  int64
}

// expireMsg ends the toast key when motion is Still and no frames run.
type expireMsg struct{ id, key int64 }

// Model is the stack of toasts, newest at the top. It is a value: every
// method returns the new stack.
type Model struct {
	items   []item
	ttl     time.Duration
	width   int
	last    int64     // the key the newest toast took
	now     time.Time // the last frame's time
	running bool
	id, tag int64
}

// New is an empty stack whose toasts stay three seconds, 36 columns wide.
func New() Model { return Model{ttl: 3 * time.Second, width: 36, id: frames.NextID()} }

// WithTTL keeps each toast on screen for d, sliding included.
func (m Model) WithTTL(d time.Duration) Model {
	if d > 0 {
		m.ttl = d
	}
	return m
}

// WithWidth draws each toast w columns wide, edges included.
func (m Model) WithWidth(w int) Model {
	m.width = max(w, 8)
	return m
}

// Len is how many toasts are showing.
func (m Model) Len() int { return len(m.items) }

// Push shows text as a toast of kind on top of the stack.
func (m Model) Push(text string, kind Kind) (Model, tea.Cmd) {
	if m.id == 0 { // a Model never built with New
		m = New().WithWidth(max(m.width, 8)).WithTTL(m.ttl)
	}
	m.last++
	m.items = append([]item{{text: text, kind: kind, key: m.last}}, m.items...)
	if frames.Still() {
		id, key := m.id, m.last
		return m, tea.Tick(m.ttl, func(time.Time) tea.Msg { return expireMsg{id: id, key: key} })
	}
	if m.running {
		return m, nil
	}
	m.running = true
	m.tag++
	return m, frames.Tick(frames.Frame, m.id, m.tag)
}

// Update slides the toasts on the stack's own frames and drops the ones whose
// time is up, asking for frames while any is showing.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if e, ok := msg.(expireMsg); ok && e.id == m.id {
		m.items = drop(m.items, func(it item) bool { return it.key == e.key })
		return m, nil
	}
	f, ok := frames.Is(msg, m.id, m.tag)
	if !ok || !m.running {
		return m, nil
	}
	m = m.Advance(f.At)
	if len(m.items) == 0 {
		m.running = false
		return m, nil
	}
	m.tag++
	return m, frames.Tick(frames.Frame, m.id, m.tag)
}

// Advance is the stack at now: newcomers stamped, the expired gone.
func (m Model) Advance(now time.Time) Model {
	items := make([]item, len(m.items))
	copy(items, m.items)
	for i := range items {
		if items[i].born.IsZero() {
			items[i].born = now
		}
	}
	m.items = drop(items, func(it item) bool { return frames.Paced(now.Sub(it.born)) >= m.ttl })
	m.now = now
	return m
}

// FrameAt is the stack's frame for at, for a test that delivers frames itself.
func (m Model) FrameAt(at time.Time) tea.Msg { return frames.NewMsg(at, m.id, m.tag) }

// drop is items without the ones gone says.
func drop(items []item, gone func(item) bool) []item {
	out := make([]item, 0, len(items))
	for _, it := range items {
		if !gone(it) {
			out = append(out, it)
		}
	}
	return out
}

// offset is how many columns right of its place a toast is drawn now: the
// whole width as it arrives and as it leaves, nothing between.
func (m Model) offset(it item) int {
	if it.born.IsZero() {
		return m.width + 1
	}
	if frames.Still() || m.now.IsZero() {
		return 0
	}
	age := frames.Paced(m.now.Sub(it.born))
	share := 0.0
	switch {
	case age < slide:
		share = 1 - float64(age)/float64(slide)
	case age > m.ttl-slide:
		share = float64(age-(m.ttl-slide)) / float64(slide)
	}
	share = min(max(share, 0), 1)
	return int(share * share * float64(m.width+1))
}

// box is one toast drawn width columns wide, its edge in its kind's colour.
func (m Model) box(it item, th theme.Resolved) []string {
	edge := map[Kind]lipgloss.Style{Info: th.Info, Success: th.Success, Warning: th.Warning, Error: th.Danger}[it.kind]
	inner := m.width - 4
	lines := draw.Wrap(it.text, inner)
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(edge.GetForeground()).Padding(0, 1).Width(m.width)
	return strings.Split(style.Render(strings.Join(lines, "\n")), "\n")
}

// View is the stack, newest first, each toast slid by how far it is from
// its place; columns it has slid out of are blank.
func (m Model) View(th theme.Resolved) string {
	var out []string
	for _, it := range m.items {
		off := m.offset(it)
		for _, l := range m.box(it, th) {
			out = append(out, strings.Repeat(" ", min(off, m.width))+ansi.Truncate(l, max(m.width-off, 0), ""))
		}
	}
	return strings.Join(out, "\n")
}

// Over draws the stack on screen's top right corner, one row down and one
// column in, the screen showing wherever a toast is not.
func (m Model) Over(screen string, th theme.Resolved) string {
	w, _ := draw.BlockSize(screen)
	y := 1
	for _, it := range m.items {
		off := m.offset(it)
		x := w - 1 - m.width + off
		for _, l := range m.box(it, th) {
			if visible := m.width - off; visible > 0 && x < w {
				screen = draw.Composite(ansi.Truncate(l, visible, ""), screen, x, y)
			}
			y++
		}
	}
	return screen
}
