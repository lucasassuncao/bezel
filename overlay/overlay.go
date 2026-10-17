// Package overlay is the stack of modals floating over the panes. The top one
// gets every key; Esc pops it unless it says it owns Esc; the rest of the app
// sees nothing while the stack is non-empty.
package overlay

import (
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
)

// Overlay is one floating modal. View gets the body rect and picks its own
// box within it.
type Overlay interface {
	Update(msg tea.Msg) (Overlay, tea.Cmd)
	View(body layout.Rect) string
	Legend() []legend.Entry
}

// EscOwner is implemented by overlays with their own Esc semantics, such as
// a nested editor that steps out one level at a time.
type EscOwner interface{ OwnsEsc() bool }

// BottomAnchored is implemented by overlays drawn flush with the bottom of
// the body instead of centred: a command line and its candidates.
type BottomAnchored interface{ AnchorBottom() bool }

// StatusLiner is implemented by overlays that draw the status row while on
// top: a command line typed where the status usually is.
type StatusLiner interface{ StatusLine(width int) string }

// Resizer is implemented by overlays that size what they show to the body,
// such as how far a page scrolls. The stack calls Resize when it pushes one
// and whenever its body changes, so Update knows before anything is drawn.
type Resizer interface {
	Resize(body layout.Rect) Overlay
}

// CloseMsg pops the top overlay.
type CloseMsg struct{}

// PushMsg opens Overlay above the top one.
type PushMsg struct{ Overlay Overlay }

// Close is the command an overlay returns to dismiss itself.
func Close() tea.Cmd { return func() tea.Msg { return CloseMsg{} } }

// Open is the command that pushes o on top of the stack.
func Open(o Overlay) tea.Cmd { return func() tea.Msg { return PushMsg{Overlay: o} } }

// TopMsg carries an overlay's own answer back to the top overlay. Other
// non-key messages pass an open modal by: they are the app's.
type TopMsg struct{ Msg tea.Msg }

// ToTop wraps cmd so its message returns to the top overlay as a TopMsg.
func ToTop(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		if msg := cmd(); msg != nil {
			return TopMsg{Msg: msg}
		}
		return nil
	}
}

// Stack is a value: every method returns the new stack.
type Stack struct {
	items []Overlay
	body  layout.Rect // what Resize last gave, zero before
}

// Push is the stack with o on top, sized to the stack's body when it has one.
func (s Stack) Push(o Overlay) Stack {
	items := make([]Overlay, len(s.items)+1)
	copy(items, s.items)
	items[len(s.items)] = s.fit(o)
	return Stack{items: items, body: s.body}
}

// Pop is the stack without its top overlay; an empty stack stays empty.
func (s Stack) Pop() Stack {
	if len(s.items) == 0 {
		return s
	}
	n := len(s.items) - 1
	return Stack{items: s.items[:n:n], body: s.body}
}

// Resize hands body to every Resizer layer, the hidden ones too. A shell does
// it on each layout; an app driving a Stack itself calls it on every resize.
func (s Stack) Resize(body layout.Rect) Stack {
	if body == s.body {
		return s
	}
	s.body = body
	items := make([]Overlay, len(s.items))
	for i, o := range s.items {
		items[i] = s.fit(o)
	}
	s.items = items
	return s
}

// fit is o sized to the body, or o as it is without a body or a Resize.
func (s Stack) fit(o Overlay) Overlay {
	if r, ok := o.(Resizer); ok && s.body.W > 0 && s.body.H > 0 {
		return r.Resize(s.body)
	}
	return o
}

// Top is the overlay that gets the keys, or nil when the stack is empty.
func (s Stack) Top() Overlay {
	if len(s.items) == 0 {
		return nil
	}
	return s.items[len(s.items)-1]
}

// Len is how many overlays are open.
func (s Stack) Len() int { return len(s.items) }

// Update routes msg to the top overlay. handled is false only when the stack
// is empty and msg is not a stack message.
func (s Stack) Update(msg tea.Msg) (Stack, bool, tea.Cmd) {
	switch m := msg.(type) {
	case CloseMsg:
		return s.Pop(), true, nil
	case PushMsg:
		return s.Push(m.Overlay), true, nil
	}
	top := s.Top()
	if top == nil {
		_, late := msg.(TopMsg)
		return s, late, nil
	}
	if m, ok := msg.(TopMsg); ok {
		msg = m.Msg
	}
	// Esc closes through CloseMsg like every other way out, so the app sees the
	// dialog go whichever key closed it.
	if k, ok := msg.(tea.KeyPressMsg); ok && k.Code == tea.KeyEscape {
		if owner, ok := top.(EscOwner); !ok || !owner.OwnsEsc() {
			return s, true, Close()
		}
	}
	updated, cmd := top.Update(msg)
	return s.replaceTop(updated), true, cmd
}

func (s Stack) replaceTop(o Overlay) Stack {
	items := make([]Overlay, len(s.items))
	copy(items, s.items)
	items[len(items)-1] = o
	return Stack{items: items, body: s.body}
}

// View composites every layer, bottom first, centred over background.
func (s Stack) View(body layout.Rect, background string) string {
	out := draw.PadHeight(background, body.Y+body.H)
	for _, o := range s.items {
		fg := o.View(body)
		fgW, fgH := draw.BlockSize(fg)
		x := body.X + max(0, (body.W-fgW)/2)
		y := body.Y + max(0, (body.H-fgH)/2)
		if b, ok := o.(BottomAnchored); ok && b.AnchorBottom() {
			x, y = body.X, body.Y+max(0, body.H-fgH)
		}
		out = draw.Composite(fg, out, x, y)
	}
	return out
}
