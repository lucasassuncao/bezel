package tree

import (
	"strings"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

// Branch is one row of a tree to print, and the rows under it.
type Branch struct {
	Text     string
	Children []Branch
}

// Static is the whole tree as text to print, every branch open: each top
// branch on its own line, the rows under it hung from dimmed guides.
func Static(branches []Branch, th theme.Resolved) string {
	var rows []row
	var walk func(bs []Branch, depth int)
	walk = func(bs []Branch, depth int) {
		for _, b := range bs {
			rows = append(rows, row{text: b.Text, depth: depth, leaf: len(b.Children) == 0})
			walk(b.Children, depth+1)
		}
	}
	walk(branches, 0)
	guides := New(rows).Guides()
	lines := make([]string, len(rows))
	for i, r := range rows {
		lines[i] = r.text
		if guides[i] != "" {
			lines[i] = th.Dim.Render(guides[i]) + r.text
		}
		lines[i] = draw.ExpandTabs(lines[i]) // tab stops count the guide's cells
	}
	return strings.Join(lines, "\n")
}

// row is a Branch flattened, open, for Guides to measure.
type row struct {
	text  string
	depth int
	leaf  bool
}

func (r row) TreeInfo() Info  { return Info{Depth: r.depth, Leaf: r.leaf, Expanded: true} }
func (r row) Expand(bool) row { return r }
