package tree

import (
	"maps"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/key"
)

// WithMulti lets the tree check nodes for a batch: Check flips the node under
// the cursor, CheckAll every node shown, and Checked lists them. View draws
// a box before each row.
func (m Model[N]) WithMulti(on bool) Model[N] {
	m.multi = on
	return m
}

// Checked is a copy of every checked node, in the tree's order.
func (m Model[N]) Checked() []N {
	var out []N
	for i, n := range m.nodes {
		if m.checked[i] {
			out = append(out, n)
		}
	}
	return out
}

// IsChecked reports whether node i of Nodes is checked.
func (m Model[N]) IsChecked(i int) bool { return m.checked[i] }

// ClearChecks unchecks every node.
func (m Model[N]) ClearChecks() Model[N] {
	m.checked = nil
	return m
}

// toggle is Check or CheckAll pressed, with the Toggled event when a node
// changed.
func (m Model[N]) toggle(msg tea.KeyPressMsg, keys Keys) (Model[N], bezel.Result) {
	var nodes []N
	if key.Matches(msg, keys.Check) {
		m, nodes = m.check()
	} else {
		m, nodes = m.checkAll()
	}
	r := bezel.Result{Handled: true}
	if len(nodes) > 0 {
		r.Event = Toggled[N]{Nodes: nodes}
	}
	return m, r
}

// check flips the node under the cursor; what changed comes back for the
// Toggled event.
func (m Model[N]) check() (Model[N], []N) {
	i := m.Index()
	if i < 0 || m.nodes[i].TreeInfo().Section {
		return m, nil
	}
	m.checked = maps.Clone(m.checked)
	if m.checked == nil {
		m.checked = map[int]bool{}
	}
	if m.checked[i] {
		delete(m.checked, i)
	} else {
		m.checked[i] = true
	}
	return m, []N{m.nodes[i]}
}

// checkAll checks every node shown, or unchecks them all when none was left.
func (m Model[N]) checkAll() (Model[N], []N) {
	var rows []int
	on := false
	for _, i := range m.Visible() {
		if !m.nodes[i].TreeInfo().Section {
			rows = append(rows, i)
			on = on || !m.checked[i]
		}
	}
	m.checked = maps.Clone(m.checked)
	if m.checked == nil {
		m.checked = map[int]bool{}
	}
	var changed []N
	for _, i := range rows {
		if m.checked[i] != on {
			changed = append(changed, m.nodes[i])
		}
		if on {
			m.checked[i] = true
		} else {
			delete(m.checked, i)
		}
	}
	return m, changed
}

// checkBox is the box drawn before row i while the tree is multi.
func (m Model[N]) checkBox(i int, st Styles) string {
	switch {
	case !m.multi:
		return ""
	case m.nodes[i].TreeInfo().Section:
		return "    "
	case m.checked[i]:
		return st.Checked.Render("[x]") + " "
	}
	return st.Box.Render("[ ]") + " "
}
