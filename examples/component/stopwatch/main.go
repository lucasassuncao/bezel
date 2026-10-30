// Stopwatch: time counted up a tenth at a time. space starts and stops it,
// r puts it back to zero, running or not.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/stopwatch"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

type (
	toggleMsg struct{}
	resetMsg  struct{}
)

type model struct {
	sh shell.Shell
	th theme.Resolved
	sw stopwatch.Model
}

func newModel() model {
	th := theme.Resolve(theme.Mint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("stopwatch")),
		Theme:  th,
		Title:  "stopwatch",
		Actions: []shell.Action{
			shell.Custom("space", "start/stop", shell.Send(toggleMsg{})),
			shell.Custom("r", "reset", shell.Send(resetMsg{})),
			shell.Quit(),
		},
	})
	sw := stopwatch.New().WithEvery(time.Second / 10).WithFormat(tenths)
	return model{sh: sh, th: th, sw: sw}
}

// tenths writes d to a tenth of a second, so the width never jumps between
// "900ms" and "1s" as time.Duration's own String would.
func tenths(d time.Duration) string { return fmt.Sprintf("%.1fs", d.Seconds()) }

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case stopwatch.TickMsg:
		var r bezel.Result
		m.sw, r = m.sw.Update(msg)
		return m, r.Cmd
	case toggleMsg:
		m.sw, cmd = m.sw.Toggle()
		return m, cmd
	case resetMsg:
		m.sw = m.sw.Reset()
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) body(r layout.Rect) string {
	out := []string{
		draw.Heading("Elapsed", m.th.Section, r.W),
		"",
		"  " + m.th.Dim.Render("elapsed ") + m.sw.View(m.th),
	}
	return strings.Join(out, "\n")
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"stopwatch": {Title: "stopwatch", Body: m.body},
	}))
	v.AltScreen = true
	v.BackgroundColor, v.ForegroundColor = m.th.Background, m.th.Foreground
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
