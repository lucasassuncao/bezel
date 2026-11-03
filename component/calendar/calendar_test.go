package calendar

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Plain(), true)

func day(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

func press(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }

func cursor(m Model) time.Time { d, _ := m.Selected(); return d }

func lines(m Model) []string { return strings.Split(ansi.Strip(m.View(th)), "\n") }

func TestViewIsTheMonthInSixWeeks(t *testing.T) {
	got := lines(New(day(2026, 11, 2)))
	require.Len(t, got, Height)
	require.Equal(t, "   November 2026    ", got[0])
	require.Equal(t, "Su Mo Tu We Th Fr Sa", got[1])
	require.Equal(t, " 1  2  3  4  5  6  7", got[2], "November 2026 starts on a Sunday")
	require.Equal(t, "29 30  1  2  3  4  5", got[6], "December's days fill the grid")
	for _, l := range got {
		require.Equal(t, Width, ansi.StringWidth(l))
	}
}

func TestWeeksCanStartOnMonday(t *testing.T) {
	got := lines(New(day(2026, 11, 2)).WithFirstWeekday(time.Monday))
	require.Equal(t, "Mo Tu We Th Fr Sa Su", got[1])
	require.Equal(t, "26 27 28 29 30 31  1", got[2])
}

func TestTheKeysMoveADayAWeekAndAMonth(t *testing.T) {
	m := New(day(2026, 1, 31).Add(15 * time.Hour))
	require.Equal(t, day(2026, 1, 31), cursor(m), "at midnight")
	m, r := m.Update(press(tea.KeyRight))
	require.True(t, r.Handled)
	require.Equal(t, day(2026, 2, 1), cursor(m))
	m, _ = m.Update(press(tea.KeyUp))
	require.Equal(t, day(2026, 1, 25), cursor(m))
	m, _ = m.Update(press(tea.KeyPgDown))
	require.Equal(t, day(2026, 2, 25), cursor(m))

	m = New(day(2026, 1, 31))
	m, _ = m.Update(press(tea.KeyPgDown))
	require.Equal(t, day(2026, 2, 28), cursor(m), "held at the shorter month's end")
}

func TestEnterChoosesTheDay(t *testing.T) {
	_, r := New(day(2026, 11, 2)).Update(press(tea.KeyEnter))
	require.Equal(t, Chosen{Day: day(2026, 11, 2)}, r.Event)
	_, r = New(day(2026, 11, 2)).Update(press('x'))
	require.False(t, r.Handled)
}

func TestWithRangeHoldsTheCursor(t *testing.T) {
	m := New(day(2026, 11, 2)).WithRange(day(2026, 11, 1), day(2026, 11, 30))
	m, _ = m.Update(press(tea.KeyUp))
	require.Equal(t, day(2026, 11, 1), cursor(m))
	m, _ = m.Update(press(tea.KeyPgDown))
	require.Equal(t, day(2026, 11, 30), cursor(m))
	require.Equal(t, day(2026, 11, 30), cursor(New(day(2027, 1, 1)).WithRange(time.Time{}, day(2026, 11, 30))))
}

func TestTodayAndTheCursorWearTheirStyles(t *testing.T) {
	st := DefaultStyles(th)
	v := New(day(2026, 11, 2)).WithToday(day(2026, 11, 5)).View(th)
	require.Contains(t, v, st.Cursor.Render(" 2"))
	require.Contains(t, v, st.Today.Render(" 5"))
}

func TestAClickMovesThenChooses(t *testing.T) {
	m := New(day(2026, 11, 2))
	at := bezel.ClickMsg{MouseClickMsg: tea.MouseClickMsg{X: 9, Y: 3, Button: tea.MouseLeft}}
	m, r := m.Update(at)
	require.True(t, r.Handled)
	require.Equal(t, day(2026, 11, 11), cursor(m), "row 2, the fourth column")
	_, r = m.Update(at)
	require.Equal(t, Chosen{Day: day(2026, 11, 11)}, r.Event)
	_, r = m.Update(bezel.ClickMsg{MouseClickMsg: tea.MouseClickMsg{Y: 0}})
	require.False(t, r.Handled, "the title is no day")
}

func TestVimKeysAddTheLetters(t *testing.T) {
	m, r := New(day(2026, 11, 2)).WithKeys(VimKeys()).Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	require.True(t, r.Handled)
	require.Equal(t, day(2026, 11, 3), cursor(m))
	require.Len(t, New(day(2026, 11, 2)).Legend(), 4)
}
