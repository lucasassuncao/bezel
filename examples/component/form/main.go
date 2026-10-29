// Form: one of each control, filled in and submitted. tab and shift+tab move
// between them; Create reports what the form holds on the status row.
package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/component/form"
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
}

func newModel() model {
	th := theme.Resolve(theme.Mint, true)
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
	f := form.New(th,
		form.NewInput("Name").WithPlaceholder("my-project").WithWidth(28),
		form.NewSelect("Image", []string{"ubuntu", "debian", "alpine"}),
		form.NewRadio("Shell", []string{"bash", "zsh", "fish"}),
		form.NewToggle("Docker"),
		form.NewCheckbox("Forward port 3000"),
		form.NewButton("Create", func() tea.Msg { return createMsg{} }),
	)
	return model{sh: sh, th: th, form: f}
}

// Init starts the caret blinking in the input the form opens on.
func (m model) Init() tea.Cmd { return m.form.Init() }

// summary is what Create reports: every value the form holds.
func (m model) summary() string {
	// Get is false for a wrong index or type, and a zero control reads as empty.
	in, _ := form.Get[form.Input](m.form, fName)
	image, _ := form.Get[form.Select](m.form, fImage)
	sh, _ := form.Get[form.Radio](m.form, fShell)
	docker, _ := form.Get[form.Toggle](m.form, fDocker)
	port, _ := form.Get[form.Checkbox](m.form, fPort)
	name := in.Value()
	if name == "" {
		name = "my-project"
	}
	return fmt.Sprintf("created %s: %s, %s, docker %v, port 3000 %v", name,
		image.Value(), sh.Value(), docker.On(), port.Checked())
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
	v.BackgroundColor, v.ForegroundColor = m.th.Background, m.th.Foreground
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
