// Viewport: an access log, numbered in a gutter, its ERROR lines in red and
// every ERROR highlighted. The arrows, pgup/pgdown and home/end scroll, ← and
// → pan, n and N go from error to error, w wraps, and the wheel scrolls.
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/viewport"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

const lines = 300

type (
	nextMsg struct{}
	prevMsg struct{}
	wrapMsg struct{}
)

// failed reports whether log line i is an error, one line in 37.
func failed(i int) bool { return i%37 == 36 }

// accessLog is the generated log, a line about 100 cells wide.
func accessLog() string {
	out := make([]string, lines)
	for i := range out {
		level, status := "", 200
		if failed(i) {
			level, status = "ERROR ", 500
		}
		out[i] = fmt.Sprintf("12:%02d:%02d %sGET /api/items?page=%d&size=50&sort=name&filter=active status=%d bytes=%d",
			i/60, i%60, level, i+1, status, 1000+i*37)
	}
	return strings.Join(out, "\n")
}

// gutter numbers each line, leaves a wrapped line's other rows blank and
// marks the rows past the end.
func gutter(r viewport.GutterRow) string {
	switch {
	case r.Line >= r.Lines:
		return "   ~ │ "
	case r.Continued:
		return "     │ "
	}
	return fmt.Sprintf("%4d │ ", r.Line+1)
}

type model struct {
	sh shell.Shell
	th theme.Resolved
	vp viewport.Model
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("log")),
		Theme:  th,
		Title:  "viewport",
		Actions: []shell.Action{
			shell.Custom("↑/↓", "scroll", nil, shell.DisplayOnly()),
			shell.Custom("←/→", "pan", nil, shell.DisplayOnly()),
			shell.Custom("pgup/pgdn", "page", nil, shell.DisplayOnly()),
			shell.Custom("n", "next error", shell.Send(nextMsg{})),
			shell.Custom("N", "previous error", shell.Send(prevMsg{})),
			shell.Custom("w", "wrap", shell.Send(wrapMsg{})),
			shell.Quit(),
		},
	})
	content := accessLog()
	vp := viewport.New().
		WithGutter(gutter).
		WithFill(true).
		WithLineStyle(func(i int) lipgloss.Style {
			if failed(i) {
				return th.Danger
			}
			return lipgloss.NewStyle()
		}).
		SetContent(content).
		SetHighlights(regexp.MustCompile(`ERROR`).FindAllStringIndex(content, -1)).
		ToTop()
	return model{sh: sh, th: th, vp: vp}
}

func (m model) Init() tea.Cmd { return nil }

// sized fits the viewport to the pane's inside, as the shell last laid it out.
func (m model) sized() model {
	r := draw.InnerRect(m.sh.Rect("log"))
	m.vp = m.vp.SetWidth(r.W).SetHeight(r.H)
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m.sized(), r.Cmd
	}
	m = m.sized()
	switch msg := msg.(type) {
	case nextMsg:
		m.vp = m.vp.NextHighlight()
	case prevMsg:
		m.vp = m.vp.PrevHighlight()
	case wrapMsg:
		m.vp = m.vp.SetWrap(!m.vp.Wrap())
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m.vp, _ = m.vp.Update(msg)
	case tea.MouseWheelMsg:
		m.vp, _ = m.vp.Update(msg)
	}
	return m, nil
}

func (m model) View() tea.View {
	vp := m.sized().vp
	v := m.sh.View(map[string]shell.Pane{
		"log": {
			Title:  "access.log",
			Bottom: fmt.Sprintf("%3.0f%%", vp.ScrollPercent()*100),
			Scroll: vp.Scroll(),
			Body:   func(layout.Rect) string { return vp.View(m.th) },
		},
	})
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
