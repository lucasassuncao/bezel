// Package legend is the key legend: entries tagged with the capability they
// need, filtered by what the session allows, packed into a few rows with
// whole entries dropped off the end and counted.
package legend

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/draw"
)

// Capability names something a session may refuse: "write", "apply", ...
// The zero value is always allowed.
type Capability string

// Entry is one key and what it does, plus the capability it needs. Group
// sorts it into a row: each group starts a new one while the rows last.
type Entry struct {
	key.Binding
	Needs Capability
	Group int
}

// New builds an entry from the key label, its description and, optionally,
// the capability it needs. keys is what the legend prints, e.g. "↑/↓".
func New(keys, desc string, needs ...Capability) Entry {
	e := Entry{Binding: key.NewBinding(key.WithKeys(keys), key.WithHelp(keys, desc))}
	if len(needs) > 0 {
		e.Needs = needs[0]
	}
	return e
}

// disabled reports an entry switched off with SetEnabled(false). Enabled()
// is also false for an entry with no keys, which still shows, so check both.
func (e Entry) disabled() bool { return len(e.Keys()) > 0 && !e.Enabled() }

// shown drops the disabled entries, in a new slice.
func shown(entries []Entry) []Entry {
	return slices.DeleteFunc(slices.Clone(entries), Entry.disabled)
}

// Filter keeps the entries can allows and that are not disabled, in a new
// slice. A nil can allows every capability.
func Filter(entries []Entry, can func(Capability) bool) []Entry {
	kept := make([]Entry, 0, len(entries))
	for _, e := range shown(entries) {
		if e.Needs == "" || can == nil || can(e.Needs) {
			kept = append(kept, e)
		}
	}
	return kept
}

// Style colours the bracketed key and the action text. HelpKey is the key
// that opens the full list: when set, "[HelpKey] help" opens the legend and
// the dropped-count mark names it. Empty means the app has none.
type Style struct {
	Key     lipgloss.Style
	Text    lipgloss.Style
	HelpKey string
}

func (s Style) pair(e Entry) string {
	return s.Key.Render("["+e.Help().Key+"]") + " " + s.Text.Render(e.Help().Desc)
}

// HintLine is every entry on one line, for a modal's own footer.
func HintLine(entries []Entry, st Style) string {
	entries = shown(entries)
	parts := make([]string, len(entries))
	for i, e := range entries {
		parts[i] = st.pair(e)
	}
	return strings.Join(parts, st.Text.Render(" · "))
}

// Pack lays entries over up to maxLines rows of width. Entries are dropped
// whole from the end, never the first, and the last row says how many. An
// entry for HelpKey stays where it was put; without one, help is pinned first.
func Pack(entries []Entry, width, maxLines int, st Style) []string {
	entries = shown(entries)
	slices.SortStableFunc(entries, func(a, b Entry) int { return a.Group - b.Group })

	parts := make([]string, 0, len(entries)+1)
	// starts marks where each group begins in parts; help belongs to the first.
	starts := []int{0}
	listed := slices.ContainsFunc(entries, func(e Entry) bool { return e.Help().Key == st.HelpKey })
	if st.HelpKey != "" && !listed {
		parts = append(parts, st.pair(New(st.HelpKey, "help")))
	}
	for i, e := range entries {
		if i > 0 && e.Group != entries[i-1].Group && len(parts) > starts[len(starts)-1] {
			starts = append(starts, len(parts))
		}
		parts = append(parts, st.pair(e))
	}
	if len(parts) == 0 {
		return nil
	}
	sep := st.Text.Render("  ")
	if lines, ok := packGroups(parts, starts, sep, width, maxLines); ok {
		return lines
	}
	mark := func(dropped int) string {
		if st.HelpKey == "" {
			return st.Text.Render(fmt.Sprintf("  +%d more", dropped))
		}
		return st.Text.Render(fmt.Sprintf("  +%d in [%s]", dropped, st.HelpKey))
	}

	if width <= 0 || maxLines < 1 {
		return []string{strings.Join(parts, sep)}
	}
	lines, placed := packGreedy(parts, sep, width, maxLines, 0)
	if placed < len(parts) {
		lines, placed = packGreedy(parts, sep, width, maxLines, ansi.StringWidth(mark(len(parts))))
		lines[len(lines)-1] += mark(len(parts) - placed)
	}
	for i := range lines {
		lines[i] = draw.Truncate(lines[i], width)
	}
	return lines
}

// packGroups gives each group rows of its own, or reports false when there is
// one group or the groups do not all fit whole: Pack then fills rows as usual.
func packGroups(parts []string, starts []int, sep string, width, maxLines int) ([]string, bool) {
	if len(starts) < 2 || width <= 0 {
		return nil, false
	}
	var lines []string
	for i, from := range starts {
		to := len(parts)
		if i+1 < len(starts) {
			to = starts[i+1]
		}
		left := maxLines - len(lines)
		if left < 1 {
			return nil, false
		}
		rows, placed := packGreedy(parts[from:to], sep, width, left, 0)
		if placed < to-from {
			return nil, false
		}
		lines = append(lines, rows...)
	}
	for i := range lines {
		lines[i] = draw.Truncate(lines[i], width)
	}
	return lines, true
}

// packGreedy fills rows left to right and reports how many parts it placed.
// reserveOnLast is held back on the final row for the dropped-count mark.
func packGreedy(parts []string, sep string, width, maxLines, reserveOnLast int) (lines []string, placed int) {
	sepW := ansi.StringWidth(sep)
	current, currentW := "", 0

	for _, part := range parts {
		partW := ansi.StringWidth(part)
		cost := partW
		if current != "" {
			cost += sepW
		}
		budget := width
		if len(lines)+1 == maxLines {
			budget -= reserveOnLast
		}
		if current != "" && currentW+cost > budget {
			if len(lines)+1 == maxLines {
				break
			}
			lines = append(lines, current)
			current, currentW = "", 0
			cost = partW
		}
		if current == "" {
			current = part
		} else {
			current += sep + part
		}
		currentW += cost
		placed++
	}
	if current != "" || len(lines) == 0 {
		lines = append(lines, current)
	}
	return lines, placed
}

// FromBindings turns bubbles bindings into entries, skipping disabled ones.
// For apps that already keep their keys as key.Binding.
func FromBindings(bs []key.Binding) []Entry {
	out := make([]Entry, 0, len(bs))
	for _, b := range bs {
		if b.Enabled() {
			out = append(out, Entry{Binding: b})
		}
	}
	return out
}
