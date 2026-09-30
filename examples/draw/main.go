// Draw: one scrolling pane with a section per helper, each drawn at the
// pane's own width.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

const paragraph = "Wrap folds plain text into rows of at most the given width, breaking after spaces and hyphens, and cuts a word wider than a row."

const source = "package main\n\nfunc main() {\n    println(\"hi\")\n}"

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	offset int
}

func newModel() model {
	th := theme.Resolve(theme.ThemeMint, true)
	sh := shell.New(shell.Config{
		Layout:  layout.Columns(layout.Fill("draw")),
		Theme:   th,
		Title:   "draw",
		Actions: []shell.Action{shell.Scroll(), shell.Quit()},
	})
	return model{sh: sh, th: th}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			m.offset = max(0, m.offset-1)
		case "down":
			r := draw.InnerRect(m.sh.Rect("draw"))
			m.offset = min(m.offset+1, max(0, len(m.lines(r.W))-r.H))
		}
	}
	return m, nil
}

// sections draws every helper at width w, one block per helper.
func (m model) sections(w int) []string {
	var out []string
	add := func(name string, lines ...string) {
		out = append(out, draw.Heading(name, m.th.Section, w))
		out = append(out, lines...)
		out = append(out, "")
	}
	add("Header", draw.Header(w, "my app", "kv/app/prod", "v1.2.0", m.th.Chrome))
	add("Tabs", draw.Tabs(w, []string{"secrets", "policies", "auth"}, 1, m.th.Chrome))
	add("Breadcrumb", draw.Breadcrumb([]string{"kv", "app", "prod", "db"}))
	// KV ends its row with a newline, so rows stack by concatenation.
	kv := draw.KV("address", "http://127.0.0.1:8200", m.th.Dim, w) +
		draw.KV("version", "4 (2 deleted)", m.th.Dim, w)
	add("KV", strings.TrimSuffix(kv, "\n"))
	add("Wrap", draw.Wrap(paragraph, min(w, 40))...)
	add("Truncate", draw.Truncate("a line much longer than the room it has to fit in", 24))
	add("NumberLines", draw.NumberLines(source, m.th.Muted))
	add("EmptyState", draw.EmptyState(layout.Rect{W: w, H: 4}, "nothing here", "press n to add one", m.th.Chrome))
	return out
}

// lines is the whole content at width w, one entry per screen row.
func (m model) lines(w int) []string {
	return strings.Split(strings.Join(m.sections(w), "\n"), "\n")
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"draw": {Title: "draw", Body: func(r layout.Rect) string {
			lines := m.lines(r.W)
			return strings.Join(lines[min(m.offset, len(lines)-1):], "\n")
		}},
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
