package overlay

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Confirm asks yes/no with two buttons, Yes focused. ←/→ or tab move the
// focus, enter or space picks it, y and n answer at once, esc or q cancel.
type Confirm struct {
	look
	title   string
	message string
	onYes   tea.Cmd
	no      bool // focus is on No
	keys    ConfirmKeys
}

// ConfirmKeys are the bindings a Confirm answers to besides esc. No has no
// help of its own: Yes's "y/n" names both.
type ConfirmKeys struct {
	Toggle, Accept, Yes, No key.Binding
}

// DefaultConfirmKeys are ←/→ or tab to move, enter or space to pick, and y
// or n to answer at once; q answers no.
func DefaultConfirmKeys() ConfirmKeys {
	return ConfirmKeys{
		Toggle: key.NewBinding(key.WithKeys("left", "right", "tab", "shift+tab"), key.WithHelp("←/→", "choose")),
		Accept: key.NewBinding(key.WithKeys("enter", "space", " "), key.WithHelp("enter", "pick")),
		Yes:    key.NewBinding(key.WithKeys("y", "Y"), key.WithHelp("y/n", "answer")),
		No:     key.NewBinding(key.WithKeys("n", "N", "esc", "q")),
	}
}

// NewConfirm asks message under title; onYes runs on a yes and may be nil.
func NewConfirm(title, message string, onYes tea.Cmd, th theme.Resolved) Confirm {
	return Confirm{title: title, message: message, onYes: onYes, look: look{th: th}, keys: DefaultConfirmKeys()}
}

// WithKeys rebinds the confirm; its legend follows the bindings' help.
func (c Confirm) WithKeys(k ConfirmKeys) Confirm {
	c.keys = k
	return c
}

func (c Confirm) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return c, nil
	}
	switch {
	case key.Matches(k, c.keys.Toggle):
		c.no = !c.no
		return c, nil
	case key.Matches(k, c.keys.Accept):
		if c.no {
			return c, Close()
		}
		return c, tea.Batch(c.onYes, Close())
	case key.Matches(k, c.keys.Yes):
		return c, tea.Batch(c.onYes, Close())
	case key.Matches(k, c.keys.No):
		return c, Close()
	}
	return c, nil
}

func (c Confirm) Legend() []legend.Entry {
	keys := legend.FromBindings([]key.Binding{c.keys.Toggle, c.keys.Accept, c.keys.Yes, c.keys.No})
	return append(keys, legend.New("esc", "cancel"))
}

func (c Confirm) View(body layout.Rect) string {
	ink := accent(c.style())
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, button("Yes", ink, !c.no), "  ", button("No", ink, c.no))
	return dialog(c.title, []string{c.message}, buttons, c.style(), body)
}
