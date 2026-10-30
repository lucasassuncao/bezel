package main

import (
	"strings"
	"time"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/theme"
)

const (
	field  = 4  // rows in the points' field; it is track cells wide
	routeW = 32 // the path's field, in columns
	routeH = 5  // and rows
)

type (
	moveMsg  struct{} // m: the points go to the next corner
	routeMsg struct{} // p: the marker travels the route again
)

// corners are where the points go next, in turn.
var corners = []animation.Point{{X: track - 1, Y: field - 1}, {X: 0, Y: field - 1}, {X: track - 1, Y: 0}, {X: 0, Y: 0}}

// moves is the third tab: two points that travel, one eased and one on a
// spring, and a marker carried along a curve by a Path.
type moves struct {
	move   animation.Move
	follow animation.Follow
	corner int
	path   animation.Path
}

func newMoves() moves {
	return moves{
		move:   animation.NewMove(animation.Point{}, animation.Point{}, 0),
		follow: animation.NewFollow(animation.Point{}).WithDamping(0.4),
	}.route()
}

// travel sends both points to the next corner: one eased, one on a spring.
func (v moves) travel() moves {
	to := corners[v.corner%len(corners)]
	v.corner++
	v.move = animation.NewMove(v.move.Value(), to, 700*time.Millisecond)
	v.follow, _ = v.follow.To(to)
	return v
}

// route sends the marker along the curve from its start.
func (v moves) route() moves {
	points := animation.Bezier(animation.Point{X: 0, Y: routeH - 1}, animation.Point{X: 8, Y: -3},
		animation.Point{X: 22, Y: routeH + 3}, animation.Point{X: routeW - 1, Y: 0}, 40)
	v.path = animation.NewPath(1800*time.Millisecond, points...)
	return v
}

func (v moves) advance(at time.Time) moves {
	v.move, _ = v.move.Advance(at)
	v.follow, _ = v.follow.Advance(at)
	v.path, _ = v.path.Advance(at)
	return v
}

func (v moves) moving() bool { return v.move.Active() || v.follow.Active() || v.path.Active() }

// rows is the tab's body: the points' field, then the route.
func (v moves) rows(th theme.Resolved, label func(kind, variant string) string) string {
	var rows []string
	for y, line := range v.points(th) {
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
	rows = append(rows, "")
	px, py := v.path.Value().Round()
	for y := range routeH {
		var b strings.Builder
		for x := range routeW {
			if x == px && y == py {
				b.WriteString(th.Accent.Render("●"))
			} else {
				b.WriteString(th.Muted.Render("·"))
			}
		}
		kind, variant := "", ""
		switch y {
		case routeH / 2:
			kind, variant = "path", "bezier"
		case routeH/2 + 1:
			variant = "p again"
		}
		rows = append(rows, label(kind, variant)+b.String())
	}
	return strings.Join(rows, "\n")
}

// points is the field, a row each, with the eased point over the sprung one
// where they meet; either is kept on the field when it overshoots.
func (v moves) points(th theme.Resolved) []string {
	clamp := func(p animation.Point) (int, int) {
		x, y := p.Round()
		return min(max(x, 0), track-1), min(max(y, 0), field-1)
	}
	mx, my := clamp(v.move.Value())
	fx, fy := clamp(v.follow.Value())
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
