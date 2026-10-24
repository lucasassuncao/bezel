package main

import (
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/theme"
)

const (
	swing  = 10 // cells the pendulum swings either way
	limit  = 20 // the rubber's top; the track runs past it so the give shows
	sky    = 5  // rows the orbit takes
	flight = 26 // columns the orbit takes
)

type (
	dropMsg   struct{} // d: drop again
	flockMsg  struct{} // f: the flock flies to the next corner
	rubberMsg struct{} // b: pull the rubber, or let it go
)

// physics is the ninth tab. The pendulum and the orbit go on their own ticks;
// the drop, the flock and the rubber run on the gallery's clock.
type physics struct {
	drop   animation.Drop
	pend   animation.Pendulum
	orbit  animation.Orbit
	flock  animation.Flock
	corner int
	rubber animation.Rubber
	pulled bool
}

func newPhysics() physics {
	return physics{
		drop:   animation.NewDrop(0, track-1),
		pend:   animation.NewPendulum(swing, 1600*time.Millisecond),
		orbit:  animation.NewOrbit(animation.Point{X: flight / 2, Y: sky / 2}, flight/2-1, sky/2, 2*time.Second),
		flock:  animation.NewFlock(5, animation.Point{}),
		rubber: animation.NewRubber(0, limit).WithGive(6),
	}
}

func (p physics) start() (physics, tea.Cmd) {
	return p, tea.Batch(p.pend.Tick(), p.orbit.Tick())
}

func (p physics) dropAgain() physics {
	p.drop = animation.NewDrop(0, track-1)
	return p
}

// fly sends the flock's leader to the next corner of the field.
func (p physics) fly() physics {
	p.flock, _ = p.flock.To(corners[p.corner%len(corners)])
	p.corner++
	return p
}

// stretch pulls the rubber far past its top, or lets it spring back.
func (p physics) stretch() physics {
	if p.pulled {
		p.rubber, _ = p.rubber.Release()
	} else {
		p.rubber = p.rubber.Pull(limit + 40)
	}
	p.pulled = !p.pulled
	return p
}

func (p physics) update(msg tea.Msg) (physics, tea.Cmd) {
	var a, b tea.Cmd
	p.pend, a = p.pend.Update(msg)
	p.orbit, b = p.orbit.Update(msg)
	return p, tea.Batch(a, b)
}

func (p physics) advance(at time.Time) physics {
	p.drop, _ = p.drop.Advance(at)
	p.flock, _ = p.flock.Advance(at)
	p.rubber, _ = p.rubber.Advance(at)
	return p
}

func (p physics) moving() bool { return p.drop.Active() || p.flock.Active() || p.rubber.Active() }

func (p physics) rows(th theme.Resolved, label func(kind, variant string) string) string {
	rows := []string{
		label("drop", "d again") + p.dropped(th),
		label("pendulum", "forever") + p.swung(th),
		"",
	}
	for y, line := range p.orbiting(th) {
		kind, variant := "", ""
		if y == sky/2 {
			kind, variant = "orbit", "near over far"
		}
		rows = append(rows, label(kind, variant)+line)
	}
	rows = append(rows, "")
	for y, line := range p.flocking(th) {
		kind, variant := "", ""
		if y == 0 {
			kind, variant = "flock", "f flies"
		}
		rows = append(rows, label(kind, variant)+line)
	}
	return strings.Join(append(rows, "", label("rubber", "b stretches")+p.stretched(th)), "\n")
}

// dropped is the track with the ball where it is and the floor at the end.
func (p physics) dropped(th theme.Resolved) string {
	at := min(max(int(math.Round(p.drop.Cur)), 0), track-1)
	return th.Muted.Render(strings.Repeat("·", at)) + th.Accent.Render("●") +
		th.Muted.Render(strings.Repeat("·", track-1-at)) + th.Bold.Render("▐")
}

// swung is a rod with the bob where the pendulum has it.
func (p physics) swung(th theme.Resolved) string {
	at := min(max(swing+int(math.Round(p.pend.Cur)), 0), 2*swing)
	cells := []rune(strings.Repeat("─", 2*swing+1))
	cells[swing] = '┼'
	cells[at] = '●'
	return th.Muted.Render(string(cells[:at])) + th.Accent.Render("●") + th.Muted.Render(string(cells[at+1:]))
}

// orbiting is the sky with a sun in the middle and the point going round it,
// drawn behind the sun on the far half.
func (p physics) orbiting(th theme.Resolved) []string {
	px, py := p.orbit.Pos().Round()
	cx, cy := flight/2, sky/2
	lines := make([]string, sky)
	for y := range sky {
		var b strings.Builder
		for x := range flight {
			switch {
			case x == px && y == py && (p.orbit.Near() || x != cx || y != cy):
				glyph := "●"
				if !p.orbit.Near() {
					glyph = "∙"
				}
				b.WriteString(th.Info.Render(glyph))
			case x == cx && y == cy:
				b.WriteString(th.Warning.Render("◉"))
			default:
				b.WriteString(" ")
			}
		}
		lines[y] = b.String()
	}
	return lines
}

// flocking is the field with the leader and its trail, the trail dimmer.
func (p physics) flocking(th theme.Resolved) []string {
	at := map[[2]int]string{}
	pts := p.flock.Points()
	for i := len(pts) - 1; i >= 0; i-- {
		x, y := pts[i].Round()
		x, y = min(max(x, 0), track-1), min(max(y, 0), field-1)
		at[[2]int{x, y}] = th.Muted.Render("•")
		if i == 0 {
			at[[2]int{x, y}] = th.Accent.Render("●")
		}
	}
	lines := make([]string, field)
	for y := range field {
		var b strings.Builder
		for x := range track {
			if s, ok := at[[2]int{x, y}]; ok {
				b.WriteString(s)
			} else {
				b.WriteString(th.Muted.Render("·"))
			}
		}
		lines[y] = b.String()
	}
	return lines
}

// stretched is the rubber's track up to its top, the give beyond it, and the
// value where it is.
func (p physics) stretched(th theme.Resolved) string {
	at := min(max(int(math.Round(p.rubber.Pos())), 0), track-1)
	var b strings.Builder
	for x := range track {
		switch {
		case x == at:
			b.WriteString(th.Accent.Render("●"))
		case x == limit:
			b.WriteString(th.Bold.Render("┤"))
		case x < limit:
			b.WriteString(th.Muted.Render("─"))
		default:
			b.WriteString(th.Muted.Render("·"))
		}
	}
	return b.String()
}
