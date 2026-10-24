package text

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/edit"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// catchUp is how many graphemes behind a Stream may fall before it speeds
// up: twice as fast at twice this backlog, and so on.
const catchUp = 120

// Stream shows text that arrives in pieces at a steady pace, however the
// pieces come: an answer streamed from a model, a log tailed in bursts. It
// speeds up when it falls far behind, so what arrived is never long unseen.
type Stream struct {
	text    []string // every grapheme pushed so far
	shown   int
	perChar time.Duration
	credit  float64   // graphemes earned but not yet shown
	last    time.Time // the frame the credit was counted to
	running bool
	id, tag int64
}

// NewStream reveals what it is pushed at one grapheme per perChar, 10ms when
// perChar is not positive.
func NewStream(perChar time.Duration) Stream {
	if perChar <= 0 {
		perChar = 10 * time.Millisecond
	}
	return Stream{perChar: perChar, id: frames.NextID()}
}

// Push adds text to the end and starts revealing it if the stream had caught
// up. With motion Still it shows at once.
func (s Stream) Push(text string) (Stream, tea.Cmd) {
	s.text = append(s.text[:len(s.text):len(s.text)], edit.Graphemes(text)...)
	if animation.Still() {
		s.shown = len(s.text)
		return s, nil
	}
	if s.running || s.shown == len(s.text) {
		return s, nil
	}
	if s.id == 0 { // a Stream never built with NewStream
		s.id = frames.NextID()
	}
	s.running, s.last = true, time.Time{}
	s.tag++
	return s, s.Tick()
}

// Skip shows everything pushed so far at once.
func (s Stream) Skip() Stream {
	s.shown, s.running, s.credit = len(s.text), false, 0
	s.tag++
	return s
}

// View is the text revealed so far.
func (s Stream) View() string { return strings.Join(s.text[:s.shown], "") }

// Behind is how many graphemes have arrived but are not shown yet.
func (s Stream) Behind() int { return len(s.text) - s.shown }

// Active reports whether s is still revealing.
func (s Stream) Active() bool { return s.running }

// Tick is the command for s's next frame, nil once it caught up.
func (s Stream) Tick() tea.Cmd {
	if !s.running {
		return nil
	}
	return frames.Tick(animation.Frame, s.id, s.tag)
}

// FrameAt is s's frame for at, for a test or a host that delivers frames itself.
func (s Stream) FrameAt(at time.Time) animation.FrameMsg { return frames.NewMsg(at, s.id, s.tag) }

// Update reveals on its own frames and asks for the next until it caught up.
func (s Stream) Update(msg tea.Msg) (Stream, tea.Cmd) {
	f, ok := frames.Is(msg, s.id, s.tag)
	if !ok || !s.running {
		return s, nil
	}
	if s, _ = s.Advance(f.At); !s.running {
		return s, nil
	}
	s.tag++
	return s, s.Tick()
}

// Advance is s at now, and whether it caught up.
func (s Stream) Advance(now time.Time) (Stream, bool) {
	if !s.running {
		return s, true
	}
	if !s.last.IsZero() {
		speed := 1 + float64(s.Behind())/catchUp
		s.credit += speed * float64(frames.Paced(now.Sub(s.last))) / float64(s.perChar)
	}
	s.last = now
	n := min(int(s.credit), s.Behind())
	s.shown += n
	s.credit -= float64(n)
	if s.shown == len(s.text) {
		s.running, s.credit = false, 0
	}
	return s, !s.running
}
