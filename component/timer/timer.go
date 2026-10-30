// Package timer counts a time down a step per tick and says when it runs out.
// The time comes from its ticks, never from reading the clock.
package timer

import (
	"cmp"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/theme"
)

// TickMsg takes a step off the timer whose ID it carries. A host hands it on
// the way it hands on keys; a timer ignores every other timer's.
type TickMsg struct {
	ID  int64
	tag int64 // the run that armed it; one Stop ended lands spent
}

// TimeoutMsg says the timer whose ID it carries ran out.
type TimeoutMsg struct {
	ID int64
}

// Styles are how the time left is painted in each state.
type Styles struct {
	Running  lipgloss.Style
	Stopped  lipgloss.Style
	TimedOut lipgloss.Style
}

// DefaultStyles leaves a running timer plain, dims a stopped one and marks
// one that ran out as danger.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{Running: lipgloss.NewStyle(), Stopped: th.Dim, TimedOut: th.Danger}
}

var _ bezel.Model[Model] = Model{}

// Model is a value: every method returns the new timer.
type Model struct {
	left    time.Duration
	every   time.Duration
	running bool
	id, tag int64
	format  func(time.Duration) string
	edit    func(Styles) Styles
}

// New is a stopped timer with timeout left, stepping a second a tick. A
// negative timeout is zero.
func New(timeout time.Duration) Model {
	return Model{left: max(timeout, 0), every: time.Second, id: frames.NextID()}
}

// WithEvery steps d a tick; zero or less is a second.
func (m Model) WithEvery(d time.Duration) Model {
	if d <= 0 {
		d = time.Second
	}
	m.every = d
	return m
}

// WithFormat writes the time left with f; nil is time.Duration's String.
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

// Start runs the timer. The command is its first tick; nil when it is
// already running or has run out.
func (m Model) Start() (Model, tea.Cmd) {
	if m.running || m.TimedOut() {
		return m, nil
	}
	if m.id == 0 { // a Model never built with New
		m.id = frames.NextID()
	}
	m.running = true
	m.tag++
	return m, m.tick()
}

// Stop holds the timer; a tick already on its way lands spent.
func (m Model) Stop() Model {
	m.running = false
	m.tag++
	return m
}

// Toggle stops a running timer and starts a stopped one.
func (m Model) Toggle() (Model, tea.Cmd) {
	if m.running {
		return m.Stop(), nil
	}
	return m.Start()
}

// Reset stops the timer with timeout left.
func (m Model) Reset(timeout time.Duration) Model {
	m = m.Stop()
	m.left = max(timeout, 0)
	return m
}

func (m Model) tick() tea.Cmd {
	msg := TickMsg{ID: m.id, tag: m.tag}
	return tea.Tick(m.Every(), func(time.Time) tea.Msg { return msg })
}

// Update takes a step off on its own tick and arms the next; at zero it
// stops, and the command says it timed out.
func (m Model) Update(msg tea.Msg) (Model, bezel.Result) {
	t, ok := msg.(TickMsg)
	if !ok || t.ID != m.id {
		return m, bezel.Result{}
	}
	if !m.running || t.tag != m.tag {
		return m, bezel.Result{Handled: true}
	}
	m.left = max(m.left-m.Every(), 0)
	if !m.TimedOut() {
		return m, bezel.Result{Handled: true, Cmd: m.tick()}
	}
	m.running = false
	out := TimeoutMsg{ID: m.id}
	return m, bezel.Result{Handled: true, Cmd: func() tea.Msg { return out }}
}

// Left is the time left.
func (m Model) Left() time.Duration { return m.left }

// Every is the step a tick takes off; a second for a Model never built with
// New.
func (m Model) Every() time.Duration { return cmp.Or(m.every, time.Second) }

// Running reports whether the timer is counting down.
func (m Model) Running() bool { return m.running }

// TimedOut reports whether no time is left.
func (m Model) TimedOut() bool { return m.left <= 0 }

// ID tells this timer's messages from another's.
func (m Model) ID() int64 { return m.id }

// View is the time left in the style of the timer's state.
func (m Model) View(th theme.Resolved) string {
	st := DefaultStyles(th)
	if m.edit != nil {
		st = m.edit(st)
	}
	text := m.left.String()
	if m.format != nil {
		text = m.format(m.left)
	}
	switch {
	case m.TimedOut():
		return st.TimedOut.Render(text)
	case m.running:
		return st.Running.Render(text)
	}
	return st.Stopped.Render(text)
}
