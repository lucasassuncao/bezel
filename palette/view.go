package palette

import (
	"fmt"
	"strings"

	"github.com/lucasassuncao/bezel/draw"
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
		lines = append(lines, m.th.Dim.Render("  no command matches"))
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
			lines = append(lines, m.row(cands[i], nameW, i == m.cursor, w))
		}
		if end < len(cands) {
			lines = append(lines, m.th.Dim.Render(fmt.Sprintf("  … %d more", len(cands)-end)))
		}
	}
	if m.err != "" {
		lines = append(lines, m.th.Danger.Render("  "+m.err))
	}
	lines = append([]string{m.th.Dim.Render(strings.Repeat("─", w))}, lines...)
	for i, l := range lines {
		lines[i] = draw.Pad(draw.Cut(l, w), w)
	}
	return strings.Join(lines, "\n")
}

// rowLimit is how many candidates fit a body height h, after the rule, the
// error line and the "… N more" row. h <= 0 is a body not yet sized.
func (m Model) rowLimit(h, n int) int {
	limit := Rows
	if h > 0 {
		limit = min(Rows, h-1)
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

func (m Model) row(it Item, nameW int, picked bool, w int) string {
	name := draw.Pad(label(it), nameW)
	k := draw.Pad(it.Key, keyWidth)
	if picked {
		return draw.Cut(m.th.Cursor.Render("› "+name+"  "+k+"  "+it.Desc), w)
	}
	return draw.Cut("  "+m.th.Key.Render(name)+"  "+m.th.Legend.Key.Render(k)+"  "+m.th.Dim.Render(it.Desc), w)
}

// label is how a command is written: ":goto <path>" for one taking an argument.
func label(it Item) string {
	if it.Arg == "" {
		return ":" + it.Name
	}
	return ":" + it.Name + " <" + it.Arg + ">"
}
