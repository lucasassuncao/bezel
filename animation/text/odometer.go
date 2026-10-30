package text

import (
	"fmt"
	"strconv"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// Odometer rolls a whole number to a new value digit by digit, each column
// turning through the digits between, the right-hand ones spinning most: a
// counter that changes often and should be seen to change.
type Odometer struct {
	from, to int
	width    int // digits, the wider of the two values, zero-padded
	dur      time.Duration
	p        float64
	run      frames.Run
}

var _ bezel.Model[Odometer] = Odometer{}

// NewOdometer rolls from from to to over dur, easing out. Values below zero
// read as zero.
func NewOdometer(from, to int, dur time.Duration) Odometer {
	from, to = max(from, 0), max(to, 0)
	o := Odometer{from: from, to: to, dur: dur, width: max(len(strconv.Itoa(from)), len(strconv.Itoa(to)))}
	if from != to {
		o.run = frames.NewRun(dur)
		o.run.Ease = animation.EaseOut
	}
	return o
}

// To rolls on from the value o is heading for to n, over the same duration.
func (o Odometer) To(n int) (Odometer, tea.Cmd) {
	next := NewOdometer(o.to, n, o.dur)
	return next, next.Tick()
}

// Value is the number o is rolling to, or resting on.
func (o Odometer) Value() int { return o.to }

// View is the number as the columns show it now, zero-padded to the width.
func (o Odometer) View() string {
	if !o.Active() {
		return fmt.Sprintf("%0*d", o.width, o.to)
	}
	ease := frames.Or(o.run.Ease)
	out := make([]byte, o.width)
	for j := range out {
		k := o.width - 1 - j // the column's power of ten
		a, b := digit(o.from, k), digit(o.to, k)
		steps := (b-a+10)%10 + 10*j // the right-hand columns spin more full turns
		out[j] = "0123456789"[wrap10(a+int(float64(steps)*ease(o.p)))]
	}
	return string(out)
}

// digit is n's digit for 10^k.
func digit(n, k int) int {
	for range k {
		n /= 10
	}
	return n % 10
}

// Active reports whether the columns are still turning.
func (o Odometer) Active() bool { return o.run.Active() }

// Tick is the command for o's next frame, nil at rest.
func (o Odometer) Tick() tea.Cmd { return o.run.Tick() }

// FrameAt is o's frame for at, for a test or a host that delivers frames itself.
func (o Odometer) FrameAt(at time.Time) animation.FrameMsg { return o.run.FrameAt(at) }

// Update turns on its own frames and asks for the next until it rests.
func (o Odometer) Update(msg tea.Msg) (Odometer, bezel.Result) {
	f, ok := o.run.Own(msg)
	if !ok {
		return o, bezel.Result{Handled: ok}
	}
	if o, _ = o.Advance(f.At); !o.Active() {
		return o, bezel.Result{Handled: true}
	}
	var cmd tea.Cmd
	o.run, cmd = o.run.Next()
	return o, bezel.Result{Handled: true, Cmd: cmd}
}

// Advance is o at now, and whether it came to rest.
func (o Odometer) Advance(now time.Time) (Odometer, bool) {
	if !o.Active() {
		return o, true
	}
	o.run, o.p = o.run.Advance(now)
	return o, !o.Active()
}

// WithEase is o following e instead of EaseOut.
func (o Odometer) WithEase(e animation.Ease) Odometer {
	o.run.Ease = e
	return o
}

// wrap10 is n's last digit, counted round from 9 when a curve that pulls back
// takes it below zero.
func wrap10(n int) int { return (n%10 + 10) % 10 }
