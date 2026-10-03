// Progress: a count that fills on its own, and a wizard's stages. r restarts
// the count; n and p move through the stages, drawn at the pane's width and
// again at 30 columns, where the stepper folds to "Step 2 of 4".
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/icon"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/progress"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

const total = 40

var steps = []string{"Template", "Features", "Ports", "Review"}

type (
	tickMsg    struct{}
	restartMsg struct{}
	nextMsg    struct{}
	prevMsg    struct{}
)

// tick paces the count: fast enough to watch, slow enough to read.
func tick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

type model struct {
	sh   shell.Shell
	th   theme.Resolved
	done int
	step int
}

func newModel() model {
	th := theme.Resolve(theme.ThemeMint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("progress")),
		Theme:  th,
		Title:  "progress",
		Actions: []shell.Action{
			shell.Custom("r", "restart", shell.Send(restartMsg{})),
			shell.Custom("n", "next step", shell.Send(nextMsg{})),
			shell.Custom("p", "previous step", shell.Send(prevMsg{})),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th}
}

func (m model) Init() tea.Cmd { return tick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case tickMsg:
		if m.done < total {
			m.done++
			return m, tick()
		}
	case restartMsg:
		running := m.done < total
		m.done = 0
		if !running {
			return m, tick()
		}
	case nextMsg:
		m.step = min(len(steps), m.step+1)
	case prevMsg:
		m.step = max(0, m.step-1)
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) body(r layout.Rect) string {
	w := r.W - 2
	out := []string{
		draw.Heading("Line", m.th.Section, r.W),
		" " + progress.Line("downloading", m.done, total, w, m.th),
		"",
		draw.Heading("Bar", m.th.Section, r.W),
		" " + progress.Bar(m.done, total, min(w, 30), m.th),
		"",
		draw.Heading("Stepper", m.th.Section, r.W),
		" " + progress.Stepper(steps, m.step, w, m.th, icon.Unicode),
		"",
		" " + m.th.Muted.Render("at 30 columns"),
		" " + progress.Stepper(steps, m.step, 30, m.th, icon.Unicode),
	}
	return strings.Join(out, "\n")
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"progress": {Title: "progress", Body: m.body},
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
