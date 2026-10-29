// Package text animates a line of text: typed, scrolled, shimmered, rolled
// like an odometer or a departures board, waved, struck and highlighted.
// Non-themable: colours come as arguments, never from the theme.
package text

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/internal/edit"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// Typewriter shows text a grapheme at a time, for a status or a banner that
// arrives rather than appears.
type Typewriter struct {
	text  []string // the graphemes
	shown int
	run   frames.Run
}

// NewTypewriter reveals text at one grapheme per perChar. Plain text only: an
// escape sequence cut half way would colour the rest of the line.
func NewTypewriter(text string, perChar time.Duration) Typewriter {
	g := edit.Graphemes(text)
	w := Typewriter{text: g, run: frames.NewRun(time.Duration(len(g)) * perChar)}
	if !w.run.Active() {
		w.shown = len(g)
	}
	return w
}

// View is the text shown so far.
func (w Typewriter) View() string { return strings.Join(w.text[:w.shown], "") }

// Skip is w with all its text shown, for a key that hurries it.
func (w Typewriter) Skip() Typewriter {
	w.shown, w.run = len(w.text), frames.Run{}
	return w
}

// Active reports whether w is still typing.
func (w Typewriter) Active() bool { return w.run.Active() }

// Tick is the command for w's next frame, nil once it is done.
func (w Typewriter) Tick() tea.Cmd { return w.run.Tick() }

// FrameAt is w's frame for at, for a test or a host that delivers frames itself.
func (w Typewriter) FrameAt(at time.Time) animation.FrameMsg { return w.run.FrameAt(at) }

// Update types on its own frames and asks for the next until it is done.
func (w Typewriter) Update(msg tea.Msg) (Typewriter, tea.Cmd) {
	f, ok := w.run.Own(msg)
	if !ok {
		return w, nil
	}
	if w, _ = w.Advance(f.At); !w.Active() {
		return w, nil
	}
	var cmd tea.Cmd
	w.run, cmd = w.run.Next()
	return w, cmd
}

// Advance is w at now, and whether it is done.
func (w Typewriter) Advance(now time.Time) (Typewriter, bool) {
	if !w.Active() {
		return w, true
	}
	var p float64
	w.run, p = w.run.Advance(now)
	w.shown = min(int(p*float64(len(w.text))), len(w.text))
	return w, !w.Active()
}

// WithDelay holds w empty for d after its first frame.
func (w Typewriter) WithDelay(d time.Duration) Typewriter {
	w.run.Delay = max(d, 0)
	return w
}

// WithRepeat types w n more times after the first, or without end for Forever.
func (w Typewriter) WithRepeat(n int) Typewriter {
	w.run.Repeat = max(n, animation.Forever)
	return w
}

// WithYoyo erases w on every other run, as if typed backwards.
func (w Typewriter) WithYoyo(on bool) Typewriter {
	w.run.Yoyo = on
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
	return Marquee{text: text, width: max(width, 0), every: 120 * time.Millisecond, id: frames.NextID()}
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
	if m.running || animation.Still() || ansi.StringWidth(m.text) <= m.width {
		return m, nil
	}
	if m.id == 0 { // a Marquee never built with NewMarquee
		m.id = frames.NextID()
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
	return frames.Tick(m.every, m.id, m.tag)
}

// FrameAt is m's frame for at, for a test or a host that delivers frames itself.
func (m Marquee) FrameAt(at time.Time) animation.FrameMsg { return frames.NewMsg(at, m.id, m.tag) }

// Update steps on its own ticks and asks for the next while it runs.
func (m Marquee) Update(msg tea.Msg) (Marquee, tea.Cmd) {
	f, ok := frames.Is(msg, m.id, m.tag)
	if !ok || !m.running {
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
	m.offset = int(frames.Paced(now.Sub(m.start))/m.every) % cycle
	return m, false
}

// pad is s filled out with spaces to width cells: a wide grapheme cut at the
// edge leaves one short.
func pad(s string, width int) string {
	return s + strings.Repeat(" ", max(width-ansi.StringWidth(s), 0))
}
