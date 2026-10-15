package tree

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/theme"
)

// bigTree holds parents expanded rows of 9 leaves each.
func bigTree(parents, height int) Model[node] {
	nodes := make([]node, 0, parents*10)
	for i := range parents {
		nodes = append(nodes, node{label: fmt.Sprint("p", i), expanded: true})
		for j := range 9 {
			nodes = append(nodes, node{label: fmt.Sprint("l", j), depth: 1, leaf: true})
		}
	}
	return New(nodes, height)
}

var benchSizes = []int{100, 10_000}

func BenchmarkUpdateDown(b *testing.B) {
	down := press(tea.KeyDown)
	for _, n := range benchSizes {
		b.Run(fmt.Sprint(n*10, "nodes"), func(b *testing.B) {
			m := bigTree(n, 40)
			for b.Loop() {
				m, _ = m.Update(down)
			}
		})
	}
}

func BenchmarkView(b *testing.B) {
	th := theme.Resolve(theme.Plain, true)
	row := func(n node, _ int, _ bool) string { return n.label }
	for _, n := range benchSizes {
		b.Run(fmt.Sprint(n*10, "nodes"), func(b *testing.B) {
			m := bigTree(n, 40)
			for b.Loop() {
				_ = m.View(th, row)
			}
		})
	}
}
