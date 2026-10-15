package tree_test

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/theme"
	"github.com/lucasassuncao/bezel/tree"
)

// dir is the caller's node: the tree reads it through TreeInfo and
// opens it through Expand, never through the fields.
type dir struct {
	name  string
	depth int
	leaf  bool
	open  bool
}

func (d dir) TreeInfo() tree.Info {
	return tree.Info{Depth: d.depth, Leaf: d.leaf, Expanded: d.open}
}

func (d dir) Expand(open bool) dir {
	d.open = open
	return d
}

func Example() {
	m := tree.New([]dir{
		{name: "cmd"},
		{name: "main.go", depth: 1, leaf: true},
		{name: "go.mod", leaf: true},
	}, 5)
	row := func(d dir, _ int, selected bool) string {
		mark := "  "
		if selected {
			mark = "> "
		}
		return mark + fmt.Sprintf("%*s%s", 2*d.depth, "", d.name)
	}
	th := theme.Resolve(theme.Plain, true)

	m, action := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	fmt.Println(action == tree.Expanded)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	fmt.Print(m.View(th, row))
	// Output:
	// true
	//   cmd
	// >   main.go
	//   go.mod
}
