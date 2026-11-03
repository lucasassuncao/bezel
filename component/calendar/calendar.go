// Package calendar is a month in a grid with a cursor on a day: the arrows
// move a day or a week, pgup and pgdown a month, and enter chooses the day.
// Nothing reads the clock: the app says which day is today.
package calendar

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Width and Height are the cells View always takes: seven days of two cells
// with a gap, under the month and the weekdays, over six weeks.
const (
	Width  = 7*3 - 1
	Height = 2 + weeks
)

// weeks is the rows of days: six cover any month, so the grid never jumps.
const weeks = 6

// Keys are the bindings that move the cursor and choose a day. Each pair's
// second has no help of its own: the first names both.
type Keys struct {
	Left, Right, Up, Down key.Binding
	PrevMonth, NextMonth  key.Binding
	Enter                 key.Binding
}

// DefaultKeys are the arrows for a day and a week, pgup/pgdown for a month
// and enter to choose.
func DefaultKeys() Keys {
	return Keys{
		Left:      key.New("left").WithHelp("←/→", "day"),
		Right:     key.New("right"),
		Up:        key.New("up").WithHelp("↑/↓", "week"),
		Down:      key.New("down"),
		PrevMonth: key.New("pgup").WithHelp("pgup/pgdn", "month"),
		NextMonth: key.New("pgdown"),
		Enter:     key.New("enter").WithHelp("enter", "choose"),
	}
}

// VimKeys are DefaultKeys with vim's letters added: h/l for a day, k/j for a
// week and ctrl+b/f for a month.
func VimKeys() Keys {
	k := DefaultKeys()
	k.Left, k.Right = key.New("left", "h").WithHelp("←/→ h/l", "day"), key.New("right", "l")
	k.Up, k.Down = key.New("up", "k").WithHelp("↑/↓ k/j", "week"), key.New("down", "j")
	k.PrevMonth, k.NextMonth = key.New("pgup", "ctrl+b").WithHelp("^b/^f", "month"), key.New("pgdown", "ctrl+f")
	return k
}

// Styles are how a calendar paints its parts.
type Styles struct {
	Title    lipgloss.Style // the month and year
	Weekday  lipgloss.Style // the row of weekday names
	Day      lipgloss.Style // a day of the month shown
	Outside  lipgloss.Style // a day of the month before or after
	Today    lipgloss.Style
	Cursor   lipgloss.Style // the day under the cursor
	Disabled lipgloss.Style // a day out of WithRange
}

// DefaultStyles bold the month, mute the weekdays and the days around it,
// mark today in the accent and the cursor as the caret.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{
		Title: th.Bold, Weekday: th.Muted, Day: lipgloss.NewStyle(), Outside: th.Muted,
		Today: th.Accent.Bold(true), Cursor: th.Caret, Disabled: th.Muted.Faint(true),
	}
}

var _ bezel.Model[Model] = Model{}

// Model is a value: every method returns the new calendar.
type Model struct {
	day    time.Time // the cursor's, at midnight
	today  time.Time // zero for none
	first  time.Weekday
	lo, hi time.Time // zero for unbounded
	keys   Keys
	edit   func(Styles) Styles
}

// New is the month of day with the cursor on it, weeks from Sunday.
func New(day time.Time) Model {
	return Model{day: date(day), keys: DefaultKeys()}
}

