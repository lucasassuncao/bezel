package palette

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/fuzzy"
	"github.com/lucasassuncao/bezel/layout"
)

// keyWidth keeps the key column even on rows with no key: that column is the
// slow path teaching the fast one.
const keyWidth = 7

// View is the candidate list, a rule above it, every line padded to the full
// width so the panes do not show through.
func (m Model) View(body layout.Rect) string {
	w := max(body.W, 0)
	var lines []string
	cands := m.candidates()
	if len(cands) == 0 {
		lines = append(lines, m.styles().Note.Render("  no command matches"))
	} else {
		limit := m.rowLimit(body.H, len(cands))
		// The window follows the cursor, clamped so a stale one cannot overrun.
		start := min(max(m.cursor-limit+1, 0), max(0, len(cands)-1))
		end := min(start+limit, len(cands))
		// Measured over every candidate, not the window, so scrolling keeps the columns.
		nameW := 0
		for _, it := range cands {
			nameW = max(nameW, draw.Width(label(it)))
		}
		for i := start; i < end; i++ {
			lines = append(lines, m.row(cands[i], nameW, i == m.cursor, w, m.hits(cands[i])))
		}
		if end < len(cands) {
			lines = append(lines, m.styles().Note.Render(fmt.Sprintf("  … %d more", len(cands)-end)))
		}
	}
	if m.err != "" {
		lines = append(lines, m.styles().Error.Render("  "+m.err))
	}
	lines = append([]string{draw.Divider(w, "", draw.DividerStyle{Line: m.styles().Divider})}, lines...)
	for i, l := range lines {
		lines[i] = draw.Pad(draw.Cut(l, w), w)
	}
	return strings.Join(lines, "\n")
}

// rowLimit is how many candidates fit a body height h, after the rule, the
// error line and the "… N more" row. h <= 0 is a body not yet sized.
func (m Model) rowLimit(h, n int) int {
	limit := m.Height()
	if h > 0 {
		limit = min(limit, h-1)
		if m.err != "" {
			limit--
		}
		if n > limit {
			limit--
		}
	}
	return max(limit, 1)
}

// StatusLine is the typed line, drawn where the status usually is.
func (m Model) StatusLine(width int) string { return draw.Fit(m.input.View(), width) }

func (m Model) row(it Item, nameW int, picked bool, w int, hits []int) string {
	st, lead := m.styles().Name, "  "
	if picked {
		st, lead = m.styles().Cursor, "› "
	}
	name := underlined(label(it), hits, st) + st.Render(strings.Repeat(" ", nameW-draw.Width(label(it))))
	k := draw.Pad(it.Key, keyWidth)
	if picked {
		return draw.Cut(st.Render(lead)+name+st.Render("  "+k+"  "+it.Desc), w)
	}
	return draw.Cut(lead+name+"  "+m.styles().Key.Render(k)+"  "+m.styles().Desc.Render(it.Desc), w)
}

// hits is where the typed name's letters fell in it's label, for a fuzzy
// match only: a prefix needs no marking.
func (m Model) hits(it Item) []int {
	name, _ := splitLine(m.input.Value())
	if name == "" || strings.HasPrefix(it.Name, name) {
		return nil
	}
	_, at, _ := fuzzy.Match(name, it.Name)
	for i := range at {
		at[i]++ // the label's ":"
	}
	return at
}

// underlined is s in st, the runes at hits underlined.
func underlined(s string, hits []int, st lipgloss.Style) string {
	if len(hits) == 0 {
		return st.Render(s)
	}
	var b strings.Builder
	for i, r := range []rune(s) {
		if slices.Contains(hits, i) {
			b.WriteString(st.Underline(true).Render(string(r)))
		} else {
			b.WriteString(st.Render(string(r)))
		}
	}
	return b.String()
}

// label is how a command is written: ":goto <path>" for one taking an argument.
func label(it Item) string {
	if it.Arg == "" {
		return ":" + it.Name
	}
	return ":" + it.Name + " <" + it.Arg + ">"
}
