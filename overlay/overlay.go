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

// CloseMsg pops the top overlay; PushMsg opens one above it.
type CloseMsg struct{}
type PushMsg struct{ Overlay Overlay }

// Close is the command an overlay returns to dismiss itself.
func Close() tea.Cmd { return func() tea.Msg { return CloseMsg{} } }

// Open is the command that pushes o on top of the stack.
func Open(o Overlay) tea.Cmd { return func() tea.Msg { return PushMsg{Overlay: o} } }

// Stack is a value: every method returns the new stack.
type Stack struct{ items []Overlay }

func (s Stack) Push(o Overlay) Stack {
	items := make([]Overlay, len(s.items)+1)
	copy(items, s.items)
	items[len(s.items)] = o
	return Stack{items: items}
}

func (s Stack) Pop() Stack {
	if len(s.items) == 0 {
		return s
	}
	n := len(s.items) - 1
	return Stack{items: s.items[:n:n]}
}

func (s Stack) Top() Overlay {
	if len(s.items) == 0 {
		return nil
	}
	return s.items[len(s.items)-1]
}

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
		return s, false, nil
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
	return Stack{items: items}
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
