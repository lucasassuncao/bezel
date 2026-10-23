package animation

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/internal/edit"
)

// Typewriter shows text a grapheme at a time, for a status or a banner that
// arrives rather than appears.
type Typewriter struct {
	text  []string // the graphemes
	shown int
	run   run
}

// NewTypewriter reveals text at one grapheme per perChar. Plain text only: an
// escape sequence cut half way would colour the rest of the line.
func NewTypewriter(text string, perChar time.Duration) Typewriter {
	g := edit.Graphemes(text)
	w := Typewriter{text: g, run: newRun(time.Duration(len(g)) * perChar)}
	w.run.ease = Linear
	if !w.run.active() {
		w.shown = len(g)
	}
	return w
}

// View is the text shown so far.
func (w Typewriter) View() string { return strings.Join(w.text[:w.shown], "") }

// Skip is w with all its text shown, for a key that hurries it.
func (w Typewriter) Skip() Typewriter {
	w.shown, w.run = len(w.text), run{}
	return w
}

// Active reports whether w is still typing.
func (w Typewriter) Active() bool { return w.run.active() }

// Tick is the command for w's next frame, nil once it is done.
func (w Typewriter) Tick() tea.Cmd { return w.run.tick() }

// FrameAt is w's frame for at, for a test or a host that delivers frames itself.
func (w Typewriter) FrameAt(at time.Time) FrameMsg { return w.run.frameAt(at) }

// Update types on its own frames and asks for the next until it is done.
func (w Typewriter) Update(msg tea.Msg) (Typewriter, tea.Cmd) {
	f, ok := w.run.own(msg)
	if !ok {
		return w, nil
	}
	if w, _ = w.Advance(f.At); !w.Active() {
		return w, nil
	}
	var cmd tea.Cmd
	w.run, cmd = w.run.next()
	return w, cmd
}

// Advance is w at now, and whether it is done.
func (w Typewriter) Advance(now time.Time) (Typewriter, bool) {
	if !w.Active() {
		return w, true
	}
	var p float64
	w.run, p = w.run.advance(now)
	w.shown = min(int(p*float64(len(w.text))), len(w.text))
	return w, !w.Active()
}

// WithDelay holds w empty for d after its first frame.
func (w Typewriter) WithDelay(d time.Duration) Typewriter {
	w.run.delay = max(d, 0)
	return w
}

// WithRepeat types w n more times after the first, or without end for Forever.
func (w Typewriter) WithRepeat(n int) Typewriter {
	w.run.repeat = max(n, Forever)
	return w
}

// WithYoyo erases w on every other run, as if typed backwards.
func (w Typewriter) WithYoyo() Typewriter {
	w.run.yoyo = true
	return w
}

// marqueeGap is what passes between the end of the text and its start again.
const marqueeGap = "   "

// Marquee scrolls text that is wider than its room, a cell at a time, round
// and round until stopped: a long name in a list row under the cursor.
type Marquee struct {
	text    string
	width   int
	every   time.Duration
	start   time.Time
	offset  int
	running bool
	id, tag int64
}

// NewMarquee is text in width cells, still until Start, stepping every 120ms.
func NewMarquee(text string, width int) Marquee {
	return Marquee{text: text, width: max(width, 0), every: 120 * time.Millisecond, id: lastID.Add(1)}
}

// WithEvery steps the text once every d instead of every 120ms.
func (m Marquee) WithEvery(d time.Duration) Marquee {
	if d > 0 {
		m.every = d
	}
	return m
}

// Start sets the text moving; text that fits, or motion that is Still, is left
// alone and asks nothing.
func (m Marquee) Start() (Marquee, tea.Cmd) {
	if m.running || Still() || ansi.StringWidth(m.text) <= m.width {
		return m, nil
	}
	if m.id == 0 { // a Marquee never built with NewMarquee
		m.id = lastID.Add(1)
	}
	m.running, m.start = true, time.Time{}
	m.tag++
	return m, m.Tick()
}

// Stop halts the text back at its start; ticks already sent are dropped.
func (m Marquee) Stop() Marquee {
	m.running, m.offset = false, 0
	m.tag++
	return m
}

// Active reports whether the text is moving.
func (m Marquee) Active() bool { return m.running }

// View is exactly width cells of the text, from where it has scrolled to.
func (m Marquee) View() string {
	if !m.running {
		return pad(ansi.Truncate(m.text, m.width, ""), m.width)
	}
	loop := m.text + marqueeGap + m.text
	return pad(ansi.Cut(loop, m.offset, m.offset+m.width), m.width)
}

// Tick is the command for m's next step, nil while it is still.
func (m Marquee) Tick() tea.Cmd {
	if !m.running {
		return nil
	}
	id, tag := m.id, m.tag
	return tea.Tick(m.every, func(at time.Time) tea.Msg { return FrameMsg{At: at, id: id, tag: tag} })
}

// FrameAt is m's frame for at, for a test or a host that delivers frames itself.
func (m Marquee) FrameAt(at time.Time) FrameMsg { return FrameMsg{At: at, id: m.id, tag: m.tag} }

// Update steps on its own ticks and asks for the next while it runs.
func (m Marquee) Update(msg tea.Msg) (Marquee, tea.Cmd) {
	f, ok := msg.(FrameMsg)
	if !ok || !m.running || f.id != m.id || f.tag != m.tag {
		return m, nil
	}
	m, _ = m.Advance(f.At)
	m.tag++
	return m, m.Tick()
}

// Advance is m at now; it never finishes, so done is only true when stopped.
func (m Marquee) Advance(now time.Time) (Marquee, bool) {
	if !m.running {
		return m, true
	}
	if m.start.IsZero() {
		m.start = now
	}
	cycle := ansi.StringWidth(m.text) + len(marqueeGap)
	m.offset = int(now.Sub(m.start)/m.every) % cycle
	return m, false
}

// pad is s filled out with spaces to width cells: a wide grapheme cut at the
// edge leaves one short.
func pad(s string, width int) string {
	return s + strings.Repeat(" ", max(width-ansi.StringWidth(s), 0))
}
