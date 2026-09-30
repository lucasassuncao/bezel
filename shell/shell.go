// Package shell is the drawer an app embeds: it owns the terminal size, tabs,
// focus, status, the overlay stack and the legend rule, and renders the app's
// panes into a declared layout. The app keeps everything that is its own.
package shell

import (
	"slices"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/theme"
)

type Capability = legend.Capability

// Level is how a status message is drawn: neutral, a success, or an error.
type Level int

const (
	Info Level = iota
	OK
	Error
)

// Context is what a Tab gets when asked for its legend: enough to change
// its wording by width, focus and what the session allows.
type Context struct {
	Width, Height int
	Focus         string
	Narrow        bool
	Can           func(Capability) bool
}

// Tab answers for one screen of the app: its name in the strip, the status
// text and the keys for where the user is. Status and Actions run several
// times a frame, so keep them cheap and free of side effects.
type Tab interface {
	Name() string
	Status(ctx Context) string
	Actions(ctx Context) []Action
}

// Config is everything a Shell needs up front.
type Config struct {
	Layout   layout.Node
	Tabs     []Tab // nil for a single-screen app
	Theme    theme.Resolved
	Title    string
	Subtitle string
	Version  string
	// Header replaces the default one-row header, one string per row. It runs
	// several times a frame, like Tab.Status.
	Header func(ctx Context, width int) []string
	// Can decides which legend entries survive. nil allows everything.
	Can func(Capability) bool
	// LegendLines is how many rows the legend may take: 1 for a compact
	// footer, 2 (the default, also for 0 or less) when the app has many keys.
	LegendLines int
	// Actions are active on every screen, before the tab's or SetActions's.
	Actions []Action
	// StatusTTL is how long the palette's errors and Copy's result stay on the
	// status row; 5s when zero. An app with its own banner clock passes that.
	StatusTTL time.Duration
	// Clipboard is what Copy writes to; nil means the system clipboard.
	Clipboard func(string) error
}

// Shell is a value: every mutation returns the new shell.
type Shell struct {
	cfg Config

	width, height int
	rects         map[string]layout.Rect
	collapsed     bool

	tab       int
	focus     []string // per tab, the focused leaf name
	ringFocus []string // per tab, the last focused leaf that tab stops at

	status    string
	level     Level
	statusSeq uint
	input     *textinput.Model

	busy string // what the spinner says, empty when idle
	spin spinner.Model

	overlays overlay.Stack
	present  map[string]bool // leaves the last View was given
	actions  []Action        // what the app set with SetActions
}

func New(cfg Config) Shell {
	if cfg.LegendLines < 1 {
		cfg.LegendLines = 2
	}
	if cfg.StatusTTL <= 0 {
		cfg.StatusTTL = defaultStatusTTL
	}
	tabs := max(1, len(cfg.Tabs))
	first := ""
	if ring := layout.Ring(cfg.Layout); len(ring) > 0 {
		first = ring[0]
	}
	focus := make([]string, tabs)
	for i := range focus {
		focus[i] = first
	}
	return Shell{
		cfg: cfg, focus: focus, ringFocus: slices.Clone(focus), rects: map[string]layout.Rect{},
		spin: spinner.New(spinner.WithSpinner(spinner.Dot)),
	}
}

// FocusMsg is sent when ChangePane moves focus, so the app can run what
// entering To and leaving From mean for it.
type FocusMsg struct{ From, To string }

func (s Shell) Size() (int, int) { return s.width, s.height }

// Rect is where a leaf was placed. The zero Rect means it is not on screen.
func (s Shell) Rect(name string) layout.Rect { return s.rects[name] }

func (s Shell) Focus() string { return s.focus[s.tab] }

// SetFocus moves focus to a leaf of the layout. Unknown names are ignored.
func (s Shell) SetFocus(name string) Shell {
	if slices.Contains(layout.Order(s.cfg.Layout), name) {
		return s.focusOn(name)
	}
	return s
}

// focusOn moves focus, remembering it as the way back when name is on the ring.
func (s Shell) focusOn(name string) Shell {
	s.focus = withAt(s.focus, s.tab, name)
	if slices.Contains(layout.Ring(s.cfg.Layout), name) {
		s.ringFocus = withAt(s.ringFocus, s.tab, name)
	}
	return s
}

