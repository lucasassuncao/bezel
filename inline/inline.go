// Package inline is for CLIs that print logs and now and then ask one
// question or show one bar: nothing here takes over the screen.
package inline

import (
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/theme"
)

// ErrAborted is what Confirm returns when the user pressed esc or ctrl+c.
var ErrAborted = errors.New("inline: aborted")

// Confirm asks a yes/no question below the current line. The default is No.
func Confirm(title string, th theme.Resolved) (bool, error) {
	final, err := tea.NewProgram(newConfirm(title, th)).Run()
	if err != nil {
		return false, err
	}
	c, ok := final.(confirm)
	if !ok {
		return false, fmt.Errorf("inline: unexpected final model %T", final)
	}
	if c.aborted {
		return false, ErrAborted
	}
	return c.yes, nil
}

type confirm struct {
	title   string
	th      theme.Resolved
	yes     bool
	done    bool
	aborted bool
}

func newConfirm(title string, th theme.Resolved) confirm { return confirm{title: title, th: th} }

func (c confirm) Init() tea.Cmd { return nil }

func (c confirm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return c, nil
	}
	switch k.String() {
	case "left", "right", "tab", "shift+tab":
		c.yes = !c.yes
		return c, nil
	case "y", "Y":
		c.yes = true
	case "n", "N":
		c.yes = false
	case "enter":
	case "esc", "ctrl+c":
		c.aborted = true
	default:
		return c, nil
	}
	c.done = true
	return c, tea.Quit
}

func (c confirm) View() tea.View {
	title := c.th.Bold.Render(c.title)
	if c.done {
		answer := "No"
		if c.yes {
			answer = "Yes"
		}
		if c.aborted {
			answer = "cancelled"
		}
		return tea.NewView(title + " " + c.th.Dim.Render(answer) + "\n")
	}
	button := func(label string, on bool) string {
		if on {
			return c.th.Badge.Padding(0, 2).Render(label)
		}
		return c.th.Dim.Padding(0, 2).Render(label)
	}
	hint := c.th.Muted.Render("←/→ choose · y/n answer · enter confirm · esc cancel")
	return tea.NewView(title + "\n\n  " + button("Yes", c.yes) + " " + button("No", !c.yes) + "\n\n" + hint + "\n")
}

// Bar is a done/total progress bar width cells wide, for a caller that
// redraws it in place with \r.
func Bar(done, total, width int, th theme.Resolved) string {
	width = max(width, 1)
	filled := 0
	if total > 0 {
		filled = min(max(done, 0)*width/total, width)
	}
	return th.Accent.Render(strings.Repeat("█", filled)) + th.Muted.Render(strings.Repeat("░", width-filled))
}
