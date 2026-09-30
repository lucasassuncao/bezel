package draw

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/layout"
)

// PanelStyle colours a panel: Border inks the edges, Title the label in the
// top edge. Edge is the box-drawing set, rounded when zero; a focused panel
// can wear a thicker one so focus survives a terminal without colour.
type PanelStyle struct {
	Border lipgloss.Style
	Title  lipgloss.Style
	Edge   lipgloss.Border
}

func (st PanelStyle) edge() lipgloss.Border {
	if st.Edge.Top == "" {
		return lipgloss.RoundedBorder()
	}
	return st.Edge
}

// Panel draws a box with the title in its top edge, exactly r.W by r.H.
// Below three rows only the title edge is drawn; below four columns
// nothing is.
func Panel(r layout.Rect, title, body string, st PanelStyle) string {
	if r.W < 4 || r.H < 1 {
		return ""
	}
	innerW := r.W - 2
	e := st.edge()

	label := st.Title.Render(" " + title + " ")
	if maxLabel := innerW - 1; ansi.StringWidth(label) > maxLabel {
		label = ansi.Truncate(label, maxLabel, "… ")
	}
	fill := max(0, innerW-1-ansi.StringWidth(label))
	top := st.Border.Render(e.TopLeft+e.Top) + label + st.Border.Render(strings.Repeat(e.Top, fill)+e.TopRight)
	if r.H < 3 {
		return top
	}

	content := InnerRect(r)
	lines := strings.Split(FitBlock(body, content.W, content.H), "\n")
	rows := make([]string, 0, r.H)
	rows = append(rows, top)
	left, right := st.Border.Render(e.Left), st.Border.Render(e.Right)
	for i := 0; i < content.H; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		rows = append(rows, left+" "+Fit(line, content.W)+" "+right)
	}
	rows = append(rows, st.Border.Render(e.BottomLeft+strings.Repeat(e.Bottom, innerW)+e.BottomRight))
	return strings.Join(rows, "\n")
}
