package overlay

import (
	"fmt"
	"slices"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// unlockHint is the problem an unmatched WithRequire reports.
const unlockHint = "type the exact text to unlock"

// Prompt asks for one line of text. Enter accepts once the text passes
// WithRequire and WithValidate, running onAccept and closing; esc is the
// stack's. A host that owns its keys can drive it with Value and Accepts.
type Prompt struct {
	look
	title    string
	lines    []string
	input    textinput.Model
	label    string
	preview  func(string) string
	previewL string
	require  string
	validate func(string) string
	accept   string
	onAccept func(value string) tea.Cmd
}

// NewPrompt builds a prompt with a "> " input; onAccept may be nil.
func NewPrompt(title string, onAccept func(value string) tea.Cmd, th theme.Resolved) Prompt {
	in := draw.WithCaret(textinput.New(), th.CaretBlink)
	in.Prompt = "> "
	in.Focus()
	return Prompt{title: title, input: in, accept: "confirm", onAccept: onAccept, look: look{th: th}}
}

// WithInput replaces the input, for a caller that configures its own prompt,
// placeholder, width or caret.
func (p Prompt) WithInput(in textinput.Model) Prompt {
	in.Focus()
	p.input = in
	return p
}

// WithLines is the explanation drawn with the input.
func (p Prompt) WithLines(lines ...string) Prompt {
	p.lines = slices.Clone(lines)
	return p
}

// WithLabel draws the input as a labelled row leading the prompt; without a
// label the input follows the lines, the way a typed confirmation reads.
func (p Prompt) WithLabel(label string) Prompt {
	p.label = label
	return p
}

// WithPreview draws a labelled row under the input computed from the text,
// such as the path a name will create. It needs WithLabel to be drawn.
func (p Prompt) WithPreview(label string, preview func(typed string) string) Prompt {
	p.previewL, p.preview = label, preview
	return p
}

// WithRequire accepts only want, typed exactly: the friction for an action
// that cannot be undone.
func (p Prompt) WithRequire(want string) Prompt {
	p.require = want
	if p.input.Placeholder == "" {
		p.input.Placeholder = want
	}
	return p
}

// WithValidate refuses text for which check returns a problem, shown in place
// of the hint.
func (p Prompt) WithValidate(check func(typed string) string) Prompt {
	p.validate = check
	return p
}

// WithAcceptHint is what the hint says enter does; "confirm" by default.
func (p Prompt) WithAcceptHint(desc string) Prompt {
	p.accept = desc
	return p
}

// Value is the text typed so far.
func (p Prompt) Value() string { return p.input.Value() }

// Problem is why the text cannot be accepted yet, or "".
func (p Prompt) Problem() string {
	if p.require != "" && p.Value() != p.require {
		return unlockHint
	}
	if p.validate != nil {
		return p.validate(p.Value())
	}
	return ""
}

// Accepts reports whether enter would accept the text now.
func (p Prompt) Accepts() bool { return p.Problem() == "" }

func (p Prompt) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok && k.Code == tea.KeyEnter {
		switch {
		case !p.Accepts():
			return p, nil
		case p.onAccept == nil:
			return p, Close()
		}
		return p, tea.Batch(p.onAccept(p.Value()), Close())
	}
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	return p, cmd
}

// Legend offers enter only when it would do something.
func (p Prompt) Legend() []legend.Entry {
	esc := legend.New("esc", "cancel")
	if !p.Accepts() {
		return []legend.Entry{esc}
	}
	return []legend.Entry{legend.New("enter", p.accept), esc}
}

func (p Prompt) View(body layout.Rect) string {
	lines := []string{accent(p.style()).Bold(true).Render(p.title), ""}
	if p.label != "" {
		w := max(draw.Width(p.label), draw.Width(p.previewL))
		lines = append(lines, p.row(p.label, w, p.input.View()))
		if p.preview != nil {
			lines = append(lines, p.row(p.previewL, w, p.preview(p.Value())))
		}
		lines = append(lines, "")
	}
	lines = append(lines, p.lines...)
	if p.label == "" {
		lines = append(lines, "", p.input.View())
	}
	return Box(append(lines, "", p.hintLine()), body, p.style())
}

// row is one labelled line, the labels padded to one column.
func (p Prompt) row(label string, width int, content string) string {
	return p.hint().Key.Render(fmt.Sprintf("  %-*s  ", width, label)) + content
}

// hintLine says what the keys do, or what stands in the way of enter.
func (p Prompt) hintLine() string {
	if problem := p.Problem(); problem != "" {
		return p.hint().Text.Render(problem) + "   " + legend.HintLine(p.Legend(), p.hint())
	}
	return legend.HintLine(p.Legend(), p.hint())
}