func (s Shell) ActiveTab() int { return s.tab }

func (s Shell) SetTab(i int) Shell {
	if i < 0 || i >= len(s.cfg.Tabs) {
		return s
	}
	s.tab = i
	return s.relayout() // the tab's legend may take a different number of rows
}

func (s Shell) Narrow() bool { return s.collapsed }

// SetLayout swaps the layout tree and re-resolves it. For screens whose
// shape changes at runtime, such as a panel that animates open.
func (s Shell) SetLayout(n layout.Node) Shell {
	s.cfg.Layout = n
	s.present = nil // the old frame's panes say nothing about the new tree
	return s.relayout()
}

// SetTheme swaps the styles, for a terminal that answered light or dark
// after the shell was built.
func (s Shell) SetTheme(th theme.Resolved) Shell {
	s.cfg.Theme = th
	return s.relayout()
}

func (s Shell) SetSubtitle(text string) Shell {
	s.cfg.Subtitle = text
	return s
}

// SetActions is the current state's own actions, joining the global ones and
// the tab's. Variadic; the legend rows count against the body.
func (s Shell) SetActions(actions ...Action) Shell {
	s.actions = slices.Clone(actions)
	return s.relayout()
}

// SetLegendLines changes how many rows the legend may take, 1 or more.
func (s Shell) SetLegendLines(n int) Shell {
	s.cfg.LegendLines = max(1, n)
	return s.relayout()
}

// Status is the banner text currently set, empty when none.
func (s Shell) Status() string { return s.status }

// ShowHelp opens the Help overlay the ? key would.
func (s Shell) ShowHelp() Shell { return s.Push(s.helpOverlay()).relayout() }

func (s Shell) Context() Context {
	return Context{Width: s.width, Height: s.height, Focus: s.Focus(), Narrow: s.collapsed, Can: s.cfg.Can}
}

func (s Shell) HasOverlay() bool            { return s.overlays.Len() > 0 }
func (s Shell) TopOverlay() overlay.Overlay { return s.overlays.Top() }

func (s Shell) Push(o overlay.Overlay) Shell {
	s.overlays = s.overlays.Push(o)
	return s
}

func (s Shell) Pop() Shell {
	s.overlays = s.overlays.Pop()
	return s
}

// nextPane moves focus by step through the leaves currently on screen. From
// an Info pane it goes back to the ring pane that had focus before.
func (s Shell) nextPane(step int) Shell {
	order := s.visibleOrder()
	if len(order) == 0 {
		return s
	}
	cur := slices.Index(order, s.Focus())
	if cur < 0 {
		if back := s.ringFocus[s.tab]; slices.Contains(order, back) {
			return s.focusOn(back)
		}
		return s.focusOn(order[0])
	}
	return s.focusOn(order[((cur+step)%len(order)+len(order))%len(order)])
}

// visibleOrder is the focus ring minus the leaves the layout dropped. Before
// the first resize nothing is placed yet, so every leaf counts.
func (s Shell) visibleOrder() []string {
	var out []string
	for _, n := range layout.Ring(s.cfg.Layout) {
		if _, ok := s.rects[n]; ok || len(s.rects) == 0 {
			out = append(out, n)
		}
	}
	return out
}

// relayout re-resolves the layout for the current size and legend, and moves
// focus off a leaf that vanished.
func (s Shell) relayout() Shell {
	if s.width <= 0 || s.height <= 0 {
		return s // not sized yet: a zero body would collapse and steal focus
	}
	present := func(name string) bool {
		if s.present == nil {
			return true
		}
		return s.present[name]
	}
	s.rects, s.collapsed = layout.Resolve(s.cfg.Layout, s.Body(), present)
	if _, ok := s.rects[s.Focus()]; !ok {
		if order := s.visibleOrder(); len(order) > 0 {
			s = s.focusOn(order[0])
		}
	}
	return s
}

func withAt(xs []string, i int, v string) []string {
	out := make([]string, len(xs))
	copy(out, xs)
	out[i] = v
	return out
}
