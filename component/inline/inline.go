// Package inline is for CLIs that print logs and now and then ask one
// question, beside progress.Bar redrawn with : nothing takes the screen.
// Theme only: the functions take a theme but keep no Styles to edit.
package inline

import (
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// ErrAborted is what Confirm returns when the user pressed esc or ctrl+c.
var ErrAborted = errors.New("inline: aborted")

// Confirm asks a yes/no question below the current line. The default is No.
// opts reach the program as given: tea.WithContext to cancel the question,
// tea.WithInput and tea.WithOutput to answer it from a test.
func Confirm(title string, th theme.Resolved, opts ...tea.ProgramOption) (bool, error) {
	final, err := tea.NewProgram(newConfirm(title, th), opts...).Run()
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
	th      *theme.Resolved // shared by every copy and never written
	yes     bool
	done    bool
	aborted bool
	keys    confirmKeys
}

// confirmKeys are what the question answers to; the hint under it is their help.
type confirmKeys struct {
	Toggle, Yes, No, Accept, Abort key.Binding
}

func defaultConfirmKeys() confirmKeys {
	return confirmKeys{
		Toggle: key.New("left", "right", "tab", "shift+tab").WithHelp("←/→", "choose"),
		Yes:    key.New("y", "Y").WithHelp("y/n", "answer"),
		No:     key.New("n", "N"),
		Accept: key.New("enter").WithHelp("enter", "confirm"),
		Abort:  key.New("esc", "ctrl+c").WithHelp("esc", "cancel"),
	}
}

func newConfirm(title string, th theme.Resolved) confirm {
	return confirm{title: title, th: &th, keys: defaultConfirmKeys()}
}

// noTheme is what a zero confirm draws with; it is never written.
var noTheme theme.Resolved

// styles is the theme, held by pointer so a copy of the confirm costs a word,
// not the 29KB a theme.Resolved takes; the zero one for a confirm never built.
func (c *confirm) styles() *theme.Resolved {
	if c.th == nil {
		return &noTheme
	}
	return c.th
}

func (c confirm) Init() tea.Cmd { return nil }

func (c confirm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return c, nil
	}
	switch {
	case key.Matches(k, c.keys.Toggle):
		c.yes = !c.yes
		return c, nil
	case key.Matches(k, c.keys.Yes):
		c.yes = true
	case key.Matches(k, c.keys.No):
		c.yes = false
	case key.Matches(k, c.keys.Accept):
	case key.Matches(k, c.keys.Abort):
		c.aborted = true
	default:
		return c, nil
	}
	c.done = true
	return c, tea.Quit
}

func (c confirm) View() tea.View {
	title := c.styles().Bold.Render(c.title)
	if c.done {
		answer := "No"
		if c.yes {
			answer = "Yes"
		}
		if c.aborted {
			answer = "cancelled"
		}
		return tea.NewView(title + " " + c.styles().Dim.Render(answer) + "\n")
	}
	button := func(label string, on bool) string {
		if on {
			return c.styles().Badge.Padding(0, 2).Render(label)
		}
		return c.styles().Dim.Padding(0, 2).Render(label)
	}
	hint := c.styles().Muted.Render(c.hint())
	return tea.NewView(title + "\n\n  " + button("Yes", c.yes) + " " + button("No", !c.yes) + "\n\n" + hint + "\n")
}

// hint is every key with help and what it does, on one line.
func (c confirm) hint() string {
	km := c.keys
	var parts []string
	for _, e := range legend.FromBindings([]key.Binding{km.Toggle, km.Yes, km.No, km.Accept, km.Abort}) {
		parts = append(parts, e.Help().Key+" "+e.Help().Desc)
	}
	return strings.Join(parts, " · ")
}
