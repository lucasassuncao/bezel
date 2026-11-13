// Theme: every colour role and a list drawn with it; n and p step through
// a few presets.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

var presets = []string{"mint", "gruvbox", "dracula", "nord", "tokyonight", "ubuntu", "noctisazureus", "catppuccin", "rosepine", "retro", "minimal", "cyberpunk", "ocean", "terminal", "plain", "default"}

type stepMsg struct{ by int }

type model struct {
	sh     shell.Shell
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

// apply hands preset i to the shell, which resolves it light or dark.
func (m model) apply(i int) model {
	m.preset = (i + len(presets)) % len(presets)
	t, err := theme.Lookup(presets[m.preset])
	if err != nil {
		panic(err) // every name in presets is built in
	}
	m.sh = m.sh.SetBase(t).SetSubtitle(presets[m.preset])
	return m
}

func (m model) Init() tea.Cmd { return m.sh.Init() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	if step, ok := msg.(stepMsg); ok {
		return m.apply(m.preset + step.by), nil
	}
	return m, nil
}

func (m model) roles() string {
	th := m.sh.Theme()
	rows := []struct {
		name  string
		style lipgloss.Style
	}{
		{"Accent", th.Accent}, {"Success", th.Success}, {"Warning", th.Warning},
		{"Danger", th.Danger}, {"Info", th.Info}, {"Dim", th.Dim}, {"Muted", th.Muted},
	}
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(r.style.Render(fmt.Sprintf("■■ %-8s", r.name)) + "\n")
	}
	b.WriteString("\n" + th.Badge.Render(" badge ") + " " + th.Key.Render("ctrl+s") + " saves")
	return b.String()
}

func (m model) sample() string {
	th := m.sh.Theme()
	return strings.Join([]string{
		th.Section.Render("FRUIT"),
		th.Cursor.Render("> banana"),
		"  mango",
		th.Dim.Render("  kiwi (out of season)"),
		"",
		th.Success.Render("✓ saved") + "  " + th.Danger.Render("✗ 2 errors"),
	}, "\n")
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"roles":  {Title: "roles", Body: func(layout.Rect) string { return m.roles() }},
		"sample": {Title: "sample", Body: func(layout.Rect) string { return m.sample() }},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
