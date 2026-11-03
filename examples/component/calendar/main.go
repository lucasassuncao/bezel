// Calendar: a month with a cursor on a day. The arrows move a day or a week,
// pgup and pgdown a month, enter or a second click chooses the day; weeks
// start on Monday and the days before today are out of range.
package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/calendar"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// today is fixed, so the recording and the tests show the same month.
var today = time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)

type model struct {
	sh     shell.Shell
	th     theme.Resolved
	cal    calendar.Model
	chosen *time.Time
}

func newModel() model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(layout.Fixed("calendar", layout.Lines(calendar.Width+4)), layout.Fill("day")),
		Theme:  th,
		Title:  "calendar",
		Mouse:  true,
		Actions: []shell.Action{
			shell.Custom("←/→", "day", nil, shell.DisplayOnly()),
			shell.Custom("↑/↓", "week", nil, shell.DisplayOnly()),
			shell.Custom("pgup/pgdn", "month", nil, shell.DisplayOnly()),
			shell.Custom("enter", "choose", nil, shell.DisplayOnly()),
			shell.Quit(),
		},
	})
	cal := calendar.New(today).WithToday(today).WithFirstWeekday(time.Monday).WithRange(today, time.Time{})
	return model{sh: sh, th: th, cal: cal}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case shell.ClickMsg:
		if msg.Pane != "calendar" {
			return m, nil
		}
	}
	m.cal, r = m.cal.Update(msg)
	if c, ok := r.Event.(calendar.Chosen); ok {
		m.chosen = &c.Day
	}
	return m, nil
}

// detail is the day under the cursor and the one chosen last.
func (m model) detail() string {
	day, _ := m.cal.Selected()
	out := m.th.Dim.Render("cursor  ") + day.Format("Monday, 2 January 2006")
	out += "\n" + m.th.Dim.Render("        ") + fmt.Sprintf("%d days from today", int(day.Sub(today).Hours()/24))
	if m.chosen != nil {
		out += "\n\n" + m.th.Dim.Render("chosen  ") + m.th.Accent.Render(m.chosen.Format(time.DateOnly))
	}
	return out
}

func (m model) View() tea.View {
	return m.sh.View(map[string]shell.Pane{
		"calendar": {Title: "month", Body: func(layout.Rect) string { return m.cal.View(m.th) }},
		"day":      {Title: "day", Body: func(r layout.Rect) string { return draw.FitBlock(m.detail(), r.W, r.H) }},
	})
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
