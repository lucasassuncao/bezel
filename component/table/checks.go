package table

// WithMulti lets the table check rows for a batch: Check flips the row under
// the cursor, CheckAll every row, and Checked lists them.
func (m Model[T]) WithMulti(on bool) Model[T] {
	m.multi = on
	return m
}

// Checked is a copy of every checked row, in the table's order.
func (m Model[T]) Checked() []Item[T] {
	var out []Item[T]
	for _, it := range m.items {
		if it.Checked {
			out = append(out, cloneItems([]Item[T]{it})[0])
		}
	}
	return out
}

// ClearChecks unchecks every row.
func (m Model[T]) ClearChecks() Model[T] {
	m.items = cloneItems(m.items)
	for i := range m.items {
		m.items[i].Checked = false
	}
	return m
}

// check flips the row under the cursor; the rows that changed come back for
// the Toggled event.
func (m Model[T]) check() (Model[T], []Item[T]) {
	if m.win.Cursor >= len(m.items) {
		return m, nil
	}
	m.items = cloneItems(m.items)
	it := &m.items[m.win.Cursor]
	it.Checked = !it.Checked
	return m, cloneItems([]Item[T]{*it})
}

// checkAll checks every row, or unchecks them all when none was left.
func (m Model[T]) checkAll() (Model[T], []Item[T]) {
	on := false
	for _, it := range m.items {
		on = on || !it.Checked
	}
	m.items = cloneItems(m.items)
	var changed []Item[T]
	for i := range m.items {
		if m.items[i].Checked != on {
			m.items[i].Checked = on
			changed = append(changed, m.items[i])
		}
	}
	return m, cloneItems(changed)
}

// keepChecks checks again the rows of items whose first cell was checked in
// before, so a reload keeps what the user checked; items is already a copy.
func keepChecks[T any](before, items []Item[T]) {
	was := map[string]bool{}
	for _, it := range before {
		if it.Checked {
			was[first(it)] = true
		}
	}
	for i := range items {
		items[i].Checked = items[i].Checked || was[first(items[i])]
	}
}

// box is the check column of a row while the table is multi, "" otherwise.
func (m Model[T]) box(it Item[T], st Styles) string {
	switch {
	case !m.multi:
		return ""
	case it.Checked:
		return st.Checked.Render("[x]") + " "
	}
	return st.Box.Render("[ ]") + " "
}
