package list

// WithMulti lets the list check rows for a batch: Check flips the row under
// the cursor, CheckAll every row, and Checked lists them.
func (m Model[T]) WithMulti(on bool) Model[T] {
	m.multi = on
	return m
}

// Checked is a copy of every checked row, in the list's order.
func (m Model[T]) Checked() []Row[T] {
	var out []Row[T]
	for _, r := range m.rows {
		if r.Checked && !r.Section {
			out = append(out, r)
		}
	}
	return out
}

// ClearChecks unchecks every row.
func (m Model[T]) ClearChecks() Model[T] {
	m.rows = cloneRows(m.rows, func(r *Row[T]) { r.Checked = false })
	return m
}

// check flips the row under the cursor, the filter's while it is open; the
// rows that changed come back for the Toggled event.
func (m Model[T]) check() (Model[T], []Row[T]) {
	i := m.at()
	if i < 0 {
		return m, nil
	}
	m.rows = cloneRows(m.rows, nil)
	m.rows[i].Checked = !m.rows[i].Checked
	return m, []Row[T]{m.rows[i]}
}

// checkAll checks every row, or unchecks them all when none was left.
func (m Model[T]) checkAll() (Model[T], []Row[T]) {
	on := false
	for _, r := range m.rows {
		if !r.Section && !r.Checked {
			on = true
		}
	}
	var changed []Row[T]
	m.rows = cloneRows(m.rows, func(r *Row[T]) {
		if !r.Section && r.Checked != on {
			r.Checked = on
			changed = append(changed, *r)
		}
	})
	return m, changed
}

// keepChecks is a copy of rows with the labels checked in before checked
// again, so a reload keeps what the user checked.
func keepChecks[T any](before, rows []Row[T]) []Row[T] {
	was := map[string]bool{}
	for _, r := range before {
		if r.Checked {
			was[r.Label] = true
		}
	}
	return cloneRows(rows, func(r *Row[T]) { r.Checked = r.Checked || (!r.Section && was[r.Label]) })
}

// cloneRows is a copy of rows with edit applied to each, when given.
func cloneRows[T any](rows []Row[T], edit func(*Row[T])) []Row[T] {
	out := make([]Row[T], len(rows))
	copy(out, rows)
	if edit != nil {
		for i := range out {
			edit(&out[i])
		}
	}
	return out
}

// checkBox is the box drawn before a row's label while the list is multi.
func (m Model[T]) checkBox(r Row[T], st *Styles) string {
	switch {
	case !m.multi || r.Section:
		return ""
	case r.Checked:
		return st.Checked.Render("[x]") + " "
	}
	return st.Box.Render("[ ]") + " "
}
