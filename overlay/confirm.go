package overlay

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/key"
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
		Toggle: key.New("left", "right", "tab", "shift+tab").WithHelp("←/→", "choose"),
		Accept: key.New("enter", "space", " ").WithHelp("enter", "pick"),
		Yes:    key.New("y", "Y").WithHelp("y/n", "answer"),
		No:     key.New("n", "N", "esc", "q"),
	}
}

// NewConfirm asks message under title; onYes runs on a yes and may be nil.
func NewConfirm(title, message string, onYes tea.Cmd, th theme.Resolved) Confirm {
	return Confirm{title: title, message: message, onYes: onYes, look: lookOf(th), keys: DefaultConfirmKeys()}
}

// WithKeys rebinds the confirm; its legend follows the bindings' help.
func (c Confirm) WithKeys(k ConfirmKeys) Confirm {
	c.keys = k
	return c
}

func (c Confirm) Update(msg tea.Msg) (Overlay, bezel.Result) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return c, bezel.Result{}
	}
	switch {
	case key.Matches(k, c.keys.Toggle):
		c.no = !c.no
		return c, bezel.Result{Handled: true}
	case key.Matches(k, c.keys.Accept):
		if c.no {
			return c, bezel.Result{Handled: true, Cmd: Close()}
		}
		return c, bezel.Result{Handled: true, Cmd: tea.Batch(c.onYes, Close())}
	case key.Matches(k, c.keys.Yes):
		return c, bezel.Result{Handled: true, Cmd: tea.Batch(c.onYes, Close())}
	case key.Matches(k, c.keys.No):
		return c, bezel.Result{Handled: true, Cmd: Close()}
	}
	return c, bezel.Result{Handled: true}
}

func (c Confirm) Legend() []legend.Entry {
	keys := legend.FromBindings([]key.Binding{c.keys.Toggle, c.keys.Accept, c.keys.Yes, c.keys.No})
	return append(keys, legend.New("esc", "cancel"))
}

func (c Confirm) View(body layout.Rect) string {
	return dialog(c.title, []string{c.message}, strings.Join(c.buttons(), "  "), c.style(), body)
}

// buttons are Yes and No as View draws them.
func (c Confirm) buttons() []string {
	return []string{c.button("Yes", !c.no), c.button("No", c.no)}
}

// Click answers yes or no on a click on its button.
func (c Confirm) Click(x, y int, body layout.Rect) (Overlay, bezel.Result) {
	switch buttonAt(c.View(body), "", dialogFrame(c.style()), c.buttons(), x, y) {
	case 0:
		return c, bezel.Result{Handled: true, Cmd: tea.Batch(c.onYes, Close())}
	case 1:
		return c, bezel.Result{Handled: true, Cmd: Close()}
	}
	return c, bezel.Result{Handled: true}
}
