// Canvas: a small solar system drawn in braille. Orbits and planets are
// circles, the sun's rays lines, the belt points, the names labels and a box
// follows the outer planet. space pauses it and + and - zoom.
package main

import (
	"fmt"
	"math"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/canvas"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

type (
	tickMsg  struct{}
	pauseMsg struct{}
	zoomMsg  float64
)

func tick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

// planet is one body: its orbit's radius, its own, its turns a tick and
// where it starts, in radians.
type planet struct {
	name         string
	orbit, r     float64
	speed, phase float64
	series       int
}

var planets = []planet{
	{"ember", 4, 0.5, 0.050, 0.4, 3},
	{"tide", 7, 0.8, 0.031, 2.6, 1},
	{"dusk", 11, 0.6, 0.018, 4.4, 5},
}

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	n      int
	zoom   float64 // the half-height of the view, in the system's units
	paused bool
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Fill("sky"),
		Theme:  th,
		Title:  "canvas",
		Actions: []shell.Action{
			shell.Custom("space", "pause", shell.Send(pauseMsg{})),
			shell.Custom("+", "zoom in", shell.Send(zoomMsg(0.8))),
			shell.Custom("-", "zoom out", shell.Send(zoomMsg(1.25))),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th, zoom: 13}
}

func (m model) Init() tea.Cmd { return tick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	switch msg := msg.(type) {
	case tickMsg:
		if m.paused {
			return m, nil
		}
		m.n++
		return m, tick()
	case pauseMsg:
		m.paused = !m.paused
		if !m.paused {
			return m, tick()
		}
	case zoomMsg:
		m.zoom = min(max(m.zoom*float64(msg), 3), 40)
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// where is p's place at tick n.
func where(p planet, n int) (x, y float64) {
	a := p.phase + p.speed*float64(n)*math.Pi
	return p.orbit * math.Cos(a), p.orbit * math.Sin(a)
}

// shapes is the system at tick n, back to front.
func shapes(n int) []canvas.Shape {
	var out []canvas.Shape
	for i := range 8 { // the sun's rays, turning slowly
		a := float64(i)*math.Pi/4 + float64(n)*0.02
		out = append(out, canvas.Line{X0: 1.8 * math.Cos(a), Y0: 1.8 * math.Sin(a), X1: 2.6 * math.Cos(a), Y1: 2.6 * math.Sin(a), Series: 4})
	}
	out = append(out, canvas.Circle{R: 1.3, Series: 2}, canvas.Label{X: -1.2, Y: -3.2, Text: "sun", Series: 2})
	belt := make([]canvas.Point, 0, 90)
	for i := range 90 { // the same scatter every run, drifting as one
		a := float64(i)*2.399 + float64(n)*0.01
		r := 8.6 + 1.1*math.Sin(float64(i)*7.3)
		belt = append(belt, canvas.Point{X: r * math.Cos(a), Y: r * math.Sin(a)})
	}
	out = append(out, canvas.Points{Coords: belt, Series: 0})
	for _, p := range planets {
		x, y := where(p, n)
		out = append(out,
			canvas.Circle{R: p.orbit, Series: 3},
			canvas.Circle{X: x, Y: y, R: p.r, Series: p.series},
			canvas.Label{X: x + p.r + 0.4, Y: y + p.r, Text: p.name, Series: p.series},
		)
	}
	last := planets[len(planets)-1]
	x, y := where(last, n)
	return append(out, canvas.Rect{X: x - 1.6, Y: y - 1.6, W: 3.2, H: 3.2, Series: 1})
}

func (m model) sky(r layout.Rect) string {
	// A dot is about as wide as it is tall, so x spans what the width gives.
	half := m.zoom * float64(r.W*2) / float64(max(r.H*4, 1))
	c := canvas.New(-half, half, -m.zoom, m.zoom).SetWidth(r.W).SetHeight(r.H).Draw(shapes(m.n)...)
	return c.View(m.th)
}

func (m model) View() tea.View {
	state := "live"
	if m.paused {
		state = "paused"
	}
	return m.sh.View(map[string]shell.Pane{
		"sky": {Title: fmt.Sprintf("orbits · %s · zoom %.0f", state, m.zoom), Body: m.sky},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
