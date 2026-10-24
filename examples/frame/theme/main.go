// Theme: every colour role and a list drawn with it; n and p step through
// a few presets.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

var presets = []string{"mint", "banana", "strawberry", "blueberry", "grape", "plain"}

type stepMsg struct{ by int }

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	preset int
}

func newModel() model {
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fixed("roles", layout.Ratio(1, 2)), layout.Fill("sample")),
		Title:  "theme",
		Actions: []shell.Action{
			shell.Custom("n", "next theme", shell.Send(stepMsg{by: 1})),
			shell.Custom("p", "previous theme", shell.Send(stepMsg{by: -1})),
			shell.Quit(),
		},
	})
	return model{sh: sh}.apply(0)
}

// apply resolves preset i and hands it to the shell and to the panes.
func (m model) apply(i int) model {
	m.preset = (i + len(presets)) % len(presets)
	t, err := theme.Lookup(presets[m.preset])
	if err != nil {
		panic(err) // every name in presets is built in
	}
	m.th = theme.Resolve(t, true)
	m.sh = m.sh.SetTheme(m.th).SetSubtitle(presets[m.preset])
	return m
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case stepMsg:
		return m.apply(m.preset + msg.by), nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) roles() string {
	rows := []struct {
		name  string
		style lipgloss.Style
	}{
		{"Accent", m.th.Accent}, {"Success", m.th.Success}, {"Warning", m.th.Warning},
		{"Danger", m.th.Danger}, {"Info", m.th.Info}, {"Dim", m.th.Dim}, {"Muted", m.th.Muted},
	}
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(r.style.Render(fmt.Sprintf("■■ %-8s", r.name)) + "\n")
	}
	b.WriteString("\n" + m.th.Badge.Render(" badge ") + " " + m.th.Key.Render("ctrl+s") + " saves")
	return b.String()
}

func (m model) sample() string {
	return strings.Join([]string{
		m.th.Section.Render("FRUIT"),
		m.th.Cursor.Render("> banana"),
		"  mango",
		m.th.Dim.Render("  kiwi (out of season)"),
		"",
		m.th.Success.Render("✓ saved") + "  " + m.th.Danger.Render("✗ 2 errors"),
	}, "\n")
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"roles":  {Title: "roles", Body: func(layout.Rect) string { return m.roles() }},
		"sample": {Title: "sample", Body: func(layout.Rect) string { return m.sample() }},
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
