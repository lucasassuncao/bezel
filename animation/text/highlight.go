package text

import (
	"image/color"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/edit"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// Highlight runs a marker under text from left to right and leaves it there,
// as a highlighter pen would: the line a search found, the word that changed.
type Highlight struct {
	text   []string // the graphemes
	marker color.Color
	p      float64
	run    frames.Run
}

var _ bezel.Model[Highlight] = Highlight{}

// NewHighlight marks text with marker over dur. Still, it is born marked.
func NewHighlight(text string, marker color.Color, dur time.Duration) Highlight {
	h := Highlight{text: edit.Graphemes(text), marker: marker, run: frames.NewRun(dur)}
	if !h.Active() {
		h.p = 1
	}
	return h
}

// WithEase is h following e instead of EaseInOut.
func (h Highlight) WithEase(e animation.Ease) Highlight {
	h.run.Ease = e
	return h
}

// WithDelay holds the marker off for d after the first frame.
func (h Highlight) WithDelay(d time.Duration) Highlight {
	h.run.Delay = max(d, 0)
	return h
}

// View is the text with the marked graphemes on the marker.
func (h Highlight) View() string {
	return swept(h.text, frames.Or(h.run.Ease)(h.p), lipgloss.NewStyle().Background(h.marker))
}

// Active reports whether the marker is still moving.
func (h Highlight) Active() bool { return h.run.Active() }

// Tick is the command for h's next frame, nil once it ended.
func (h Highlight) Tick() tea.Cmd { return h.run.Tick() }

// FrameAt is h's frame for at, for a test or a host that delivers frames itself.
func (h Highlight) FrameAt(at time.Time) animation.FrameMsg { return h.run.FrameAt(at) }

// Update marks on its own frames and asks for the next until it ends.
func (h Highlight) Update(msg tea.Msg) (Highlight, bezel.Result) {
	r, p, ok, cmd := h.run.Update(msg)
	if ok {
		h.run, h.p = r, p
	}
	return h, bezel.Result{Handled: ok, Cmd: cmd}
}

// Advance is h at now, and whether it ended.
func (h Highlight) Advance(now time.Time) (Highlight, bool) {
	if !h.Active() {
		return h, true
	}
	h.run, h.p = h.run.Advance(now)
	return h, !h.Active()
}
