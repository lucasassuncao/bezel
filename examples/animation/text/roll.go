package main

import (
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/animation/text"
	"github.com/lucasassuncao/bezel/theme"
)

const (
	loading = "Loading your library…"
	serial  = "7f3a-91bc-44de-0b2e"
	title   = "bezel · animation · gallery"
)

type (
	scrambleMsg struct{} // x: the serial settles again
	odometerMsg struct{} // o: the odometer rolls on
)

// readings are the values the odometer rolls between, round and round.
var readings = []int{1284, 98765, 40210, 7}

// rolls is the second tab: the shimmer and the gradient keep their own ticks,
// as the marquee does; the scramble and the odometer run on the gallery's clock.
type rolls struct {
	shimmer  text.Shimmer
	gradient text.Gradient
	scramble text.Scramble
	odometer text.Odometer
	reading  int
}

func newRolls(th theme.Resolved) rolls {
	stops := []color.Color{th.Accent.GetForeground(), th.Info.GetForeground(), th.Success.GetForeground(), th.Warning.GetForeground()}
	return rolls{
		shimmer:  text.NewShimmer(loading, th.Muted.GetForeground(), th.Accent.GetForeground(), 1200*time.Millisecond),
		gradient: text.NewGradient(title, stops, 2*time.Second),
		scramble: text.NewScramble(serial, 0),
		odometer: text.NewOdometer(0, 0, 900*time.Millisecond),
	}
}

// start sets the shimmer and the gradient going on their own ticks.
func (e rolls) start() (rolls, tea.Cmd) {
	return e, tea.Batch(e.shimmer.Tick(), e.gradient.Tick())
}

// decode scrambles the serial again.
func (e rolls) decode() rolls {
	e.scramble = text.NewScramble(serial, 900*time.Millisecond)
	return e
}

// roll turns the odometer on to the next reading.
func (e rolls) roll() rolls {
	e.odometer, _ = e.odometer.To(readings[e.reading%len(readings)])
	e.reading++
	return e
}

// update hands a frame to the two that keep their own ticks.
func (e rolls) update(msg tea.Msg) (rolls, tea.Cmd) {
	var a, b bezel.Result
	e.shimmer, a = e.shimmer.Update(msg)
	e.gradient, b = e.gradient.Update(msg)
	return e, tea.Batch(a.Cmd, b.Cmd)
}

// advance moves the ones on the shared clock to at.
func (e rolls) advance(at time.Time) rolls {
	e.scramble, _ = e.scramble.Advance(at)
	e.odometer, _ = e.odometer.Advance(at)
	return e
}

// moving reports whether anything here needs the shared clock.
func (e rolls) moving() bool {
	return e.scramble.Active() || e.odometer.Active()
}

// rows is the tab's body, labelled as the first tab's is.
func (e rolls) rows(th theme.Resolved, label func(kind, variant string) string) string {
	rows := []string{
		label("shimmer", "sweep forever") + e.shimmer.View(),
		label("scramble", "x settles") + th.Bold.Render(e.scramble.View()),
		label("odometer", "o rolls on") + th.Bold.Render(e.odometer.View()),
		label("gradient", "flows forever") + e.gradient.View(),
	}
	return strings.Join(rows, "\n")
}
