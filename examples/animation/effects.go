package main

import (
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/animation/text"
	"github.com/lucasassuncao/bezel/theme"
)

const (
	loading = "Loading your library…"
	serial  = "7f3a-91bc-44de-0b2e"
	title   = "bezel · animation · gallery"
	field   = 4 // rows in the point's field; it is track cells wide
)

type (
	scrambleMsg struct{}
	odometerMsg struct{}
	moveMsg     struct{}
)

// readings are the values the odometer rolls between, round and round.
var readings = []int{1284, 98765, 40210, 7}

// corners are where the points go next, in turn.
var corners = []animation.Point{{X: track - 1, Y: field - 1}, {X: 0, Y: field - 1}, {X: track - 1, Y: 0}, {X: 0, Y: 0}}

// effects is the second tab: the shimmer and the gradient keep their own
// ticks, as the marquee does; the rest run on the gallery's clock.
type effects struct {
	shimmer  text.Shimmer
	gradient text.Gradient
	scramble text.Scramble
	odometer text.Odometer
	reading  int
	move     animation.Move
	follow   animation.Follow
	corner   int
}

func newEffects(th theme.Resolved) effects {
	stops := []color.Color{th.Accent.GetForeground(), th.Info.GetForeground(), th.Success.GetForeground(), th.Warning.GetForeground()}
	return effects{
		shimmer:  text.NewShimmer(loading, th.Muted.GetForeground(), th.Accent.GetForeground(), 1200*time.Millisecond),
		gradient: text.NewGradient(title, stops, 2*time.Second),
		scramble: text.NewScramble(serial, 0),
		odometer: text.NewOdometer(0, 0, 900*time.Millisecond),
		move:     animation.NewMove(animation.Point{}, animation.Point{}, 0),
		follow:   animation.NewFollow(animation.Point{}).WithDamping(0.4),
	}
}

// start sets the shimmer and the gradient going on their own ticks.
func (e effects) start() (effects, tea.Cmd) {
	return e, tea.Batch(e.shimmer.Tick(), e.gradient.Tick())
}

// decode scrambles the serial again.
func (e effects) decode() effects {
	e.scramble = text.NewScramble(serial, 900*time.Millisecond)
	return e
}

// roll turns the odometer on to the next reading.
func (e effects) roll() effects {
	e.odometer, _ = e.odometer.To(readings[e.reading%len(readings)])
	e.reading++
	return e
}

// travel sends both points to the next corner: one eased, one on a spring.
func (e effects) travel() effects {
	to := corners[e.corner%len(corners)]
	e.corner++
	e.move = animation.NewMove(e.move.Cur, to, 700*time.Millisecond)
	e.follow, _ = e.follow.To(to)
	return e
}

// update hands a frame to the two that keep their own ticks.
func (e effects) update(msg tea.Msg) (effects, tea.Cmd) {
	var a, b tea.Cmd
	e.shimmer, a = e.shimmer.Update(msg)
	e.gradient, b = e.gradient.Update(msg)
	return e, tea.Batch(a, b)
}

// advance moves the ones on the shared clock to at.
func (e effects) advance(at time.Time) effects {
	e.scramble, _ = e.scramble.Advance(at)
	e.odometer, _ = e.odometer.Advance(at)
	e.move, _ = e.move.Advance(at)
	e.follow, _ = e.follow.Advance(at)
	return e
}

// moving reports whether anything here needs the shared clock.
func (e effects) moving() bool {
	return e.scramble.Active() || e.odometer.Active() || e.move.Active() || e.follow.Active()
}

// rows is the tab's body, labelled as the first tab's is.
func (e effects) rows(th theme.Resolved, label func(kind, variant string) string) string {
	rows := []string{
		label("shimmer", "sweep forever") + e.shimmer.View(),
		label("scramble", "x settles") + th.Bold.Render(e.scramble.View()),
		label("odometer", "o rolls on") + th.Bold.Render(e.odometer.View()),
		label("gradient", "flows forever") + e.gradient.View(),
		"",
	}
	for y, line := range e.points(th) {
		kind, variant := "", ""
		switch y {
		case 0:
			kind, variant = "point", "m travels"
		case 1:
			variant = "● eased"
		case 2:
			variant = "○ sprung"
		}
		rows = append(rows, label(kind, variant)+line)
	}
	return strings.Join(rows, "\n")
}

// points is the field, a row each, with the eased point over the sprung one
// where they meet; either is kept on the field when it overshoots.
func (e effects) points(th theme.Resolved) []string {
	clamp := func(p animation.Point) (int, int) {
		x, y := p.Round()
		return min(max(x, 0), track-1), min(max(y, 0), field-1)
	}
	mx, my := clamp(e.move.Cur)
	fx, fy := clamp(e.follow.Pos())
	lines := make([]string, field)
	for y := range lines {
		var b strings.Builder
		for x := range track {
			switch {
			case x == mx && y == my:
				b.WriteString(th.Accent.Render("●"))
			case x == fx && y == fy:
				b.WriteString(th.Info.Render("○"))
			default:
				b.WriteString(th.Muted.Render("·"))
			}
		}
		lines[y] = b.String()
	}
	return lines
}
