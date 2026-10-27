// Package accordion is a column of sections, each a heading that opens and
// closes the body under it, sliding rather than jumping.
package accordion

import (
	"math"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Section is one heading and the text it opens onto.
type Section struct {
	Title, Body string
}

// Keys are the bindings the accordion answers to.
type Keys struct {
	Up, Down, Toggle key.Binding
}

// DefaultKeys are the arrows, and enter or space to open and close.
func DefaultKeys() Keys {
	return Keys{
		Up:     key.NewBinding(key.WithKeys("up"), key.WithHelp("↑/↓", "move")),
		Down:   key.NewBinding(key.WithKeys("down")),
		Toggle: key.NewBinding(key.WithKeys("enter", "space"), key.WithHelp("enter", "open/close")),
	}
}

// fold is how open one section is, and the run moving it.
type fold struct {
	from, to float64 // 0 closed, 1 open
	p        float64
	run      frames.Run
}

// share is how open f is now.
func (f fold) share() float64 {
	if !f.run.Active() {
		return f.to
	}
	return f.from + (f.to-f.from)*frames.Or(f.run.Ease)(f.p)
}

// Model is a value: every method returns the new accordion.
type Model struct {
	sections []Section
	folds    []fold
	cursor   int
	dur      time.Duration
	keys     Keys
}

// New is sections, every one closed, the cursor on the first; each opens and
// closes over 200ms.
func New(sections []Section) Model {
	return Model{sections: slices.Clone(sections), folds: make([]fold, len(sections)), dur: 200 * time.Millisecond, keys: DefaultKeys()}
}

// WithDuration opens and closes each section over d; 0 snaps.
func (m Model) WithDuration(d time.Duration) Model {
	m.dur = max(d, 0)
	return m
}

// WithKeys rebinds the accordion.
func (m Model) WithKeys(k Keys) Model {
	m.keys = k
	return m
}

// Cursor is the section the cursor is on.
func (m Model) Cursor() int { return m.cursor }

// IsOpen reports whether section i is open, or opening.
func (m Model) IsOpen(i int) bool { return i >= 0 && i < len(m.folds) && m.folds[i].to == 1 }

// Toggle opens section i when it is closed and closes it when it is open.
func (m Model) Toggle(i int) (Model, tea.Cmd) {
	if i < 0 || i >= len(m.folds) {
		return m, nil
	}
	m.folds = slices.Clone(m.folds)
	f := m.folds[i]
	f.from, f.to, f.p = f.share(), 1-f.to, 0
	f.run = frames.NewRun(m.dur)
	m.folds[i] = f
	return m, f.run.Tick()
}

// Active reports whether any section is still opening or closing.
func (m Model) Active() bool {
	for _, f := range m.folds {
		if f.run.Active() {
			return true
		}
	}
	return false
}

// FrameAt is section i's frame for at, for a test that delivers frames itself.
func (m Model) FrameAt(i int, at time.Time) tea.Msg { return m.folds[i].run.FrameAt(at) }

// Update moves the cursor and toggles on keys, and slides the sections on
// their own frames.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(k, m.keys.Up):
			m.cursor = max(m.cursor-1, 0)
		case key.Matches(k, m.keys.Down):
			m.cursor = min(m.cursor+1, max(len(m.sections)-1, 0))
		case key.Matches(k, m.keys.Toggle):
			return m.Toggle(m.cursor)
		}
		return m, nil
	}
	for i, f := range m.folds {
		run, p, ok, cmd := f.run.Update(msg)
		if !ok {
			continue
		}
		m.folds = slices.Clone(m.folds)
		f.run, f.p = run, p
		m.folds[i] = f
		return m, cmd
	}
	return m, nil
}

// Legend is the keys the accordion answers to.
func (m Model) Legend() []legend.Entry {
	return legend.FromBindings([]key.Binding{m.keys.Up, m.keys.Down, m.keys.Toggle})
}

// View is every heading, and under each as much of its body as it is open,
// wrapped to width.
func (m Model) View(th theme.Resolved, width int) string {
	var out []string
	for i, s := range m.sections {
		share := m.folds[i].share()
		mark := "▸ "
		if share > 0 {
			mark = "▾ "
		}
		head := draw.Truncate(mark+s.Title, width)
		if i == m.cursor {
			out = append(out, th.Cursor.Render(head))
		} else {
			out = append(out, th.Bold.Render(head))
		}
		if share == 0 {
			continue // a closed section shows none of its body: skip wrapping it
		}
		body := draw.Wrap(s.Body, max(width-2, 1))
		shown := int(math.Round(share * float64(len(body))))
		for _, l := range body[:shown] {
			out = append(out, "  "+th.Dim.Render(l))
		}
	}
	return strings.Join(out, "\n")
}
