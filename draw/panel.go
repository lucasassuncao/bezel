package draw

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/layout"
)

// PanelStyle colours a panel: Border inks the edges, Title the label in the
// top edge, Right the text at the far end of either edge, Thumb a scrollbar's
// thumb. Edge is the box-drawing set, rounded when zero.
type PanelStyle struct {
	Border lipgloss.Style
	Title  lipgloss.Style
	Right  lipgloss.Style
	Thumb  lipgloss.Style
	Edge   lipgloss.Border
}

// thumbMark is a scrollbar's thumb: a full cell, apart from any edge's line.
const thumbMark = "█"

// Edges is what a panel draws into its border besides the lines.
type Edges struct {
	Title, Right string        // the top edge; Right is dropped before Title is cut
	Bottom       string        // the far end of the bottom edge, dropped when it does not fit
	Scroll       layout.Scroll // a thumb on the right edge; the zero value draws none
}

func (st PanelStyle) edge() lipgloss.Border {
	if st.Edge.Top == "" {
		return lipgloss.RoundedBorder()
	}
	return st.Edge
}

// Panel draws a box exactly r.W by r.H with edges drawn into its border.
// Below three rows only the top edge is drawn; below four columns nothing is.
func Panel(r layout.Rect, edges Edges, body string, st PanelStyle) string {
	if r.W < 4 || r.H < 1 {
		return ""
	}
	innerW := r.W - 2
	e := st.edge()

	label := st.Title.Render(" " + edges.Title + " ")
	if maxLabel := innerW - 1; ansi.StringWidth(label) > maxLabel {
		label = ansi.Truncate(label, maxLabel, "… ")
	}
	tail := ""
	if edges.Right != "" {
		tail = st.Right.Render(" " + edges.Right + " ")
	}
	top := st.Border.Render(e.TopLeft+e.Top) + label
	if fill := innerW - 2 - ansi.StringWidth(label) - ansi.StringWidth(tail); tail != "" && fill > 0 {
		top += st.Border.Render(strings.Repeat(e.Top, fill)) + tail + st.Border.Render(e.Top+e.TopRight)
	} else {
		top += st.Border.Render(strings.Repeat(e.Top, max(0, innerW-1-ansi.StringWidth(label))) + e.TopRight)
	}
	if r.H < 3 {
		return top
	}

	content := InnerRect(r)
	lines := strings.SplitN(body, "\n", content.H+1) // Fit below cuts each row
	rows := make([]string, 0, r.H)
	rows = append(rows, top)
	left, right, thumb := st.Border.Render(e.Left), st.Border.Render(e.Right), st.Thumb.Render(thumbMark)
	pos, size := edges.Scroll.Thumb(content.H)
	for i := 0; i < content.H; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		edge := right
		if i >= pos && i < pos+size {
			edge = thumb
		}
		rows = append(rows, left+" "+Fit(line, content.W)+" "+edge)
	}
	rows = append(rows, bottomEdge(innerW, edges.Bottom, e, st))
	return strings.Join(rows, "\n")
}

// bottomEdge is the bottom border with text at its far end, the way the top
// one carries Right, or the bare line when the text does not fit.
func bottomEdge(innerW int, text string, e lipgloss.Border, st PanelStyle) string {
	if text != "" {
		tail := st.Right.Render(" " + text + " ")
		if fill := innerW - 1 - ansi.StringWidth(tail); fill > 0 {
			return st.Border.Render(e.BottomLeft+strings.Repeat(e.Bottom, fill)) + tail + st.Border.Render(e.Bottom+e.BottomRight)
		}
	}
	return st.Border.Render(e.BottomLeft + strings.Repeat(e.Bottom, innerW) + e.BottomRight)
}
