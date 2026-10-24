package animation

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/frames"
)

// Heat moves a grid of values to new ones over dur, cell by cell: a heatmap
// whose cells warm and cool instead of jumping. Hand Values to a chart.
type Heat struct {
	from, to [][]float64
	dur      time.Duration
	p        float64
	run      frames.Run
}

// NewHeat is values at rest; each Set moves them over dur.
func NewHeat(values [][]float64, dur time.Duration) Heat {
	return Heat{from: clone2(values), to: clone2(values), dur: dur, p: 1}
}

// Set moves every cell from where it is now to values. A cell the grid did
// not have before rises from zero.
func (h Heat) Set(values [][]float64) (Heat, tea.Cmd) {
	now := h.Values()
	h.to = clone2(values)
	h.from = make([][]float64, len(values))
	for y, row := range values {
		h.from[y] = make([]float64, len(row))
		for x := range row {
			if y < len(now) && x < len(now[y]) {
				h.from[y][x] = now[y][x]
			}
		}
	}
	h.run, h.p = frames.NewRun(h.dur), 0
	if !h.Active() {
		h.p = 1
	}
	return h, h.Tick()
}

// Values are the numbers to draw now.
func (h Heat) Values() [][]float64 {
	q := frames.Or(h.run.Ease)(h.p)
	if !h.Active() {
		q = 1
	}
	out := make([][]float64, len(h.to))
	for y, row := range h.to {
		out[y] = make([]float64, len(row))
		for x, v := range row {
			out[y][x] = h.from[y][x] + (v-h.from[y][x])*q
		}
	}
	return out
}

func clone2(g [][]float64) [][]float64 {
	out := make([][]float64, len(g))
	for i, row := range g {
		out[i] = append([]float64(nil), row...)
	}
	return out
}

// Active reports whether any cell is still changing.
func (h Heat) Active() bool { return h.run.Active() }

// Tick is the command for h's next frame, nil at rest.
func (h Heat) Tick() tea.Cmd { return h.run.Tick() }

// FrameAt is h's frame for at, for a test or a host that delivers frames itself.
func (h Heat) FrameAt(at time.Time) FrameMsg { return h.run.FrameAt(at) }

// Update moves on its own frames and asks for the next until it settles.
func (h Heat) Update(msg tea.Msg) (Heat, tea.Cmd) {
	r, p, ok, cmd := h.run.Update(msg)
	if ok {
		h.run, h.p = r, p
	}
	return h, cmd
}

// Advance is h at now, and whether it settled.
func (h Heat) Advance(now time.Time) (Heat, bool) {
	if !h.Active() {
		return h, true
	}
	h.run, h.p = h.run.Advance(now)
	return h, !h.Active()
}
