// Package shell is the drawer an app embeds: it owns the terminal size, tabs,
// focus, status, the overlay stack and the legend rule, and renders the app's
// panes into a declared layout. The app keeps everything that is its own.
package shell

import (
	"slices"
	"time"

	"github.com/lucasassuncao/bezel/input"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/spinner"
	"github.com/lucasassuncao/bezel/theme"
)

// Capability is legend.Capability, so an app need not import legend for it.
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
// times a frame, so keep them cheap and free of side effects. A Tab that is a
// value is read as given: hand it back through SetTabs when it changes.
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
	// LegendLines is how many rows the legend keeps, blank when unfilled, so the
	// panes hold their size as the keys change: 1 for a compact footer, 2 (the
	// default, also for 0 or less) when the app has many keys.
	LegendLines int
	// Actions are active on every screen, before the tab's or SetActions's.
	Actions []Action
	// StatusTTL is how long the palette's errors and Copy's result stay on the
	// status row; 5s when zero. An app with its own banner clock passes that.
	StatusTTL time.Duration
	// Spinner is what Busy shows; without frames it turns spinner.Dot's at its
	// own Every. A test sets an Every of a nanosecond, never to wait on a tick.
	Spinner spinner.Spinner
	// Clipboard is what Copy writes to; nil means the system clipboard.
	Clipboard func(string) error
}

// Shell is a value: every mutation returns the new shell.
type Shell struct {
	cfg *Config // shared by every copy and never written: see withCfg

	width, height int
	rects         map[string]layout.Rect
	collapsed     bool

	tab       int
	focus     []string // per tab, the focused leaf name
	ringFocus []string // per tab, the last focused leaf that tab stops at

	status    string
	level     Level
	statusSeq uint
	input     input.Model // the status row's line, while typing
	typing    bool

	busy string // what the spinner says, empty when idle
	spin spinner.Model

	overlays overlay.Stack
	present  map[string]bool // leaves the last View was given
	actions  []Action        // what the app set with SetActions
}

// New is a shell over cfg, focused on the first leaf of the focus ring. It
// draws nothing until the first tea.WindowSizeMsg.
func New(cfg Config) Shell {
	cfg.Tabs = withoutNil(cfg.Tabs)
	if cfg.LegendLines < 1 {
		cfg.LegendLines = 2
	}
	if cfg.StatusTTL <= 0 {
		cfg.StatusTTL = defaultStatusTTL
	}
	if len(cfg.Spinner.Frames) == 0 {
		cfg.Spinner.Frames = spinner.Dot().Frames
	}
	focus := resized(nil, max(1, len(cfg.Tabs)), firstLeaf(cfg.Layout))
	return Shell{
		cfg: &cfg, focus: focus, ringFocus: slices.Clone(focus), rects: map[string]layout.Rect{},
		spin: spinner.New(cfg.Spinner),
	}
}

// ready is s, or for the zero Shell the one New(Config{}) builds, so a host
// holding a Shell it has not built yet does not crash on it.
func (s Shell) ready() Shell {
	if s.cfg != nil {
		return s
	}
	return New(Config{})
}

// withCfg is s holding a changed copy of its Config. The one it held is
// shared with every other copy of the shell, so it is replaced, never written:
// a copy of the shell then costs a word, not the 29KB the theme in it takes.
func (s Shell) withCfg(change func(*Config)) Shell {
	c := *s.cfg
	change(&c)
	s.cfg = &c
	return s
}

// FocusMsg is sent when ChangePane moves focus, so the app can run what
// entering To and leaving From mean for it.
type FocusMsg struct{ From, To string }

// Size is the terminal's width and height, zero before the first resize.
func (s Shell) Size() (int, int) { return s.width, s.height }

// Rect is where a leaf was placed. The zero Rect means it is not on screen.
func (s Shell) Rect(name string) layout.Rect { return s.rects[name] }

// Focus is the name of the focused leaf on the active tab.
func (s Shell) Focus() string { s = s.ready(); return s.focus[s.tab] }

