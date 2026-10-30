package main

import (
	"strings"
	"time"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/theme"
)

// margin is the cells left either side of a curve's run, so an overshoot shows.
const margin = 4

type (
	curvesMsg struct{} // r on the curves tab: race again
	turnMsg   struct{} // i: in, out, in-out, round again
)

// family is every curve the package has, raced in this order.
var family = []struct {
	name string
	ease animation.Ease
}{
	{"linear", animation.Linear}, {"sine", animation.Sine}, {"quad", animation.Quad},
	{"cubic", animation.Cubic}, {"quart", animation.Quart}, {"quint", animation.Quint},
	{"expo", animation.Expo}, {"circ", animation.Circ}, {"back", animation.Back},
	{"elastic", animation.Elastic}, {"bounce", animation.Bounce}, {"steps 6", animation.Steps(6)},
}

// turns are the ways a curve is played: arriving first, as most motion should.
var turns = []struct {
	name string
	turn func(animation.Ease) animation.Ease
}{
	{"out", animation.Out},
	{"in", func(e animation.Ease) animation.Ease { return e }},
	{"in-out", animation.InOut},
}

// curves is the third tab: a dot per curve, all racing on the gallery's clock.
type curves struct {
	turn int
	dots []animation.Float
}

// race sets every dot off again, played the current way.
func (c curves) race() curves {
	c.dots = make([]animation.Float, len(family))
	for i, f := range family {
		c.dots[i] = animation.NewFloat(0, 1, 1200*time.Millisecond).WithEase(turns[c.turn].turn(f.ease))
	}
	return c
}

// next plays every curve the next way, and races them.
func (c curves) next() curves {
	c.turn = (c.turn + 1) % len(turns)
	return c.race()
}

func (c curves) advance(at time.Time) curves {
	for i := range c.dots {
		c.dots[i], _ = c.dots[i].Advance(at)
	}
	return c
}

func (c curves) moving() bool {
	for _, d := range c.dots {
		if d.Active() {
			return true
		}
	}
	return false
}

// rows is the tab's body: the way they are played in the kind column, then a
// track per curve with its start and end marked.
func (c curves) rows(th theme.Resolved, label func(kind, variant string) string) string {
	rows := make([]string, len(family))
	for i, f := range family {
		kind := ""
		switch i {
		case 0:
			kind = "curves"
		case 1:
			kind = turns[c.turn].name
		}
		v := 1.0
		if i < len(c.dots) {
			v = c.dots[i].Cur
		}
		rows[i] = label(kind, f.name) + dotOn(th, v)
	}
	return strings.Join(rows, "\n")
}

// dotOn is a track with the dot v of the way from its start mark to its end
// mark, kept on the track when a curve throws it past either.
func dotOn(th theme.Resolved, v float64) string {
	run := float64(track - 1 - 2*margin)
	at := min(max(margin+int(v*run+0.5), 0), track-1)
	var b strings.Builder
	for x := range track {
		switch x {
		case at:
			b.WriteString(th.Accent.Render("●"))
		case margin, track - 1 - margin:
			b.WriteString(th.Muted.Render("│"))
		default:
			b.WriteString(th.Muted.Render("·"))
		}
	}
	return b.String()
}
