// Package palette is the ":" command line: the typed line on the status row
// and the commands it could still become floating above it. What a command
// does is the caller's; the palette only resolves a line to one of them.
package palette

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/theme"
)

// Item is one command. Available false keeps it off the list but known, so
// typing it says "not available here" instead of "unknown command".
type Item struct {
	Name, Arg, Key, Desc string
	Available            bool
	Run                  func(arg string) tea.Cmd
}

// DoneMsg is the palette finishing. The owner pops it, then runs Cmd or
// shows Err; both empty means the line was empty.
type DoneMsg struct {
	Cmd tea.Cmd
	Err string
}

// Rows is how many candidates show at once: past this, type, do not scroll.
const Rows = 8

// Model is the open palette. It is an overlay anchored to the bottom.
type Model struct {
	items  []Item
	input  textinput.Model
	cursor int // -1 until an arrow picks, so enter obeys what was typed
	err    string
	th     theme.Resolved
}

// New opens a palette over items, sorted by name.
func New(items []Item, th theme.Resolved) Model {
	in := textinput.New()
	in.Prompt = ":"
	in.Focus()
	return Model{input: in, cursor: -1, th: th}.SetItems(items)
}

// SetItems replaces the commands, judged again by the owner, keeping the typed
// line; a pick past the new list's end is dropped.
func (m Model) SetItems(items []Item) Model {
	m.items = slices.Clone(items)
	slices.SortFunc(m.items, func(a, b Item) int { return strings.Compare(a.Name, b.Name) })
	if m.cursor >= len(m.candidates()) {
		m.cursor = -1
	}
	return m
}

func (m Model) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		// A paste: the input's own business.
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	switch k.String() {
	case "enter":
		return m.run()
	case "up", "down":
		return m.pick(k.String() == "up"), nil
	case "tab":
		return m.complete(), nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	// The list under a pick changed, and the last error was about the old line.
	m.cursor, m.err = -1, ""
	return m, cmd
}

func (Model) Legend() []legend.Entry {
	return []legend.Entry{legend.New("enter", "run"), legend.New("tab", "complete"), legend.New("↑/↓", "choose"), legend.New("esc", "cancel")}
}

func (Model) AnchorBottom() bool { return true }

// Value is the line typed so far, without the ":" prompt.
func (m Model) Value() string { return m.input.Value() }

// candidates is what the line could still become: the available items whose
// name starts with the word before the first space.
func (m Model) candidates() []Item {
	name, _ := splitLine(m.input.Value())
	var out []Item
	for _, it := range m.items {
		if it.Available && strings.HasPrefix(it.Name, name) {
			out = append(out, it)
		}
	}
	return out
}

// pick moves the highlight. From nothing, down lands on the first and up on
// the last, so both arrows reach the list without a wasted press.
func (m Model) pick(up bool) Model {
	n := len(m.candidates())
	if n == 0 {
		return m
	}
	if m.cursor < 0 {
		m.cursor = 0
		if up {
			m.cursor = n - 1
		}
		return m
	}
	delta := 1
	if up {
		delta = -1
	}
	m.cursor = (m.cursor + delta + n) % n
	return m
}

// complete turns the pick, or the only candidate, into typed text. An
// argument-taking command keeps the caret, and an argument already typed.
func (m Model) complete() Model {
	cands := m.candidates()
	var it Item
	switch {
	case m.cursor >= 0 && m.cursor < len(cands):
		it = cands[m.cursor]
	case len(cands) == 1:
		it = cands[0]
	default:
		return m
	}
	line := it.Name
	if it.Arg != "" {
		line += " "
		if _, arg := splitLine(m.input.Value()); arg != "" {
			line += arg
		}
	}
	m.input.SetValue(line)
	m.input.CursorEnd()
	m.cursor = -1
	return m
}

// run finishes on enter. Only an ambiguous line stays open: its answer is
// already on screen and one keystroke settles it.
func (m Model) run() (overlay.Overlay, tea.Cmd) {
	it, arg, err := m.chosen()
	var amb ambiguousError
	var ref refusedError
	switch {
	case errors.As(err, &amb), errors.As(err, &ref) && ref.siblings:
		m.err = err.Error()
		return m, nil
	case errors.Is(err, errEmpty):
		return m, finish(DoneMsg{})
	case err != nil:
		return m, finish(DoneMsg{Err: err.Error()})
	case it.Arg == "" && arg != "":
		return m, finish(DoneMsg{Err: ":" + it.Name + " takes no argument"})
	}
	var cmd tea.Cmd
	if it.Run != nil {
		cmd = it.Run(arg)
	}
	return m, finish(DoneMsg{Cmd: cmd})
}

func finish(d DoneMsg) tea.Cmd { return func() tea.Msg { return d } }

// chosen is the pick when there is one, else what the line resolves to.
func (m Model) chosen() (Item, string, error) {
	cands := m.candidates()
	if m.cursor >= 0 && m.cursor < len(cands) {
		_, arg := splitLine(m.input.Value())
		return cands[m.cursor], arg, nil
	}
	return m.resolve(m.input.Value())
}

var errEmpty = errors.New("no command")

// refusedError is an exact name that cannot run here. With siblings on the
// list the palette stays open on them, the way an ambiguous line does.
type refusedError struct {
	name     string
	siblings bool
}

func (e refusedError) Error() string { return e.name + " is not available here" }

type ambiguousError struct {
	name    string
	matches []Item
}

func (e ambiguousError) Error() string {
	names := make([]string, 0, len(e.matches))
	for _, it := range e.matches {
		names = append(names, it.Name)
	}
	slices.Sort(names)
	return e.name + " is ambiguous: " + strings.Join(names, ", ")
}

// resolve is exact name, then unique prefix. Without the first ":copy" is
// unrunnable beside ":copy-path"; without the last ":c" fires one at random.
func (m Model) resolve(line string) (Item, string, error) {
	name, arg := splitLine(line)
	if name == "" {
		return Item{}, "", errEmpty
	}
	refused := false
	for _, it := range m.items {
		if it.Name != name {
			continue
		}
		if it.Available {
			return it, arg, nil
		}
		refused = true
	}
	var prefix []Item
	for _, it := range m.items {
		if it.Available && strings.HasPrefix(it.Name, name) {
			prefix = append(prefix, it)
		}
	}
	// A refused exact name never falls through to a sibling it prefixes.
	if refused {
		return Item{}, "", refusedError{name: name, siblings: len(prefix) > 0}
	}
	switch len(prefix) {
	case 1:
		return prefix[0], arg, nil
	case 0:
		return Item{}, "", fmt.Errorf("unknown command: %s", name)
	}
	return Item{}, "", ambiguousError{name: name, matches: prefix}
}

// splitLine cuts a line into the name and its argument; only the first gap
// separates them, so the argument keeps its inner spaces.
func splitLine(line string) (name, arg string) {
	line = strings.TrimPrefix(strings.TrimSpace(line), ":")
	name, arg, found := strings.Cut(line, " ")
	if !found {
		return name, ""
	}
	return name, strings.TrimSpace(arg)
}
