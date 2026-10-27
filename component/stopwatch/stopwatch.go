// Package stopwatch counts time up a step per tick, paused and reset at will.
// The time comes from its ticks, never from reading the clock.
package stopwatch

import (
	"cmp"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/theme"
)

// TickMsg adds a step to the stopwatch whose ID it carries. A host hands it
// on the way it hands on keys; a stopwatch ignores every other one's.
type TickMsg struct {
	ID  int64
	tag int64 // the run that armed it; one Stop ended lands spent
}

// Styles are how the elapsed time is painted in each state.
type Styles struct {
	Running lipgloss.Style
	Stopped lipgloss.Style
}

// DefaultStyles leaves a running stopwatch plain and dims a stopped one.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{Running: lipgloss.NewStyle(), Stopped: th.Dim}
}

// Model is a value: every method returns the new stopwatch.
type Model struct {
	elapsed time.Duration
	every   time.Duration
	running bool
	id, tag int64
	format  func(time.Duration) string
	edit    func(Styles) Styles
}

// New is a stopped stopwatch at zero, stepping a second a tick.
func New() Model { return Model{every: time.Second, id: frames.NextID()} }

// WithEvery steps d a tick; zero or less is a second.
func (m Model) WithEvery(d time.Duration) Model {
	if d <= 0 {
		d = time.Second
	}
	m.every = d
	return m
}

// WithFormat writes the elapsed time with f; nil is time.Duration's String.
func (m Model) WithFormat(f func(time.Duration) string) Model {
	m.format = f
	return m
}

// WithStyles changes the styles derived from the theme through edit, under
// whatever theme View gets. It replaces an earlier edit.
func (m Model) WithStyles(edit func(Styles) Styles) Model {
	m.edit = edit
	return m
}

// Start runs the stopwatch. The command is its first tick; nil when it is
// already running.
func (m Model) Start() (Model, tea.Cmd) {
	if m.running {
		return m, nil
	}
	if m.id == 0 { // a Model never built with New
		m.id = frames.NextID()
	}
	m.running = true
	m.tag++
	return m, m.tick()
}

// Stop holds the stopwatch; a tick already on its way lands spent.
func (m Model) Stop() Model {
	m.running = false
	m.tag++
	return m
}

// Toggle stops a running stopwatch and starts a stopped one.
func (m Model) Toggle() (Model, tea.Cmd) {
	if m.running {
		return m.Stop(), nil
	}
	return m.Start()
}

// Reset puts the time back to zero; a running stopwatch keeps running.
func (m Model) Reset() Model {
	m.elapsed = 0
	return m
}

func (m Model) tick() tea.Cmd {
	msg := TickMsg{ID: m.id, tag: m.tag}
	return tea.Tick(m.Every(), func(time.Time) tea.Msg { return msg })
}

// Update adds a step on its own tick and arms the next. Any other message
// leaves it as it was, with no command.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	t, ok := msg.(TickMsg)
	if !ok || !m.running || t.ID != m.id || t.tag != m.tag {
		return m, nil
	}
	m.elapsed += m.Every()
	return m, m.tick()
}

// Elapsed is the time counted so far.
func (m Model) Elapsed() time.Duration { return m.elapsed }

// Every is the step a tick adds; a second for a Model never built with New.
func (m Model) Every() time.Duration { return cmp.Or(m.every, time.Second) }

// Running reports whether the stopwatch is counting.
func (m Model) Running() bool { return m.running }

// ID tells this stopwatch's ticks from another's.
func (m Model) ID() int64 { return m.id }

// View is the elapsed time in the style of the stopwatch's state.
func (m Model) View(th theme.Resolved) string {
	st := DefaultStyles(th)
	if m.edit != nil {
		st = m.edit(st)
	}
	text := m.elapsed.String()
	if m.format != nil {
		text = m.format(m.elapsed)
	}
	if m.running {
		return st.Running.Render(text)
	}
	return st.Stopped.Render(text)
}
