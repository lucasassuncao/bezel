// Form: one of each control, filled in and submitted. tab and shift+tab move
// between them; Create reports what the form holds on the status row.
package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/form"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// The fields, by index, as the summary reads them back.
const (
	fName = iota
	fImage
	fShell
	fDocker
	fPort
)

type createMsg struct{}

type model struct {
	sh   shell.Shell
	th   theme.Resolved
	form form.Form
	init tea.Cmd // the input's caret blink, started once the program runs
}

func newModel() model {
	th := theme.Resolve(theme.ThemeMint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("form")),
		Theme:  th,
		Title:  "form",
		Actions: []shell.Action{
			shell.Custom("tab", "next field", nil, shell.DisplayOnly()),
			shell.Custom("shift+tab", "previous", nil, shell.DisplayOnly()),
			shell.Custom("ctrl+c", "quit", nil, shell.DisplayOnly()),
		},
	})
	f, cmd := form.New(
		form.NewInput("Name", "my-project", 28, th),
		form.NewSelect("Image", []string{"ubuntu", "debian", "alpine"}, 0, th),
		form.NewRadio("Shell", []string{"bash", "zsh", "fish"}, 0, th),
		form.NewToggle("Docker", false, th),
		form.NewCheckbox("Forward port 3000", false, th),
		form.NewButton("Create", func() tea.Msg { return createMsg{} }, th),
	)
	return model{sh: sh, th: th, form: f, init: cmd}
}

func (m model) Init() tea.Cmd { return m.init }

// summary is what Create reports: every value the form holds.
func (m model) summary() string {
	name := m.form.Field(fName).(form.Input).Value()
	if name == "" {
		name = "my-project"
	}
	return fmt.Sprintf("created %s: %s, %s, docker %v, port 3000 %v", name,
		m.form.Field(fImage).(form.Select).Value(),
		m.form.Field(fShell).(form.Radio).Value(),
		m.form.Field(fDocker).(form.Toggle).On(),
		m.form.Field(fPort).(form.Checkbox).Checked())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case createMsg:
		var cmd tea.Cmd
		m.sh, cmd = m.sh.SetStatus(m.summary(), shell.OK, 6*time.Second)
		return m, cmd
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Keys belong to the form: the shell would read "q" or "?" typed into a field.
		var cmd tea.Cmd
		m.form, cmd = m.form.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.sh, _, cmd = m.sh.Update(msg)
	if cmd == nil {
		m.form, cmd = m.form.Update(msg)
	}
	return m, cmd
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"form": {Title: "new dev container", Body: func(r layout.Rect) string { return m.form.View(r.W) }},
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
