// Input: three one-line fields. A name that must not be empty, a command that
// tab completes from suggestions, and a password behind a mask. ↑/↓ move
// between them, enter submits; every key is text, so esc quits.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/input"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

var commands = []string{"deploy", "describe", "delete", "logs", "restart", "rollback", "status"}

var labels = []string{"name", "command", "password"}

// prompt leads the command field, and takes its width from the text's.
const prompt = "$ "

// labelWidth lines the fields up after the longest label.
const labelWidth = 10

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	fields []input.Model
	focus  int
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("input")),
		Theme:  th,
		Title:  "input",
		Actions: []shell.Action{
			// q would be text here, so esc quits instead.
			shell.Quit(shell.WithKey("esc")),
		},
	})
	name := input.New(th).WithPlaceholder("who is deploying").
		WithValidate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("a name is required")
			}
			return nil
		})
	command := input.New(th).WithPrompt(prompt).WithPlaceholder("tab completes").SetSuggestions(commands)
	password := input.New(th).WithPlaceholder("never shown").WithMask('•')
	fields := []input.Model{name, command, password}
	fields[0], _ = fields[0].Focus()
	return model{sh: sh, th: th, fields: fields}
}

func (m model) Init() tea.Cmd { return m.fields[0].Blink() }

// move gives the focus to the field d away, wrapping around.
func (m model) move(d int) (model, tea.Cmd) {
	m.fields[m.focus] = m.fields[m.focus].Blur()
	m.focus = (m.focus + d + len(m.fields)) % len(m.fields)
	var cmd tea.Cmd
	m.fields[m.focus], cmd = m.fields[m.focus].Focus()
	return m, cmd
}

func (m model) submit() (model, tea.Cmd) {
	var cmd tea.Cmd
	if err := m.fields[0].Err(); err != nil {
		m.sh, cmd = m.sh.SetStatus(err.Error(), shell.Error, 3*time.Second)
		return m, cmd
	}
	text := fmt.Sprintf("%s runs %q (password: %d characters)",
		m.fields[0].Value(), m.fields[1].Value(), len([]rune(m.fields[2].Value())))
	m.sh, cmd = m.sh.SetStatus(text, shell.Info, 4*time.Second)
	return m, cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			return m.move(-1)
		case "down":
			return m.move(1)
		case "enter":
			return m.submit()
		}
	}
	// A blink is only the focused field's, but each checks its own id.
	cmds := make([]tea.Cmd, len(m.fields))
	for i := range m.fields {
		var r bezel.Result
		m.fields[i], r = m.fields[i].Update(msg)
		cmds[i] = r.Cmd
	}
	return m, tea.Batch(cmds...)
}

func (m model) body(r layout.Rect) string {
	w := max(r.W-labelWidth-2, 4)
	var out []string
	for i, f := range m.fields {
		label := m.th.Dim.Render(fmt.Sprintf("%-*s", labelWidth, labels[i]))
		if i == m.focus {
			label = m.th.Accent.Render(fmt.Sprintf("%-*s", labelWidth, labels[i]))
		}
		fw := w
		if i == 1 {
			fw -= len(prompt)
		}
		out = append(out, " "+label+f.SetWidth(fw).View())
		if err := f.Err(); err != nil && i != m.focus {
			out = append(out, " "+strings.Repeat(" ", labelWidth)+m.th.Danger.Render(err.Error()))
		}
		out = append(out, "")
	}
	out = append(out, " "+m.th.Muted.Render("↑/↓ move · tab completes · enter submits · esc quits"))
	return strings.Join(out, "\n")
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"input": {Title: "deploy", Body: m.body},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
