package animation

import (
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/frames"
)

// Series animates a run of chart values: grown from nothing all at once,
// drawn in from the left, or fed live with each new value rising at the
// right. Hand Values to a chart every frame.
type Series struct {
	from, to []float64
	draw     bool // reveal left to right instead of growing every value at once
	dur      time.Duration
	p        float64
	run      frames.Run
}

// NewGrow raises every one of values from zero over dur: bars that grow in.
func NewGrow(values []float64, dur time.Duration) Series {
	s := Series{from: make([]float64, len(values)), to: slices.Clone(values), dur: dur, run: frames.NewRun(dur)}
	if !s.Active() {
		s.p = 1
	}
	return s
}

// NewDraw reveals values from the left over dur, each rising as the front
// reaches it: a line drawn across a chart. Values not reached yet are zero.
func NewDraw(values []float64, dur time.Duration) Series {
	s := NewGrow(values, dur)
	s.draw = true
	return s
}

// WithEase is s following e instead of EaseInOut.
func (s Series) WithEase(e Ease) Series {
	s.run.Ease = e
	return s
}

// WithDuration is how long each later Push takes to rise.
func (s Series) WithDuration(d time.Duration) Series {
	s.dur = d
	return s
}

// Push adds v at the right, dropping the oldest past keep values; the others
// hold and v rises from zero over the series' duration: a live sparkline.
func (s Series) Push(v float64, keep int) (Series, tea.Cmd) {
	next := append(s.Values(), v)
	if keep > 0 && len(next) > keep {
		next = next[len(next)-keep:]
	}
	s.to = next
	s.from = slices.Clone(next)
	s.from[len(s.from)-1] = 0
	s.draw, s.p = false, 0
	ease := s.run.Ease
	s.run = frames.NewRun(s.dur)
	s.run.Ease = ease
	if !s.Active() {
		s.p = 1
	}
	return s, s.Tick()
}

// Values are the numbers to draw now.
func (s Series) Values() []float64 {
	q := frames.Or(s.run.Ease)(s.p)
	if !s.Active() {
		q = 1
	}
	out := make([]float64, len(s.to))
	for i, v := range s.to {
		k := q
		if s.draw {
			k = min(max(q*float64(len(s.to))-float64(i), 0), 1) // the front reaches i at i/len
		}
		out[i] = s.from[i] + (v-s.from[i])*k
	}
	return out
}

// Active reports whether any value is still moving.
func (s Series) Active() bool { return s.run.Active() }

// Tick is the command for s's next frame, nil at rest.
func (s Series) Tick() tea.Cmd { return s.run.Tick() }

// FrameAt is s's frame for at, for a test or a host that delivers frames itself.
func (s Series) FrameAt(at time.Time) FrameMsg { return s.run.FrameAt(at) }

// Update moves on its own frames and asks for the next until it settles.
func (s Series) Update(msg tea.Msg) (Series, tea.Cmd) {
	r, p, ok, cmd := s.run.Update(msg)
	if ok {
		s.run, s.p = r, p
	}
	return s, cmd
}

// Advance is s at now, and whether it settled.
func (s Series) Advance(now time.Time) (Series, bool) {
	if !s.Active() {
		return s, true
	}
	s.run, s.p = s.run.Advance(now)
	return s, !s.Active()
}
