package form

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/calendar"
	"github.com/lucasassuncao/bezel/theme"
)

func dateForm() Form {
	day := time.Date(2026, 11, 2, 9, 30, 0, 0, time.UTC)
	return New(theme.Resolve(theme.Plain(), true), NewDate("Start", day))
}

func TestDateOpensACalendarAndPicks(t *testing.T) {
	f := dateForm()
	require.Contains(t, ansi.Strip(f.SetWidth(40).View()), "‹ 2026-11-02 ›")

	f, _ = press(f, "enter")
	require.True(t, f.Field(0).(Date).Holding())
	view := ansi.Strip(f.SetWidth(40).View())
	require.Contains(t, view, "November 2026", "the calendar opens under it")
	require.Len(t, strings.Split(view, "\n"), 1+calendar.Height)

	f, _ = press(f, "tab")
	require.True(t, f.Field(0).(Date).Holding(), "tab does not leave an open calendar")

	f, _ = press(f, "right", "down", "enter")
	require.False(t, f.Field(0).(Date).Holding())
	require.Equal(t, "2026-11-10", f.Field(0).(Date).Value())
}

func TestDateEscKeepsTheDay(t *testing.T) {
	f, _ := press(dateForm(), "enter", "pgdown", "esc")
	require.Equal(t, "2026-11-02", f.Field(0).(Date).Value())
	f, _ = press(f, "enter", "pgdown", "space")
	require.Equal(t, time.Date(2026, 12, 2, 0, 0, 0, 0, time.UTC), f.Field(0).(Date).Day())
}

func TestDateWithFormatAndCalendar(t *testing.T) {
	day := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	lo := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	d := NewDate("Start", day).WithFormat("02/01/2006").
		WithCalendar(func(c calendar.Model) calendar.Model { return c.WithRange(lo, time.Time{}) })
	require.Equal(t, "10/11/2026", d.Value(), "held within the range")
}
