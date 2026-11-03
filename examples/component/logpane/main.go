// Log pane: a made-up deploy logging through slog into two logger.Panes,
// bezel's Handler above and slog's JSON through Writer below. Scrolling up
// holds a pane still, end follows again and / filters the focused one.
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
const every = 600 * time.Millisecond

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
	sh     shell.Shell
	th     theme.Resolved
	pane   logger.Pane // bezel's Handler
	json   logger.Pane // slog's JSON handler, through Writer
	log    *slog.Logger
	jsonLg *slog.Logger
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Rows(layout.Fill("log"), layout.Fill("json")),
		Theme:  th,
		Title:  "log pane",
		Actions: []shell.Action{
			shell.ChangePane(),
			shell.Scroll(),
			shell.Custom("end", "follow", nil, shell.DisplayOnly()),
			shell.Custom("/", "filter", nil, shell.DisplayOnly()),
			shell.Quit(),
		},
	})
	pane := logger.NewPane(th)
	log := slog.New(pane.Handler().WithLevel(slog.LevelDebug).WithTimeFormat("15:04:05"))
	json := logger.NewPane(th)
	jsonLg := slog.New(slog.NewJSONHandler(json.Writer(), &slog.HandlerOptions{
		Level: slog.LevelDebug, ReplaceAttr: clock,
	}))
	return model{sh: sh, th: th, pane: pane, json: json, log: log, jsonLg: jsonLg}
}

// clock writes the time as the bezel pane does, so the two lines line up.
func clock(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		a.Value = slog.StringValue(a.Value.Time().Format("15:04:05"))
	}
	return a
}

// Init waits for both panes' lines and starts the deploy.
func (m model) Init() tea.Cmd { return tea.Batch(m.pane.Init(), m.json.Init(), m.next(0)) }

// next logs line n after a pause, from the command's own goroutine.
func (m model) next(n int) tea.Cmd {
	return tea.Tick(every, func(time.Time) tea.Msg { return m.emit(n) })
}

// emit logs line n of the deploy, round the events.
func (m model) emit(n int) tea.Msg {
	e := events[n%len(events)]
	m.log.Log(context.Background(), e.level, e.msg, e.attrs...)
	m.jsonLg.Log(context.Background(), e.level, e.msg, e.attrs...)
	return step{n + 1}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// While filtering every key is text for the filter, tab and q included.
	if k, ok := msg.(tea.KeyPressMsg); ok && m.focused().Filtering() {
		return m.toFocused(k)
	}
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		in := draw.InnerRect(m.sh.Rect("log"))
		m.pane = m.pane.SetWidth(in.W).SetHeight(in.H)
		in = draw.InnerRect(m.sh.Rect("json"))
		m.json = m.json.SetWidth(in.W).SetHeight(in.H)
		return m, nil
	case step:
		return m, m.next(msg.n)
	case tea.KeyPressMsg, tea.MouseWheelMsg:
		if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m.toFocused(msg)
	}
	// Lines: each pane takes only its own feed's.
	var rj bezel.Result
	m.pane, r = m.pane.Update(msg)
	m.json, rj = m.json.Update(msg)
	return m, tea.Batch(r.Cmd, rj.Cmd)
}

// focused is the pane the keys go to.
func (m model) focused() logger.Pane {
	if m.sh.Focused() == "json" {
		return m.json
	}
	return m.pane
}

// toFocused hands msg to the focused pane.
func (m model) toFocused(msg tea.Msg) (tea.Model, tea.Cmd) {
	p, r := m.focused().Update(msg)
	if m.sh.Focused() == "json" {
		m.json = p
	} else {
		m.pane = p
	}
	return m, r.Cmd
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"log":  panel("deploy · bezel", m.pane),
		"json": panel("deploy · json", m.json),
	})
}

// panel draws p with how many lines it holds and whether it follows.
func panel(title string, p logger.Pane) shell.Pane {
	state := "following"
	if !p.Following() {
		state = "held"
	}
	lines := fmt.Sprintf("%d lines", p.Len())
	if p.Filter() != "" {
		lines = fmt.Sprintf("%d of %d lines", p.Shown(), p.Len())
	}
	return shell.Pane{
		Title:  title,
		Bottom: lines + " · " + state,
		Scroll: p.Scroll(),
		Body:   func(layout.Rect) string { return p.View() },
	}
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
