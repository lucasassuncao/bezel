// Log pane: a made-up deploy logging through slog from another goroutine into
// logger.Pane, which follows the end as lines arrive. Scrolling up holds it
// still and counts what arrives meanwhile; end follows again.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/logger"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// every is how often the deploy logs a line.
const every = 350 * time.Millisecond

// event is one line the deploy logs.
type event struct {
	level slog.Level
	msg   string
	attrs []any
}

var events = []event{
	{slog.LevelInfo, "pulling image", []any{"tag", "api:1.42.0"}},
	{slog.LevelDebug, "layer cached", []any{"digest", "sha256:9f2c…e1"}},
	{slog.LevelInfo, "starting replica", []any{"pod", "api-7d9f", "node", "worker-2"}},
	{slog.LevelWarn, "health check slow", []any{"pod", "api-7d9f", "took", "1.8s"}},
	{slog.LevelInfo, "replica ready", []any{"pod", "api-7d9f"}},
	{slog.LevelDebug, "draining old replica", []any{"pod", "api-5c1a"}},
	{slog.LevelError, "connection refused", []any{"host", "db-primary:5432", "retry", 3}},
	{slog.LevelInfo, "traffic shifted", []any{"to", "api:1.42.0", "share", "100%"}},
}

// step is a line logged: the next one is due.
type step struct{ n int }

type model struct {
	sh   shell.Shell
	th   theme.Resolved
	pane logger.Pane
	log  *slog.Logger
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Rows(layout.Fill("log")),
		Theme:  th,
		Title:  "log pane",
		Actions: []shell.Action{
			shell.Scroll(),
			shell.Custom("end", "follow", nil, shell.DisplayOnly()),
			shell.Quit(),
		},
	})
	pane := logger.NewPane(th)
	log := slog.New(pane.Handler().WithLevel(slog.LevelDebug).WithTimeFormat("15:04:05"))
	return model{sh: sh, th: th, pane: pane, log: log}
}

// Init waits for the pane's lines and starts the deploy.
func (m model) Init() tea.Cmd { return tea.Batch(m.pane.Init(), m.next(0)) }

// next logs line n after a pause, from the command's own goroutine.
func (m model) next(n int) tea.Cmd {
	return tea.Tick(every, func(time.Time) tea.Msg { return m.emit(n) })
}

// emit logs line n of the deploy, round the events.
func (m model) emit(n int) tea.Msg {
	e := events[n%len(events)]
	m.log.Log(context.Background(), e.level, e.msg, e.attrs...)
	return step{n + 1}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		in := draw.InnerRect(m.sh.Rect("log"))
		m.pane = m.pane.SetWidth(in.W).SetHeight(in.H)
		return m, nil
	case step:
		return m, m.next(msg.n)
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	m.pane, r = m.pane.Update(msg) // keys, the wheel and the lines
	return m, r.Cmd
}

func (m model) View() tea.View {
	state := "following"
	if !m.pane.Following() {
		state = "held"
	}
	return m.sh.View(map[string]shell.Pane{
		"log": {
			Title:  "deploy",
			Bottom: fmt.Sprintf("%d lines · %s", m.pane.Len(), state),
			Scroll: m.pane.Scroll(),
			Body:   func(layout.Rect) string { return m.pane.View() },
		},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
