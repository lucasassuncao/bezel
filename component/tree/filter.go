package tree

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/key"
)

// Contains is the filter's default match: the node's text holds the query,
// in any case.
func Contains(query, text string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(query))
}

// filterMode is the / prompt: open, the text typed, and the node the cursor
// was on when it opened, where esc goes back to.
type filterMode struct {
	on   bool
	text string
	from int
}

// WithFilter lets / filter the tree by text, each node's text for matching:
// a match shows with every ancestor, open or not. nil turns it off.
func (m Model[N]) WithFilter(text func(N) string) Model[N] {
	m.text = text
	if text == nil {
		m.find = filterMode{}
	}
	return m.clamp()
}

// WithMatch filters with match, which gets the typed text and a node's text,
// colours stripped; nil is Contains again.
func (m Model[N]) WithMatch(match func(query, text string) bool) Model[N] {
	m.match = match
	return m.clamp()
}

// Filtering reports whether the / prompt is open and taking the keys.
func (m Model[N]) Filtering() bool { return m.find.on }

// Filter is the text typed at the / prompt.
func (m Model[N]) Filter() string { return m.find.text }

// openFilter opens the prompt, empty, remembering the node under the cursor.
func (m Model[N]) openFilter() Model[N] {
	m.find = filterMode{on: true, from: m.Index()}
	return m.clamp()
}

// updateFilter is the prompt: text narrows, the arrows walk what is left,
// enter keeps the node under the cursor and esc goes back to where it was.
func (m Model[N]) updateFilter(msg tea.KeyPressMsg, keys Keys) Model[N] {
	switch {
	case key.Matches(msg, keys.Esc):
		from := m.find.from
		m.find = filterMode{}
		return m.revealIndex(from)
	case key.Matches(msg, keys.Enter):
		at := m.Index()
		m.find = filterMode{}
		return m.revealIndex(at)
	case key.Matches(msg, keys.Erase):
		return m.refind(edit.Erase(m.find.text))
	case edit.Typed(msg) != "":
		return m.refind(m.find.text + edit.Typed(msg))
	case key.Matches(msg, keys.Up):
		return m.move(-1, m.Visible())
	case key.Matches(msg, keys.Down):
		return m.move(1, m.Visible())
	}
	return m
}

// refind narrows to text with the cursor on the first node that matches it,
// an ancestor shown only for context being passed over.
func (m Model[N]) refind(text string) Model[N] {
	m.find.text = text
	vis := m.Visible()
	m.cursor, m.offset = 0, 0
	for vi, i := range vis {
		if m.matches(i) {
			m.cursor = vi
			break
		}
	}
	return m.clampOffset(len(vis))
}

// narrowing reports a filter with text, the one that changes Visible.
func (m Model[N]) narrowing() bool { return m.text != nil && m.find.text != "" }

// matches reports whether node i holds the typed text.
func (m Model[N]) matches(i int) bool {
	match := m.match
	if match == nil {
		match = Contains
	}
	return match(m.find.text, ansi.Strip(m.text(m.nodes[i])))
}

// filtered is Visible while narrowing: each match and its ancestors, in
// order, whatever is open; a hidden node and all under it stay out.
func (m Model[N]) filtered() []int {
	keep := make([]bool, len(m.nodes))
	var path []int // the ancestors of the node at hand, by depth
	hidden := -1   // the depth of a hidden node being skipped, -1 for none
	for i, n := range m.nodes {
		in := n.TreeInfo()
		if hidden >= 0 && in.Depth > hidden {
			continue
		}
		hidden = -1
		if in.Hidden {
			hidden = in.Depth
			continue
		}
		for len(path) > 0 && m.nodes[path[len(path)-1]].TreeInfo().Depth >= in.Depth {
			path = path[:len(path)-1]
		}
		if !in.Section && m.matches(i) {
			keep[i] = true
			for _, a := range path {
				keep[a] = true
			}
		}
		path = append(path, i)
	}
	out := make([]int, 0, len(m.nodes))
	for i, k := range keep {
		if k {
			out = append(out, i)
		}
	}
	return out
}

// revealIndex puts the cursor on node i, opening its ancestors; a node gone
// since leaves the cursor where it is.
func (m Model[N]) revealIndex(i int) Model[N] {
	if i < 0 || i >= len(m.nodes) {
		return m.clamp()
	}
	nodes := slices.Clone(m.nodes)
	depth := nodes[i].TreeInfo().Depth
	for a := i - 1; a >= 0 && depth > 0; a-- {
		if in := nodes[a].TreeInfo(); in.Depth < depth {
			if !in.Leaf && !in.Expanded {
				nodes[a] = nodes[a].Expand(true)
			}
			depth = in.Depth
		}
	}
	m.nodes = nodes
	vis := m.Visible()
	if vi := slices.Index(vis, i); vi >= 0 {
		m.cursor = vi
	}
	return m.clampOffset(len(vis))
}