// SetFocus moves focus to a leaf of the layout. Unknown names are ignored,
// and a leaf the layout dropped hands focus to one on screen.
func (s Shell) SetFocus(name string) Shell {
	s = s.ready()
	if slices.Contains(layout.Order(s.cfg.Layout), name) {
		return s.focusOn(name).relayout() // a Header may size itself by focus
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

// ActiveTab is the index of the tab on screen.
func (s Shell) ActiveTab() int { return s.tab }

// SetTab shows tab i. An index out of range is ignored.
func (s Shell) SetTab(i int) Shell {
	s = s.ready()
	if i < 0 || i >= len(s.cfg.Tabs) {
		return s
	}
	s.tab = i
	return s.relayout() // its focus may be on a pane a resize dropped meanwhile
}

// SetTabs replaces the tabs, for an app whose Tab values carry its state. Each
// tab keeps its focus by position, a new one starts on the first pane, and a
// shorter list moves the shell to its last tab when the active one is gone.
func (s Shell) SetTabs(tabs ...Tab) Shell {
	s = s.ready()
	tabs = withoutNil(tabs)
	n, first := max(1, len(tabs)), firstLeaf(s.cfg.Layout)
	s = s.withCfg(func(c *Config) { c.Tabs = tabs })
	s.focus, s.ringFocus = resized(s.focus, n, first), resized(s.ringFocus, n, first)
	s.tab = min(s.tab, n-1)
	return s.relayout()
}

// Narrow reports whether a Collapse fired in the current layout.
func (s Shell) Narrow() bool { return s.collapsed }

// SetLayout swaps the layout tree and re-resolves it. For screens whose
// shape changes at runtime, such as a panel that animates open.
func (s Shell) SetLayout(n layout.Node) Shell {
	s = s.ready()
	s = s.withCfg(func(c *Config) { c.Layout = n })
	s.present = nil // the old frame's panes say nothing about the new tree
	return s.relayout()
}

// SetTheme swaps the styles, for a terminal that answered light or dark
// after the shell was built.
func (s Shell) SetTheme(th theme.Resolved) Shell {
	s = s.ready()
	s = s.withCfg(func(c *Config) { c.Theme = th })
	return s.relayout()
}

// SetSubtitle changes the text after the title in the default header.
func (s Shell) SetSubtitle(text string) Shell {
	s = s.ready()
	s = s.withCfg(func(c *Config) { c.Subtitle = text })
	return s
}

// SetActions is the current state's own actions, joining the global ones and
// the tab's. Variadic.
func (s Shell) SetActions(actions ...Action) Shell {
	s = s.ready()
	s.actions = slices.Clone(actions)
	return s.relayout()
}

// SetLegendLines changes how many rows the legend keeps, 1 or more.
func (s Shell) SetLegendLines(n int) Shell {
	s = s.ready()
	s = s.withCfg(func(c *Config) { c.LegendLines = max(1, n) })
	return s.relayout()
}

// Status is the banner text currently set, empty when none.
func (s Shell) Status() string { return s.status }

// ShowHelp opens the Help overlay the ? key would.
func (s Shell) ShowHelp() Shell { s = s.ready(); return s.Push(s.helpOverlay()).relayout() }

// Context is what the shell hands tabs and actions about where the user is.
func (s Shell) Context() Context {
	s = s.ready()
	return Context{Width: s.width, Height: s.height, Focus: s.Focus(), Narrow: s.collapsed, Can: s.cfg.Can}
}

// HasOverlay reports whether a modal is open.
func (s Shell) HasOverlay() bool { return s.overlays.Len() > 0 }

// TopOverlay is the modal that gets the keys, or nil.
func (s Shell) TopOverlay() overlay.Overlay { return s.overlays.Top() }

// Push opens o above any open modal.
func (s Shell) Push(o overlay.Overlay) Shell {
	s = s.ready()
	s.overlays = s.overlays.Push(o)
	return s
}

// Pop closes the top modal, if any.
func (s Shell) Pop() Shell {
	s = s.ready()
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

// relayout re-resolves the layout for the current size and legend, moves
// focus off a leaf that vanished, and gives the overlays the new body.
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
	// A Header sized by focus or Narrow moves the body as they settle: resolve
	// until it holds still, a few rounds at most for one that never does.
	for range 3 {
		body := s.Body()
		s.overlays = s.overlays.Resize(body)
		s.rects, s.collapsed = layout.Resolve(s.cfg.Layout, body, present)
		if _, ok := s.rects[s.Focus()]; !ok {
			if order := s.visibleOrder(); len(order) > 0 {
				s = s.focusOn(order[0])
			}
		}
		if s.Body() == body {
			break
		}
	}
	return s
}

// firstLeaf is where focus starts: the first leaf of the focus ring.
func firstLeaf(n layout.Node) string {
	if ring := layout.Ring(n); len(ring) > 0 {
		return ring[0]
	}
	return ""
}

// resized is xs cut or grown to n entries, the new ones set to fill.
func resized(xs []string, n int, fill string) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fill
	}
	copy(out, xs)
	return out
}

func withAt(xs []string, i int, v string) []string {
	out := make([]string, len(xs))
	copy(out, xs)
	out[i] = v
	return out
}
