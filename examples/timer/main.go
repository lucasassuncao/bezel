// Timer: ten seconds counted down a tenth at a time. space stops and resumes
// it, r starts it over, and the status row says when time is up.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
	"github.com/lucasassuncao/bezel/timer"
)

const timeout = 10 * time.Second

type (
	toggleMsg  struct{}
	restartMsg struct{}
)

type model struct {
	sh shell.Shell
	th theme.Resolved
	tm timer.Model
}

func newModel() model {
	th := theme.Resolve(theme.Mint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("timer")),
		Theme:  th,
		Title:  "timer",
		Actions: []shell.Action{
			shell.Custom("space", "stop/resume", shell.Send(toggleMsg{})),
			shell.Custom("r", "restart", shell.Send(restartMsg{})),
			shell.Quit(),
		},
	})
	tm := timer.New(timeout).WithEvery(time.Second / 10).WithFormat(tenths)
	return model{sh: sh, th: th, tm: tm}
}

// tenths writes d to a tenth of a second, so the width never jumps between
// "1.1s" and "900ms" as time.Duration's own String would.
func tenths(d time.Duration) string { return fmt.Sprintf("%.1fs", d.Seconds()) }

// Init starts the countdown the way r does.
func (m model) Init() tea.Cmd { return func() tea.Msg { return restartMsg{} } }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case timer.TickMsg:
		m.tm, cmd = m.tm.Update(msg)
		return m, cmd
	case timer.TimeoutMsg:
		if msg.ID == m.tm.ID() {
			m.sh, cmd = m.sh.SetStatus("time's up", shell.OK, 3*time.Second)
			return m, cmd
		}
	case toggleMsg:
		m.tm, cmd = m.tm.Toggle()
		return m, cmd
	case restartMsg:
		m.tm, cmd = m.tm.Reset(timeout).Start()
		return m, cmd
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) body(r layout.Rect) string {
	out := []string{
		draw.Heading("Countdown", m.th.Section, r.W),
		"",
		"  " + m.th.Dim.Render("left ") + m.tm.View(m.th),
	}
	return strings.Join(out, "\n")
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"timer": {Title: "timer", Body: m.body},
	}))
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
