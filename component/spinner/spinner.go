// Package spinner is a run of frames that turns while something works: twelve
// presets or frames of your own, each spinner taking only its own ticks.
package spinner

import (
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/theme"
)

// Spinner is the frames a spinner turns through and how long each shows.
type Spinner struct {
	Frames []string
	Every  time.Duration
}

// The presets are functions so that each caller gets frames of its own: a
// slice in a package variable could be changed for every program.

// Line is a bar turning around its middle.
func Line() Spinner { return Spinner{Frames: []string{"|", "/", "-", "\\"}, Every: time.Second / 10} }

// Dot is a braille cell with one dot missing, moving around it; each frame
// carries a space that sets the text after it apart.
func Dot() Spinner {
	return Spinner{Frames: []string{"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "}, Every: time.Second / 10}
}

// MiniDot is a small braille dot running in a circle.
func MiniDot() Spinner {
	return Spinner{Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}, Every: time.Second / 12}
}

// Jump is a braille dot hopping along the bottom of its cell.
func Jump() Spinner {
	return Spinner{Frames: []string{"⢄", "⢂", "⢁", "⡁", "⡈", "⡐", "⡠"}, Every: time.Second / 10}
}

// Pulse is a block fading out.
func Pulse() Spinner {
	return Spinner{Frames: []string{"█", "▓", "▒", "░"}, Every: time.Second / 8}
}

// Points is a dot walking along three.
func Points() Spinner {
	return Spinner{Frames: []string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●"}, Every: time.Second / 7}
}

// Globe is the earth turning.
func Globe() Spinner {
	return Spinner{Frames: []string{"🌍", "🌎", "🌏"}, Every: time.Second / 4}
}

// Moon is the moon through its phases.
func Moon() Spinner {
	return Spinner{Frames: []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}, Every: time.Second / 8}
}

// Monkey is the three wise monkeys.
func Monkey() Spinner {
	return Spinner{Frames: []string{"🙈", "🙉", "🙊"}, Every: time.Second / 3}
}

// Meter is three cells filling and emptying.
func Meter() Spinner {
	return Spinner{Frames: []string{"▱▱▱", "▰▱▱", "▰▰▱", "▰▰▰", "▰▰▱", "▰▱▱", "▱▱▱"}, Every: time.Second / 7}
}

// Hamburger is a trigram whose broken bar moves. Unicode counts a trigram two
// cells wide and many terminals draw it in one, so what follows can shift.
func Hamburger() Spinner {
	return Spinner{Frames: []string{"☱", "☲", "☴", "☲"}, Every: time.Second / 3}
}

// Ellipsis is three dots appearing one by one.
func Ellipsis() Spinner {
	return Spinner{Frames: []string{"", ".", "..", "..."}, Every: time.Second / 3}
}

// defaultEvery is the pace of a Spinner that sets none.
const defaultEvery = time.Second / 10

// Styles are how the frame is painted.
type Styles struct {
	Frame lipgloss.Style
}

// DefaultStyles paints the frame in the theme's accent.
func DefaultStyles(th theme.Resolved) Styles { return Styles{Frame: th.Accent} }

// Model is a value: every method returns the new spinner.
type Model struct {
	spinner Spinner
	width   int // the widest frame's cells, so the text after it never moves
	frame   int
	running bool
	id, tag int64
	edit    func(Styles) Styles
}

// New is s, stopped on its first frame. Its frames are copied, and an Every
// of zero or less is a tenth of a second.
func New(s Spinner) Model {
	s.Frames = slices.Clone(s.Frames)
	if s.Every <= 0 {
		s.Every = defaultEvery
	}
	w := 0
	for _, f := range s.Frames {
		w = max(w, ansi.StringWidth(f))
	}
	return Model{spinner: s, width: w, id: frames.NextID()}
}

// WithStyles changes the styles derived from the theme through edit, under
// whatever theme View gets. It replaces an earlier edit.
func (m Model) WithStyles(edit func(Styles) Styles) Model {
	m.edit = edit
	return m
}

// Start turns the spinner. The command is its first frame; nil when it is
// already turning, has no frames, or motion is Still, where it holds its
// first frame while Running.
func (m Model) Start() (Model, tea.Cmd) {
	if m.running || len(m.spinner.Frames) == 0 {
		return m, nil
	}
	m.running = true
	m.tag++
	return m, m.Tick()
}

// Stop holds the spinner on its frame; a tick already on its way lands spent.
func (m Model) Stop() Model {
	m.running = false
	m.tag++
	return m
}

// Tick is the command for the next frame, Every apart at BEZEL_MOTION's
// speed; nil while stopped or Still.
func (m Model) Tick() tea.Cmd {
	if !m.running || len(m.spinner.Frames) == 0 || frames.Still() {
		return nil
	}
	return frames.Tick(time.Duration(float64(m.spinner.Every)/frames.Speed()), m.id, m.tag)
}

// FrameAt is the spinner's next frame for at, for a test or a host that
// delivers frames itself.
func (m Model) FrameAt(at time.Time) tea.Msg { return frames.NewMsg(at, m.id, m.tag) }

// Owns reports whether msg is one of this spinner's frames, a spent one
// included, so a host routing frames hands it here and nowhere else.
func (m Model) Owns(msg tea.Msg) bool { return frames.Of(msg, m.id) }

// Update turns to the next frame on its own frame and arms the one after.
// Any other message leaves it as it was, with no command.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if _, ok := frames.Is(msg, m.id, m.tag); !ok || !m.running || len(m.spinner.Frames) == 0 {
		return m, nil
	}
	m.frame = (m.frame + 1) % len(m.spinner.Frames)
	return m, m.Tick()
}

// Frame is the frame showing, unstyled and unpadded; "" without frames.
func (m Model) Frame() string {
	if len(m.spinner.Frames) == 0 {
		return ""
	}
	return m.spinner.Frames[m.frame]
}

// Spinner is the frames and pace the spinner turns with.
func (m Model) Spinner() Spinner {
	s := m.spinner
	s.Frames = slices.Clone(s.Frames)
	return s
}

// Running reports whether the spinner is turning.
func (m Model) Running() bool { return m.running }

// View is the frame in the theme's style, padded to the widest frame.
func (m Model) View(th theme.Resolved) string {
	f := m.Frame()
	if f == "" && m.width == 0 {
		return ""
	}
	st := DefaultStyles(th)
	if m.edit != nil {
		st = m.edit(st)
	}
	return st.Frame.Render(f) + strings.Repeat(" ", max(0, m.width-ansi.StringWidth(f)))
}
