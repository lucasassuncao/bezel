// Overlay: the ready-made modals, one per row; enter opens the one under
// the cursor and esc closes it.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

var kinds = []string{"Alert", "Confirm", "Prompt", "Help", "Pager", "FilePicker"}

// testLog is a command's output, long enough that only a pager shows it whole.
func testLog() string {
	var b strings.Builder
	for i := 1; i <= 48; i++ {
		fmt.Fprintf(&b, "=== RUN   TestCase%02d\n--- PASS: TestCase%02d (0.0%ds)\n", i, i, i%10)
	}
	b.WriteString("PASS\nok  \texample.com/app\t1.284s")
	return b.String()
}

type (
	openMsg struct{}
	doneMsg struct{ text string }
)

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	cursor int
}

func newModel() model {
	th := theme.Resolve(theme.Mint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fill("kinds")),
		Theme:  th,
		Title:  "overlay",
		Actions: []shell.Action{
			shell.Move(),
			shell.Custom("enter", "open", shell.Send(openMsg{})),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th}
}

// done is the command a modal runs on its answer: a status for the app.
func done(text string) tea.Cmd { return func() tea.Msg { return doneMsg{text} } }

func (m model) modal() overlay.Overlay {
	switch kinds[m.cursor] {
	case "Alert":
		return overlay.NewAlert(overlay.Warning, "Disk almost full",
			"91% of /var is in use.", m.th)
	case "Confirm":
		return overlay.NewConfirm("Delete notes.txt?", "It cannot be restored.",
			done("deleted notes.txt"), m.th).WithKind(overlay.Danger)
	case "Prompt":
		return overlay.NewPrompt("New file", func(v string) tea.Cmd { return done("created " + v) }, m.th).
			WithLabel("name").
			WithValidate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("a name is required")
				}
				return nil
			})
	case "Pager":
		return overlay.NewPager("go test ./...", testLog(), m.th)
	case "FilePicker":
		return overlay.NewFilePicker("Open a file", ".", func(path string) tea.Cmd { return done("picked " + filepath.Base(path)) },
			m.th)
	default:
		return overlay.NewHelp([]overlay.HelpSection{
			{Name: "Navigation", Entries: []legend.Entry{legend.New("↑/↓", "move"), legend.New("enter", "open")}},
			{Name: "App", Entries: []legend.Entry{legend.New("q", "quit")}},
		}, m.th).WithTitle("Keys")
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	switch msg := msg.(type) {
	case openMsg:
		o := m.modal()
		m.sh = m.sh.Push(o)
		// The picker reads its directory once it is on top to receive it.
		if fp, ok := o.(overlay.FilePicker); ok {
			return m, fp.Init()
		}
	case doneMsg:
		m.sh, cmd = m.sh.SetStatus(msg.text, shell.OK, 3*time.Second)
		return m, cmd
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			m.cursor = max(0, m.cursor-1)
		case "down":
			m.cursor = min(len(kinds)-1, m.cursor+1)
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"kinds": {Title: "overlays", Body: func(layout.Rect) string {
			var b strings.Builder
			for i, k := range kinds {
				if i == m.cursor {
					b.WriteString(m.th.Cursor.Render("> "+k) + "\n")
					continue
				}
				b.WriteString("  " + k + "\n")
			}
			return b.String()
		}},
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
