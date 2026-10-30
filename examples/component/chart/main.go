// Chart: one run of numbers drawn as an area and a spark line, fed a value per
// tick, beside labelled bars, a donut and a week's heat grid. space pauses the
// feed and r starts it over.
package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/chart"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// top is the value that fills a chart: the load is a percentage.
const top = 100

type (
	tickMsg    struct{}
	pauseMsg   struct{}
	restartMsg struct{}
)

func tick() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

// load is the n-th value of a made-up CPU load: two waves and a ripple, the
// same every run so the recording and the tests agree.
func load(n int) float64 {
	x := float64(n)
	v := 50 + 25*math.Sin(x/9) + 12*math.Sin(x/3.1) + 6*math.Sin(x*1.7)
	return math.Max(2, math.Min(98, v))
}

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	n      int
	paused bool
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(
			layout.Fill("load"),
			layout.Fixed("services", layout.Ratio(2, 5), layout.Min(34)),
		),
		Theme: th,
		Title: "chart",
		Actions: []shell.Action{
			shell.Custom("space", "pause", shell.Send(pauseMsg{})),
			shell.Custom("r", "restart", shell.Send(restartMsg{})),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th, n: 40}
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
	case restartMsg:
		m.n = 40
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// values is the series up to now, newest last.
func (m model) values() []float64 {
	out := make([]float64, m.n)
	for i := range out {
		out[i] = load(i)
	}
	return out
}

func (m model) loadPane(r layout.Rect) string {
	vs := m.values()
	now := vs[len(vs)-1]
	state := "live"
	if m.paused {
		state = "paused"
	}
	out := []string{
		draw.Heading("Area", m.th.Section, r.W),
		chart.Area(vs, top, r.W, max(r.H-8, 3), m.th),
		"",
		draw.Heading("Spark", m.th.Section, r.W),
		chart.Spark(vs, top, r.W-12, m.th, chart.InSeries(1)) + m.th.Dim.Render(fmt.Sprintf("  %3.0f%%", now)),
		"",
		m.th.Muted.Render(fmt.Sprintf("%d samples, %s", len(vs), state)),
	}
	return strings.Join(out, "\n")
}

func (m model) servicesPane(r layout.Rect) string {
	names := []string{"api", "worker", "search", "billing", "mailer"}
	bars := make([]chart.Bar, len(names))
	for i, name := range names {
		v := load(m.n + i*7)
		bars[i] = chart.Bar{Label: name, Value: v, Text: fmt.Sprintf("%.0f%%", v)}
		if v > 80 {
			s := lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Colors.Danger))
			bars[i].Style = &s
		}
	}
	week := make([][]float64, 7)
	for d := range week {
		week[d] = make([]float64, 12)
		for h := range week[d] {
			week[d][h] = load(d*12+h) - 30
		}
	}
	done := float64(m.n % 120)
	out := []string{
		draw.Heading("Bars", m.th.Section, r.W),
		chart.Bars(bars, top, r.W, m.th),
		"",
		draw.Heading("Heat", m.th.Section, r.W),
		chart.Heat(week, 70, m.th),
		"",
		draw.Heading("Donut", m.th.Section, r.W),
		chart.Donut(done, 120, 4, m.th),
	}
	return strings.Join(out, "\n")
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"load":     {Title: "cpu load", Body: m.loadPane},
		"services": {Title: "services", Body: m.servicesPane},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
