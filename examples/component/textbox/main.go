// Textbox: a multi-line value with its lines numbered; every key is text,
// so the app quits on esc and reports on ctrl+s. The mouse selects by dragging.
package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/textbox"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

const seed = "# notes\r\nThe seed uses CRLF; SetValue folds it,\r\nso no line comes back doubled.\r\n"

type countMsg struct{}

type model struct {
	sh shell.Shell
	ta textbox.Model
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("text")),
		Theme:  th,
		Title:  "textbox",
		Mouse:  true,
		Actions: []shell.Action{
			shell.Custom("ctrl+s", "count lines", shell.Send(countMsg{})),
			// q would be text here, so esc quits instead.
			shell.Quit(shell.WithKey("esc")),
		},
	})
	ta, _ := textbox.New(th).SetValue(seed).Focus()
	return model{sh: sh, ta: ta}
}

func (m model) Init() tea.Cmd { return m.ta.Blink() }

// fit sizes the text area to the pane the shell placed.
func (m model) fit() model {
	r := draw.InnerRect(m.sh.Rect("text"))
	m.ta = m.ta.SetWidth(r.W).SetHeight(r.H)
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m.fit(), r.Cmd
	}
	switch msg := msg.(type) {
	case countMsg:
		text := fmt.Sprintf("%d lines, %d characters", m.ta.LineCount(), len(m.ta.Value()))
		var cmd tea.Cmd
		m.sh, cmd = m.sh.SetStatus(text, shell.Info, 3*time.Second)
		return m, cmd
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case shell.ClickMsg: // the press, already in the pane's cells
		m.ta, r = m.ta.Update(msg)
		return m, r.Cmd
	case tea.MouseMotionMsg, tea.MouseReleaseMsg:
		// The drag: moved into the pane even when the pointer left it.
		m.ta, r = m.ta.Update(m.sh.MouseIn("text", msg.(tea.MouseMsg)))
		return m, r.Cmd
	}
	m.ta, r = m.ta.Update(msg)
	return m, r.Cmd
}

func (m model) View() tea.View {
	return m.sh.Screen(map[string]shell.Pane{
		"text": {Title: "notes.md", Body: func(layout.Rect) string { return m.ta.View() }},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
