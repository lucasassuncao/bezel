package form

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/calendar"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

// Date is a day picked from a calendar that opens under it: Press opens it,
// the calendar's keys move, Press picks and Close closes without changing it.
type Date struct {
	label  string
	cal    calendar.Model
	day    time.Time
	format string
	open   bool
	th     *theme.Resolved
	st     *Styles
	keys   Keys
}

// NewDate offers a day, starting on day and written as 2006-01-02.
func NewDate(label string, day time.Time) Date {
	cal := calendar.New(day)
	d, _ := cal.Selected()
	return Date{label: label, cal: cal, day: d, format: time.DateOnly, keys: DefaultKeys()}
}

// WithFormat writes the day with a time.Format layout.
func (d Date) WithFormat(layout string) Date {
	d.format = layout
	return d
}

// WithCalendar sets up the calendar that opens: its range, first weekday,
// today or keys. The day stays the field's own.
func (d Date) WithCalendar(edit func(calendar.Model) calendar.Model) Date {
	d.cal = edit(d.cal).SetDay(d.day)
	d.day, _ = d.cal.Selected()
	return d
}

// WithKeys rebinds the field: Press opens and picks, Close closes.
func (d Date) WithKeys(k Keys) Date {
	d.keys = k
	return d
}

func (d Date) rebind(k Keys) Field { return d.WithKeys(k) }

func (d Date) retheme(th *theme.Resolved, st *Styles) Field {
	d.th, d.st = th, st
	return d
}

// Day is the chosen day, at midnight.
func (d Date) Day() time.Time { return d.day }

// Value is the chosen day in the field's format.
func (d Date) Value() string { return d.day.Format(d.format) }
func (d Date) Label() string { return d.label }

// Holding reports whether the calendar is open, keeping tab and esc from the form.
func (d Date) Holding() bool           { return d.open }
func (d Date) Focus() (Field, tea.Cmd) { return d, nil }

func (d Date) Blur() Field {
	d.open = false
	return d
}

// Update opens on Press; while open, Press or the calendar's enter picks the
// day under its cursor, Close closes, and the rest moves the calendar.
func (d Date) Update(msg tea.Msg) (Field, bezel.Result) {
	if !d.open {
		if is(msg, d.keys.Press) {
			d.open, d.cal = true, d.cal.SetDay(d.day)
			return d, bezel.Result{Handled: true}
		}
		return d, bezel.Result{}
	}
	switch {
	case is(msg, d.keys.Press):
		d.open, d.day = false, selected(d.cal)
		return d, bezel.Result{Handled: true}
	case is(msg, d.keys.Close):
		d.open = false
		return d, bezel.Result{Handled: true}
	}
	var r bezel.Result
	d.cal, r = d.cal.Update(msg)
	if _, ok := r.Event.(calendar.Chosen); ok {
		d.open, d.day, r.Event = false, selected(d.cal), nil
	}
	return d, r
}

func selected(c calendar.Model) time.Time {
	day, _ := c.Selected()
	return day
}

func (d Date) View(focused bool, labelW, width int) string {
	edge := styles(d.st).Edge
	if focused {
		edge = styles(d.st).FocusedEdge
	}
	head := row(styles(d.st), focused, d.label, labelW, width, edge.Render("‹ ")+d.Value()+edge.Render(" ›"))
	if !d.open {
		return head
	}
	th := theme.Resolved{}
	if d.th != nil {
		th = *d.th
	}
	indent := strings.Repeat(" ", 2+labelW+2*min(labelW, 1))
	lines := []string{head}
	for l := range strings.SplitSeq(d.cal.View(th), "\n") {
		lines = append(lines, draw.Fit(indent+l, width))
	}
	return strings.Join(lines, "\n")
}
