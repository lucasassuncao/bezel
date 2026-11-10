// Wizard: a dev container set up over three pages. The first will not move
// on without a name, the second checks tools out of a filtered list, and
// Finish reports what the pages hold.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/form"
	"github.com/lucasassuncao/bezel/component/wizard"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// The pages, by index, as the summary reads them back.
const (
	pProject = iota
	pTools
	pReview
)

type model struct {
	sh  shell.Shell
	wiz wizard.Model
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("wizard")),
		Theme:  th,
		Title:  "wizard",
		Actions: []shell.Action{
			shell.Custom("tab", "next field", nil, shell.DisplayOnly()),
			shell.Custom("pgdn", "next page", nil, shell.DisplayOnly()),
			shell.Custom("pgup", "previous page", nil, shell.DisplayOnly()),
			shell.Custom("ctrl+c", "quit", nil, shell.DisplayOnly()),
		},
	})
	wiz := wizard.New(th,
		wizard.NewPage("Project", form.New(th,
			form.NewInput("Name").WithPlaceholder("my-api").WithWidth(28),
			form.NewSelect("Image", []string{"ubuntu", "debian", "alpine"}),
		)).WithCheck(named),
		wizard.NewPage("Tools", form.New(th,
			form.NewMultiSelect("Tools", []string{"git", "docker", "node", "golang", "python", "rust", "deno", "dotnet"}).WithHeight(5),
		)).WithCheck(someTool),
		wizard.NewPage("Review", form.New(th,
			form.NewRadio("Shell", []string{"bash", "zsh", "fish"}),
			form.NewCheckbox("Start it now"),
		)),
	)
	return model{sh: sh, wiz: wiz}
}

// named keeps the first page until the container has a name.
func named(f form.Form) error {
	if in, _ := form.Get[form.Input](f, 0); strings.TrimSpace(in.Value()) == "" {
		return errors.New("the container needs a name")
	}
	return nil
}

// someTool keeps the second page until a tool is checked.
func someTool(f form.Form) error {
	if ms, _ := form.Get[form.MultiSelect](f, 0); len(ms.Values()) == 0 {
		return errors.New("check at least one tool")
	}
	return nil
}

// Init starts the caret blinking in the name the wizard opens on.
func (m model) Init() tea.Cmd { return m.wiz.Init() }

// summary is what Finish reports: every value the pages hold.
func (m model) summary() string {
	name, _ := form.Get[form.Input](m.wiz.Page(pProject).Form(), 0)
	image, _ := form.Get[form.Select](m.wiz.Page(pProject).Form(), 1)
	tools, _ := form.Get[form.MultiSelect](m.wiz.Page(pTools).Form(), 0)
	sh, _ := form.Get[form.Radio](m.wiz.Page(pReview).Form(), 0)
	now, _ := form.Get[form.Checkbox](m.wiz.Page(pReview).Form(), 1)
	when := "later"
	if now.Checked() {
		when = "now"
	}
	return fmt.Sprintf("%s: %s with %s, %s, starts %s", name.Value(), image.Value(),
		strings.Join(tools.Values(), "+"), sh.Value(), when)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		if k.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Keys belong to the wizard: the shell would read "q" or "?" typed into a field.
		return m.wizard(msg)
	}
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	return m.wizard(msg)
}

// wizard hands msg to the wizard and reports the summary when it is done.
func (m model) wizard(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	m.wiz, r = m.wiz.Update(msg)
	if _, done := r.Event.(wizard.Done); done {
		var cmd tea.Cmd
		m.sh, cmd = m.sh.SetStatus(m.summary(), shell.OK, 6*time.Second)
		return m, tea.Batch(r.Cmd, cmd)
	}
	return m, r.Cmd
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"wizard": {Title: "new dev container", Body: func(r layout.Rect) string { return m.wiz.SetWidth(r.W).View() }},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