// date is t's day at midnight, in t's location.
func date(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// WithFirstWeekday starts each week on d instead of Sunday.
func (m Model) WithFirstWeekday(d time.Weekday) Model {
	m.first = d % 7
	return m
}

// WithToday marks day as today; the calendar never asks the clock.
func (m Model) WithToday(day time.Time) Model {
	m.today = date(day)
	return m
}

// WithRange keeps the cursor from lo to hi, both days included, and draws
// the days past them disabled; a zero end is open.
func (m Model) WithRange(lo, hi time.Time) Model {
	m.lo, m.hi = zeroOrDate(lo), zeroOrDate(hi)
	return m.SetDay(m.day)
}

func zeroOrDate(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return date(t)
}

// WithKeys rebinds the calendar.
func (m Model) WithKeys(k Keys) Model {
	m.keys = k
	return m
}

// WithStyles changes the styles derived from the theme through edit, under
// whatever theme View gets. It replaces an earlier edit.
func (m Model) WithStyles(edit func(Styles) Styles) Model {
	m.edit = edit
	return m
}

func (m Model) styles(th theme.Resolved) Styles {
	st := DefaultStyles(th)
	if m.edit != nil {
		st = m.edit(st)
	}
	return st
}

// SetDay puts the cursor on day, kept within WithRange, showing its month.
func (m Model) SetDay(day time.Time) Model {
	day = date(day)
	if !m.lo.IsZero() && day.Before(m.lo) {
		day = m.lo
	}
	if !m.hi.IsZero() && day.After(m.hi) {
		day = m.hi
	}
	m.day = day
	return m
}

// Selected is the day under the cursor, at midnight; always there.
func (m Model) Selected() (time.Time, bool) { return m.day, true }

// Legend lists the keys the calendar answers to.
func (m Model) Legend() []legend.Entry {
	k := m.keys
	return legend.FromBindings([]key.Binding{k.Left, k.Right, k.Up, k.Down, k.PrevMonth, k.NextMonth, k.Enter})
}

// Update moves the cursor on its keys, held within WithRange; enter, or a
// click on the day already under the cursor, reports a [Chosen]. A click on
// another day of the grid moves there. Anything else comes back unhandled.
func (m Model) Update(msg tea.Msg) (Model, bezel.Result) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		k := m.keys
		switch {
		case key.Matches(msg, k.Left):
			return m.SetDay(m.day.AddDate(0, 0, -1)), bezel.Result{Handled: true}
		case key.Matches(msg, k.Right):
			return m.SetDay(m.day.AddDate(0, 0, 1)), bezel.Result{Handled: true}
		case key.Matches(msg, k.Up):
			return m.SetDay(m.day.AddDate(0, 0, -7)), bezel.Result{Handled: true}
		case key.Matches(msg, k.Down):
			return m.SetDay(m.day.AddDate(0, 0, 7)), bezel.Result{Handled: true}
		case key.Matches(msg, k.PrevMonth):
			return m.SetDay(addMonths(m.day, -1)), bezel.Result{Handled: true}
		case key.Matches(msg, k.NextMonth):
			return m.SetDay(addMonths(m.day, 1)), bezel.Result{Handled: true}
		case key.Matches(msg, k.Enter):
			return m, bezel.Result{Handled: true, Event: Chosen{Day: m.day}}
		}
	case bezel.ClickMsg:
		day, ok := m.dayAt(msg.X, msg.Y)
		switch {
		case !ok || !m.allowed(day):
			return m, bezel.Result{}
		case day.Equal(m.day):
			return m, bezel.Result{Handled: true, Event: Chosen{Day: m.day}}
		}
		return m.SetDay(day), bezel.Result{Handled: true}
	}
	return m, bezel.Result{}
}

// addMonths is t moved n months, its day of the month held at the new
// month's last: a month on from January 31st is the end of February.
func addMonths(t time.Time, n int) time.Time {
	first := time.Date(t.Year(), t.Month()+time.Month(n), 1, 0, 0, 0, 0, t.Location())
	return first.AddDate(0, 0, min(t.Day(), daysIn(first))-1)
}

// daysIn is how many days t's month has.
func daysIn(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
}

// start is the first day of the grid: the week's first day on or before the
// first of the cursor's month.
func (m Model) start() time.Time {
	first := time.Date(m.day.Year(), m.day.Month(), 1, 0, 0, 0, 0, m.day.Location())
	back := (int(first.Weekday()) - int(m.first) + 7) % 7
	return first.AddDate(0, 0, -back)
}

// dayAt is the day drawn at cell (x, y) of View, false off the grid.
func (m Model) dayAt(x, y int) (time.Time, bool) {
	row, col := y-2, x/3
	if row < 0 || row >= weeks || x < 0 || col > 6 {
		return time.Time{}, false
	}
	return m.start().AddDate(0, 0, row*7+col), true
}

// allowed reports whether day is within WithRange.
func (m Model) allowed(day time.Time) bool {
	return (m.lo.IsZero() || !day.Before(m.lo)) && (m.hi.IsZero() || !day.After(m.hi))
}

// View is the month and year, the weekday names and six weeks of days, in
// Width by Height cells.
func (m Model) View(th theme.Resolved) string {
	st := m.styles(th)
	title := fmt.Sprintf("%s %d", m.day.Month(), m.day.Year())
	lines := []string{st.Title.Render(draw.Pad(strings.Repeat(" ", (Width-draw.Width(title))/2)+title, Width))}
	names := make([]string, 7)
	for i := range names {
		names[i] = ((m.first + time.Weekday(i)) % 7).String()[:2]
	}
	lines = append(lines, st.Weekday.Render(strings.Join(names, " ")))
	day := m.start()
	for range weeks {
		cells := make([]string, 7)
		for i := range cells {
			cells[i] = m.paint(day, st).Render(fmt.Sprintf("%2d", day.Day()))
			day = day.AddDate(0, 0, 1)
		}
		lines = append(lines, strings.Join(cells, " "))
	}
	return strings.Join(lines, "\n")
}

// paint is the style a day of the grid wears.
func (m Model) paint(day time.Time, st Styles) lipgloss.Style {
	switch {
	case day.Equal(m.day):
		return st.Cursor
	case !m.allowed(day):
		return st.Disabled
	case day.Month() != m.day.Month():
		return st.Outside
	case day.Equal(m.today):
		return st.Today
	}
	return st.Day
}
