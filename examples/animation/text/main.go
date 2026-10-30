// Text: every kind the animation/text package has, a tab for each family:
// typed and scrolled, lit and rolled, marked and moved, arriving and leaving.
// The timed ones share one Clock; the looping ones keep their own ticks.
package main

import (
	"fmt"
	"image/color"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// black is the gallery's dark background, what the rise and the neon fade towards.
var black = color.RGBA{0, 0, 0, 255}

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
	th    theme.Resolved
	clock animation.Clock // drives the timed rows of every tab

	ty typed
	ro rolls
	tx texts
	wd words
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Rows(layout.Fill("main")),
		Theme:  th,
		Title:  "text",
		Tabs: []shell.Tab{
			page{"typed", []shell.Action{
				shell.Custom("t", "type", shell.Send(typeMsg{})),
				shell.Custom("s", "shake", shell.Send(shakeMsg{})),
			}},
			page{"roll", []shell.Action{
				shell.Custom("x", "scramble", shell.Send(scrambleMsg{})),
				shell.Custom("o", "odometer", shell.Send(odometerMsg{})),
			}},
			page{"mark", []shell.Action{
				shell.Custom("d", "decrypt", shell.Send(decryptMsg{})),
				shell.Custom("t", "ticker", shell.Send(tickerMsg{})),
				shell.Custom("l", "highlight", shell.Send(markMsg{})),
				shell.Custom("k", "strike", shell.Send(strikeMsg{})),
				shell.Custom("c", "copy", shell.Send(copyMsg{})),
			}},
			page{"arrive", []shell.Action{
				shell.Custom("s", "stream", shell.Send(streamMsg{})),
				shell.Custom("m", "morph", shell.Send(morphMsg{})),
				shell.Custom("y", "type", shell.Send(typistMsg{})),
				shell.Custom("f", "flip", shell.Send(flipMsg{})),
				shell.Custom("b", "burn", shell.Send(burnMsg{})),
				shell.Custom("a", "fall", shell.Send(fallMsg{})),
			}},
		},
		Actions: []shell.Action{shell.ChangeTab(), shell.Quit()},
	})
	return model{
		sh: sh, th: th, clock: animation.NewClock(),
		ty: newTyped(),
		ro: newRolls(th),
		tx: newTexts(th),
		wd: newWords(th),
	}
}

// Init starts the loops as soon as the program runs.
func (m model) Init() tea.Cmd { return func() tea.Msg { return startMsg{} } }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	if next, cmd, ok := m.key(msg); ok {
		return next.run(cmd)
	}
	switch msg := msg.(type) {
	case startMsg:
		var a, b, c, d tea.Cmd
		m.ty, a = m.ty.start()
		m.ro, b = m.ro.start()
		m.tx, c = m.tx.start()
		m.wd, d = m.wd.start()
		return m.run(tea.Batch(a, b, c, d))
	case animation.FrameMsg:
		return m.frame(msg)
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// key acts on the keys of every tab, and reports whether msg was one of them.
func (m model) key(msg tea.Msg) (model, tea.Cmd, bool) {
	var cmd tea.Cmd
	switch msg.(type) {
	case typeMsg:
		m.ty = m.ty.retype()
	case shakeMsg:
		m.ty = m.ty.refuse()
	case scrambleMsg:
		m.ro = m.ro.decode()
	case odometerMsg:
		m.ro = m.ro.roll()
	case decryptMsg:
		m.tx = m.tx.land()
	case tickerMsg:
		m.tx = m.tx.move()
	case markMsg:
		m.tx = m.tx.remark()
	case strikeMsg:
		m.tx = m.tx.cross()
	case copyMsg:
		m.tx = m.tx.copied()
	case streamMsg:
		m.wd, cmd = m.wd.more()
	case morphMsg:
		m.wd = m.wd.changeWord()
	case typistMsg:
		m.wd = m.wd.retype()
	case flipMsg:
		m.wd = m.wd.turnClock()
	case burnMsg:
		m.wd = m.wd.burnIt()
	case fallMsg:
		m.wd = m.wd.drop()
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

// frame hands the loops their own frames and moves everything else to the
// clock's time.
func (m model) frame(msg animation.FrameMsg) (tea.Model, tea.Cmd) {
	var a, b, c, d, clockCmd tea.Cmd
	m.ty, a = m.ty.update(msg)
	m.ro, b = m.ro.update(msg)
	m.tx, c = m.tx.update(msg)
	m.wd, d = m.wd.update(msg)
	var at time.Time
	var ok bool
	if m.clock, at, ok = m.clock.Take(msg); ok {
		m.ty = m.ty.advance(at)
		m.ro = m.ro.advance(at)
		m.tx = m.tx.advance(at)
		m.wd = m.wd.advance(at)
		if m.moving() {
			m.clock, clockCmd = m.clock.Next()
		} else {
			m.clock = m.clock.Stop()
		}
	}
	return m, tea.Batch(a, b, c, d, clockCmd)
}

// moving reports whether anything on the shared clock still needs frames.
func (m model) moving() bool {
	return m.ty.moving() || m.ro.moving() || m.tx.moving() || m.wd.moving()
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"main": {Title: "every text", Right: still(), Body: func(layout.Rect) string {
			switch m.sh.ActiveTab() {
			case 1:
				return m.ro.rows(m.th, m.label)
			case 2:
				return m.tx.rows(m.th, m.label)
			case 3:
				return m.wd.rows(m.th, m.label)
			}
			return m.ty.rows(m.th, m.label)
		}},
	}))
	v.AltScreen = true
	v.BackgroundColor, v.ForegroundColor = m.th.Background, m.th.Foreground
	return v
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
	return m.th.Accent.Render(fmt.Sprintf("%-11s", kind)) + m.th.Muted.Render(fmt.Sprintf("%-14s", variant))
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
