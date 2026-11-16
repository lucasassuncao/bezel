// Effect: every kind the effect package has, a tab for each family: effects
// over a drawn card, lights, transitions from one block to the next, and old
// screens. The timed ones share one Clock; the loading panel keeps its own.
package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

type startMsg struct{}

// page is a tab of the gallery: its name and the keys that act on it.
type page struct {
	name    string
	actions []shell.Action
}

func (p page) Name() string                         { return p.name }
func (p page) Status(shell.Context) string          { return "" }
func (p page) Actions(shell.Context) []shell.Action { return p.actions }

type model struct {
	sh    shell.Shell
	clock animation.Clock // drives every tab but the loading panel

	dr drawing
	lt lights
	sw swaps
	rt retros
}

func newModel() model {
	sh := shell.New(shell.Config{
		Layout: layout.Rows(layout.Fill("main")),
		Base:   theme.Mint(),
		Title:  "effect",
		Tabs: []shell.Tab{
			page{"draw", []shell.Action{
				shell.Custom("e", "effect", shell.Send(effectMsg{})),
				shell.Custom("p", "pattern", shell.Send(patternMsg{})),
				shell.Custom("r", "again", shell.Send(againMsg{})),
			}},
			page{"light", []shell.Action{
				shell.Custom("e", "light", shell.Send(lightMsg{})),
				shell.Custom("r", "again", shell.Send(shineMsg{})),
				shell.Custom("f", "diff", shell.Send(diffMsg{})),
				shell.Custom("u", "undo", shell.Send(undoMsg{})),
			}},
			page{"swap", []shell.Action{
				shell.Custom("n", "next", shell.Send(swapMsg{})),
				shell.Custom("r", "back", shell.Send(againSwap{})),
			}},
			page{"retro", []shell.Action{
				shell.Custom("e", "effect", shell.Send(retroMsg{})),
				shell.Custom("r", "again", shell.Send(rewindMsg{})),
				shell.Custom("l", "loading", shell.Send(loadingMsg{})),
			}},
		},
		Actions: []shell.Action{shell.ChangeTab(), shell.Quit()},
	})
	th := sh.Theme()
	return model{
		sh: sh, clock: animation.NewClock(),
		lt: newLights(th),
		sw: newSwaps(),
		rt: newRetros(th),
	}
}

// Init plays the first effect, the diff and the loading panel at once.
func (m model) Init() tea.Cmd { return tea.Batch(m.sh.Init(), func() tea.Msg { return startMsg{} }) }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	var cmd tea.Cmd
	if next, cmd, ok := m.key(msg); ok {
		return next.run(cmd)
	}
	switch msg := msg.(type) {
	case startMsg:
		m.dr = m.dr.again(m.sh.Theme())
		m.lt = m.lt.diff(m.sh.Theme())
		m.rt, cmd = m.rt.start()
		return m.run(cmd)
	case animation.FrameMsg:
		return m.frame(msg)
	}
	return m, nil
}

// key acts on the keys of every tab, and reports whether msg was one of them.
func (m model) key(msg tea.Msg) (model, tea.Cmd, bool) {
	th := m.sh.Theme()
	var cmd tea.Cmd
	switch msg.(type) {
	case effectMsg:
		m.dr = m.dr.nextEffect(th)
	case patternMsg:
		m.dr = m.dr.nextPattern(th)
	case againMsg:
		m.dr = m.dr.again(th)
	case lightMsg:
		m.lt = m.lt.next(th)
	case shineMsg:
		m.lt = m.lt.again(th)
	case diffMsg:
		m.lt = m.lt.diff(th)
	case undoMsg:
		m.lt = m.lt.undo(th)
	case swapMsg:
		m.sw = m.sw.next()
	case againSwap:
		m.sw = m.sw.again()
	case retroMsg:
		m.rt = m.rt.next(th)
	case rewindMsg:
		m.rt = m.rt.again(th)
	case loadingMsg:
		m.rt, cmd = m.rt.toggleLoading(th)
	default:
		return m, nil, false
	}
	return m, cmd, true
}

// run starts the shared clock, if it is not running, beside cmd.
func (m model) run(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	var tick tea.Cmd
	m.clock, tick = m.clock.Start()
	return m, tea.Batch(cmd, tick)
}

// frame hands the loading panel its own frames and moves everything else to
// the clock's time.
func (m model) frame(msg animation.FrameMsg) (tea.Model, tea.Cmd) {
	var rtCmd, clockCmd tea.Cmd
	m.rt, rtCmd = m.rt.update(msg)
	var at time.Time
	var ok bool
	if m.clock, at, ok = m.clock.Take(msg); ok {
		m.dr = m.dr.advance(at)
		m.lt = m.lt.advance(at)
		m.sw = m.sw.advance(at)
		m.rt = m.rt.advance(at)
		if m.moving() {
			m.clock, clockCmd = m.clock.Next()
		} else {
			m.clock = m.clock.Stop()
		}
	}
	return m, tea.Batch(rtCmd, clockCmd)
}

// moving reports whether anything on the shared clock still needs frames.
func (m model) moving() bool {
	return m.dr.moving() || m.lt.moving() || m.sw.moving() || m.rt.moving()
}

func (m model) View() tea.View {
	th := m.sh.Theme()
	return m.sh.View(map[string]shell.Pane{
		"main": {Title: "every effect", Right: still(), Body: func(layout.Rect) string {
			switch m.sh.ActiveTab() {
			case 1:
				return m.lt.rows(th, m.label)
			case 2:
				return m.sw.rows(th, m.label)
			case 3:
				return m.rt.rows(th, m.label)
			}
			return m.dr.rows(th, m.label)
		}},
	})
}

// still names the switch when motion is off, so a frozen gallery says why.
func still() string {
	if animation.Still() {
		return "motion off"
	}
	return ""
}

// label is a row's kind and variant, in two fixed columns.
func (m model) label(kind, variant string) string {
	return m.sh.Theme().Accent.Render(fmt.Sprintf("%-11s", kind)) + m.sh.Theme().Muted.Render(fmt.Sprintf("%-14s", variant))
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
