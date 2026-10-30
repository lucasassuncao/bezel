package main

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/component/chart"
	"github.com/lucasassuncao/bezel/theme"
)

const (
	fieldW = 30 // the burst's field, in columns
	fieldH = 6  // and rows
	ring   = 6  // the donut's rows
)

type (
	burstMsg  struct{} // x: a burst in the middle of the field
	rankMsg   struct{} // v: the race takes new values
	sweepMsg  struct{} // o: the donut sweeps to its next share
	squashMsg struct{} // d: the ball drops again
)

var (
	racers = []string{"api", "db", "cache", "queue"}
	// standings are the values the race moves between.
	standings = [][]float64{{40, 75, 20, 55}, {85, 30, 60, 45}, {25, 50, 90, 70}}
	shares    = []float64{0.25, 0.8, 0.55, 1}
)

// pops is the tenth tab: a burst of sparks, a donut that sweeps, a race
// of bars that overtake each other, and a ball that squashes when it lands.
type pops struct {
	burst    animation.Burst
	donut    animation.Float
	share    int
	race     animation.Race
	standing int
	ball     animation.Drop
}

func newPops() pops {
	return pops{
		burst: animation.NewBurst(animation.Point{}, 0, 0),
		donut: animation.NewFloat(0, shares[0], 0),
		race:  animation.NewRace(racers, standings[0], 900*time.Millisecond),
		ball:  animation.NewDrop(0, track-1),
	}
}

func (p pops) pop() pops {
	p.burst = animation.NewBurst(animation.Point{X: fieldW / 2, Y: fieldH / 2}, 28, 1200*time.Millisecond).
		WithSpeed(18).WithSeed(uint64(time.Now().UnixNano())) // #nosec G115 -- any seed
	return p
}

func (p pops) rank() pops {
	p.standing = (p.standing + 1) % len(standings)
	p.race, _ = p.race.Set(standings[p.standing])
	return p
}

func (p pops) sweep() pops {
	from := p.donut.Value()
	p.share = (p.share + 1) % len(shares)
	p.donut = animation.NewFloat(from, shares[p.share], 900*time.Millisecond).WithEase(animation.Out(animation.Cubic))
	return p
}

func (p pops) dropBall() pops {
	p.ball = animation.NewDrop(0, track-1)
	return p
}

func (p pops) advance(at time.Time) pops {
	p.burst, _ = p.burst.Advance(at)
	p.donut, _ = p.donut.Advance(at)
	p.race, _ = p.race.Advance(at)
	p.ball, _ = p.ball.Advance(at)
	return p
}

func (p pops) moving() bool {
	return p.burst.Active() || p.donut.Active() || p.race.Active() || p.ball.Active()
}

// rows is the tab's body: the burst beside the donut, the race, the ball.
func (p pops) rows(th theme.Resolved, label func(kind, variant string) string) string {
	field := p.field(th)
	donut := chart.Donut(p.donut.Value(), 1, ring, th)
	top := lipgloss.JoinHorizontal(lipgloss.Top, field, "   ", donut)
	var rows []string
	for i, line := range strings.Split(top, "\n") {
		kind, variant := "", ""
		switch i {
		case 0:
			kind, variant = "burst", "x pops"
		case 1:
			kind, variant = "donut", "o sweeps"
		}
		rows = append(rows, label(kind, variant)+line)
	}
	rows = append(rows, "")
	standing := p.race.Racers()
	slices.SortStableFunc(standing, func(a, b animation.Racer) int { return cmp.Compare(a.Place, b.Place) })
	for i, r := range standing {
		kind, variant := "", ""
		if i == 0 {
			kind, variant = "race", "v re-ranks"
		}
		n := min(max(int(math.Round(r.Value/100*float64(track-8))), 0), track-8)
		rows = append(rows, label(kind, variant)+fmt.Sprintf("%-6s", r.Label)+
			th.Series[slices.Index(racers, r.Label)%len(th.Series)].Render(strings.Repeat("█", n))+th.Muted.Render(fmt.Sprintf(" %2.0f", r.Value)))
	}
	return strings.Join(append(rows, "", label("squash", "d drops")+p.squashed(th)), "\n")
}

// field is the burst's sparks on a dark field, brighter the more life left.
func (p pops) field(th theme.Resolved) string {
	at := map[[2]int]string{}
	for i, s := range p.burst.Sparks() {
		x, y := s.At.Round()
		if x < 0 || x >= fieldW || y < 0 || y >= fieldH {
			continue
		}
		glyph := "·"
		switch {
		case s.Life > 0.6:
			glyph = "✦"
		case s.Life > 0.3:
			glyph = "*"
		}
		at[[2]int{x, y}] = th.Series[i%len(th.Series)].Render(glyph)
	}
	lines := make([]string, fieldH)
	for y := range fieldH {
		var b strings.Builder
		for x := range fieldW {
			if s, ok := at[[2]int{x, y}]; ok {
				b.WriteString(s)
			} else {
				b.WriteString(" ")
			}
		}
		lines[y] = b.String()
	}
	return strings.Join(lines, "\n")
}

// squashed is the drop's track with the ball drawn as its shape has it: tall
// while fast, wide and flat against the wall, round at rest.
func (p pops) squashed(th theme.Resolved) string {
	at := min(max(int(math.Round(p.ball.Value())), 0), track-1)
	w, h := p.ball.Shape()
	glyph := "●"
	switch {
	case w > 1.2:
		glyph = "▌"
	case h > 1.15:
		glyph = "◆"
	}
	return th.Muted.Render(strings.Repeat("·", at)) + th.Accent.Render(glyph) +
		th.Muted.Render(strings.Repeat("·", track-1-at)) + th.Bold.Render("▐")
}
