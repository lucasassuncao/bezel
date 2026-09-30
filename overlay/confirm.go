package overlay

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
)

// Confirm asks yes/no with two buttons, Yes focused. ←/→ or tab move the
// focus, enter or space picks it, y and n answer at once, esc or q cancel.
type Confirm struct {
	title   string
	message string
	onYes   tea.Cmd
	style   lipgloss.Style
	hint    legend.Style
	no      bool // focus is on No
}

func NewConfirm(title, message string, onYes tea.Cmd, style lipgloss.Style, hint legend.Style) Confirm {
	return Confirm{title: title, message: message, onYes: onYes, style: style, hint: hint}
}

func (c Confirm) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return c, nil
	}
	switch k.String() {
	case "left", "right", "tab", "shift+tab":
		c.no = !c.no
		return c, nil
	case "enter", "space", " ":
		if c.no {
			return c, Close()
		}
		return c, tea.Batch(c.onYes, Close())
	case "y", "Y":
		return c, tea.Batch(c.onYes, Close())
	case "n", "N", "esc", "q":
		return c, Close()
	}
	return c, nil
}

func (c Confirm) Legend() []legend.Entry {
	return []legend.Entry{legend.New("←/→", "choose"), legend.New("enter", "pick"), legend.New("y/n", "answer"), legend.New("esc", "cancel")}
}

func (c Confirm) View(body layout.Rect) string {
	ink := accent(c.style)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, button("Yes", ink, !c.no), "  ", button("No", ink, c.no))
	return dialog(c.title, []string{c.message}, buttons, c.style, body)
}
