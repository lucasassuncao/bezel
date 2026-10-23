package animation

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
)

// rest is how close to its target, and how slow, a spring must be to stop:
// a hundredth of a cell is below anything a terminal can draw.
const rest = 0.005

// Spring follows a target with speed and inertia. Unlike a Tween it has no
// duration, so a new target mid-flight bends the motion instead of starting
// it over: a progress bar fed often, a cursor sliding to the row picked.
type Spring struct {
	Pos, Vel, Target float64
	// Freq is how stiff it is, in radians a second: 12 when zero or less.
	Freq float64
	// Damping 1 arrives without overshoot; below 1 it bounces. 1 when zero or less.
	Damping float64
	last    time.Time // the previous frame, zero until the first
	moving  bool
	id, tag int64
}

// NewSpring is a spring at rest at pos, stiffness 12 and damping 1.
func NewSpring(pos float64) Spring {
	return Spring{Pos: pos, Target: pos, Freq: 12, Damping: 1, id: lastID.Add(1)}
}

// WithFreq is s with stiffness f, in radians a second.
func (s Spring) WithFreq(f float64) Spring {
	s.Freq = f
	return s
}

// WithDamping is s with damping d: 1 settles, below it bounces, above it creeps.
func (s Spring) WithDamping(d float64) Spring {
	s.Damping = d
	return s
}

// To aims s at target. A spring at rest starts its frames; a moving one keeps
// the ones it has, and its speed. When motion is Still it jumps there.
func (s Spring) To(target float64) (Spring, tea.Cmd) {
	s.Target = target
	if Still() {
		s.Pos, s.Vel, s.moving = target, 0, false
		return s, nil
	}
	if s.moving || (s.Pos == target && s.Vel == 0) {
		return s, nil
	}
	if s.id == 0 { // a Spring never built with NewSpring
		s.id = lastID.Add(1)
	}
	s.moving, s.last = true, time.Time{}
	s.tag++
	return s, s.Tick()
}

// Active reports whether s is still moving.
func (s Spring) Active() bool { return s.moving }

// Int is Pos rounded to the nearest whole number: the cell it is drawn on.
func (s Spring) Int() int { return int(math.Round(s.Pos)) }

// Tick is the command for s's next frame, nil at rest.
func (s Spring) Tick() tea.Cmd {
	if !s.moving {
		return nil
	}
	id, tag := s.id, s.tag
	return tea.Tick(Frame, func(at time.Time) tea.Msg { return FrameMsg{At: at, id: id, tag: tag} })
}

// FrameAt is s's frame for at, for a test or a host that delivers frames itself.
func (s Spring) FrameAt(at time.Time) FrameMsg { return FrameMsg{At: at, id: s.id, tag: s.tag} }

// Update moves s on its own frames and asks for the next until it rests.
func (s Spring) Update(msg tea.Msg) (Spring, tea.Cmd) {
	f, ok := msg.(FrameMsg)
	if !ok || !s.moving || f.id != s.id || f.tag != s.tag {
		return s, nil
	}
	if s, _ = s.Advance(f.At); !s.moving {
		return s, nil
	}
	s.tag++
	return s, s.Tick()
}

// Advance is s moved to now, and whether it came to rest. The first frame
// only starts the clock; each later one solves the motion exactly over the
// time since the last, so a slow frame never throws it off.
func (s Spring) Advance(now time.Time) (Spring, bool) {
	if !s.moving {
		return s, true
	}
	if s.last.IsZero() {
		s.last = now
		return s, false
	}
	dt := now.Sub(s.last).Seconds()
	s.last = now
	x, v := step(s.Pos-s.Target, s.Vel, s.freq(), s.damping(), max(dt, 0))
	s.Pos, s.Vel = s.Target+x, v
	if math.Abs(x) < rest && math.Abs(v) < rest*10 {
		s.Pos, s.Vel, s.moving = s.Target, 0, false
	}
	return s, !s.moving
}

func (s Spring) freq() float64 {
	if s.Freq <= 0 {
		return 12
	}
	return s.Freq
}

func (s Spring) damping() float64 {
	if s.Damping <= 0 {
		return 1
	}
	return s.Damping
}

// step is the damped oscillator's offset x and speed v after dt seconds, in
// closed form for each of the three regimes.
func step(x, v, w, z, dt float64) (float64, float64) {
	switch {
	case math.Abs(z-1) < 1e-9: // critical
		a, b := x, v+w*x
		e := math.Exp(-w * dt)
		return (a + b*dt) * e, (b - w*(a+b*dt)) * e
	case z < 1: // under: it overshoots and rings
		wd := w * math.Sqrt(1-z*z)
		a, b := x, (v+z*w*x)/wd
		e, c, sn := math.Exp(-z*w*dt), math.Cos(wd*dt), math.Sin(wd*dt)
		return e * (a*c + b*sn), e * ((wd*b-z*w*a)*c - (z*w*b+wd*a)*sn)
	default: // over: it creeps in
		r := w * math.Sqrt(z*z-1)
		r1, r2 := -z*w+r, -z*w-r
		c1 := (v - r2*x) / (r1 - r2)
		c2 := x - c1
		e1, e2 := math.Exp(r1*dt), math.Exp(r2*dt)
		return c1*e1 + c2*e2, c1*r1*e1 + c2*r2*e2
	}
}
