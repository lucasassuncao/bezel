// Browser: recipe names on the left, the selected recipe on the right; tab
// moves the keys to the recipe so it scrolls.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/browser"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// recipe builds an item whose detail is numbered steps, long enough to scroll.
func recipe(name string, steps ...string) browser.Item {
	return browser.Item{Label: name, Detail: func() string {
		var b strings.Builder
		b.WriteString(strings.ToUpper(name) + "\n\n")
		for i, s := range steps {
			fmt.Fprintf(&b, "%d. %s\n\n", i+1, s)
		}
		return b.String()
	}}
}

var recipes = []browser.Item{
	recipe("pancakes", "Whisk flour, sugar, baking powder and salt.", "Beat in milk, egg and melted butter.",
		"Rest the batter 10 minutes.", "Heat a buttered pan.", "Pour a ladle, flip when bubbles form.",
		"Stack and serve warm.", "Top with fruit.", "Or with syrup.", "Or both."),
	recipe("guacamole", "Halve and stone three avocados.", "Mash with lime juice and salt.",
		"Fold in onion, tomato and coriander.", "Taste, add chilli."),
	recipe("omelette", "Beat three eggs with salt.", "Melt butter in a small pan.",
		"Pour, stir until just set.", "Fold and slide onto a plate."),
	recipe("lemonade", "Squeeze four lemons.", "Stir in sugar until it dissolves.",
		"Top with cold water and ice."),
}

type model struct {
	sh      shell.Shell
	th      theme.Resolved
	browser browser.Model
}

func newModel() model {
	th := theme.Resolve(theme.ThemeMint, true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fixed("list", layout.Ratio(1, 3), layout.Min(20)), layout.Fill("preview")),
		Theme:  th,
		Title:  "browser",
		Actions: []shell.Action{
			shell.Move(),
			shell.Custom("tab", "list/recipe", nil, shell.DisplayOnly()),
			shell.Custom("enter", "cook", nil, shell.DisplayOnly()),
			shell.Quit(),
		},
	})
	return model{sh: sh, th: th, browser: browser.New(recipes, "")}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	if key.String() == "ctrl+c" {
		return m, tea.Quit
	}
	var act browser.Action
	m.browser = m.browser.SetPreviewHeight(draw.InnerRect(m.sh.Rect("preview")).H)
	m.browser, act = m.browser.Update(key)
	// The shell's focus is only the border; the browser owns which side moves.
	if m.browser.PreviewFocus {
		m.sh = m.sh.SetFocus("preview")
	} else {
		m.sh = m.sh.SetFocus("list")
	}
	switch act {
	case browser.Chosen:
		m.sh, cmd = m.sh.SetStatus("cooking "+m.browser.Selected().Label, shell.OK, 3*time.Second)
		return m, cmd
	case browser.Dismissed:
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"list":    {Title: "recipes", Body: func(r layout.Rect) string { return m.browser.ListView(m.th, r.H) }},
		"preview": {Title: "recipe", Body: func(r layout.Rect) string { return m.browser.PreviewView(r.H) }},
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
