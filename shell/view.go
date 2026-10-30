package shell

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/overlay"
)

// Body is the rect the panes share: the terminal minus header, tabs, status
// and legend. Pure in the shell's state, so Update and View agree on it.
func (s Shell) Body() layout.Rect {
	s = s.ready()
	top, bottom, _ := s.chromeRows()
	return layout.Rect{X: 0, Y: top, W: s.width, H: max(1, s.height-top-bottom)}
}

// chromeRows measures the chrome without rendering the panes.
func (s Shell) chromeRows() (top, bottom int, legendLines []string) {
	top = s.headerRows() + 1 // a blank row always sits above the panes
	if len(s.cfg.Tabs) > 0 {
		top++ // the tab strip, above that blank row
	}
	// helpKey names the key in the "+N in [?]" mark, only where Help() is
	// active and not while the input has the keys.
	helpKey := ""
	acts := s.active()
	if help := helpAction(acts); help != nil && !s.typing {
		helpKey = help.Entry().Help().Key
	}
	keys := legend.Filter(actionEntries(acts, !s.typing), s.cfg.Can) // a modal draws its own keys in its box
	legendLines = legend.Pack(keys, s.width, s.cfg.LegendLines, helpKey, s.cfg.Theme.Legend)
	// Every row the legend may take, filled or not: fewer keys never grow the panes.
	bottom = max(s.cfg.LegendLines, len(legendLines))
	if s.statusShown() {
		bottom++
	}
	return top, bottom, legendLines
}

// toastWidth is how wide the Toast style draws a message, edges included.
const toastWidth = 48

// statusShown reports whether the status row is drawn: always in the
// StatusBar style; in the Toast style while a line is typed on it, Busy runs
// or the tab's Status has text.
func (s Shell) statusShown() bool {
	switch {
	case s.cfg.StatusStyle != Toast, s.typing, s.statusLiner() != nil, s.busy != "":
		return true
	case len(s.cfg.Tabs) > 0:
		return s.cfg.Tabs[s.tab].Status(s.Context()) != ""
	}
	return false
}

func (s Shell) headerRows() int {
	if s.cfg.Header != nil {
		return len(s.cfg.Header(s.Context(), s.width))
	}
	return 1
}

// helpOverlay lists every key for where the user is, unpacked: the top
// overlay's when one is up, else the screen's. Help itself is the panel's key.
func (s Shell) helpOverlay() overlay.Overlay {
	keys := actionEntries(s.active(), false)
	if top := s.overlays.Top(); top != nil {
		keys = top.Legend()
	}
	var sections []overlay.HelpSection
	if len(keys) > 0 {
		sections = append(sections, overlay.HelpSection{Name: s.sectionName(), Entries: keys})
	}
	return overlay.NewHelp(sections, s.cfg.Theme)
}

// sectionName heads the keys of where the user is: the active tab, else the app.
func (s Shell) sectionName() string {
	if len(s.cfg.Tabs) > 0 {
		return s.cfg.Tabs[s.tab].Name()
	}
	if s.cfg.Title != "" {
		return s.cfg.Title
	}
	return "Keys"
}

// Pane is what the app hands the shell for one leaf: a title for the border
// and a body drawn into the inner rect the shell computed.
type Pane struct {
	Title  string
	Right  string        // at the end of the top edge, while it fits beside Title
	Bottom string        // at the end of the bottom edge, while it fits
	Scroll layout.Scroll // a scrollbar's thumb down the right edge, from list.Scroll and the like
	Accent color.Color   // the edges' colour, focused or not; nil keeps the theme's
	Body   func(inner layout.Rect) string
}

// withPresent re-resolves the layout for exactly these panes. View uses it so
// a leaf missing from panes gives its space back this frame.
func (s Shell) withPresent(panes map[string]Pane) Shell {
	s.present = make(map[string]bool, len(panes))
	for name := range panes {
		s.present[name] = true
	}
	return s.relayout()
}

// View draws header, tabs, panes, status and legend, then the overlays on
// top. It is pure: nothing here writes back to the shell the app holds.
func (s Shell) View(panes map[string]Pane) string {
	s = s.ready()
	if s.width <= 0 || s.height <= 0 {
		return ""
	}
	s = s.withPresent(panes)
	th := &s.cfg.Theme
	_, _, legendLines := s.chromeRows()
	body := s.Body()

	rows := make([]string, 0, s.height)
	rows = append(rows, s.headerLines()...)
	if len(s.cfg.Tabs) > 0 {
		rows = append(rows, draw.Tabs(s.width, s.tabNames(), s.tab, th.Chrome))
	}
	rows = append(rows, draw.Fit("", s.width))

	canvas := s.paneCanvas(panes, body)
	if s.trans.Active() {
		canvas = s.trans.Apply(canvas) // the tab strip above stays put
	}
	rows = append(rows, strings.Split(canvas, "\n")...)
	if s.statusShown() {
		rows = append(rows, s.statusRow())
	}
	for _, l := range legendLines {
		rows = append(rows, draw.Fit(l, s.width))
	}
	for range s.cfg.LegendLines - len(legendLines) {
		rows = append(rows, draw.Fit("", s.width))
	}

	out := draw.FitBlock(strings.Join(rows, "\n"), s.width, s.height)
	if s.overlays.Len() > 0 {
		out = s.overlays.View(body, out)
	}
	if s.toasts.Len() > 0 {
		// Narrower than the screen, so a toast never hangs off its left edge.
		out = s.toasts.WithWidth(min(toastWidth, s.width-2)).Over(out, *th)
	}
	return out
}

func (s Shell) headerLines() []string {
	if s.cfg.Header != nil {
		header := s.cfg.Header(s.Context(), s.width)
		lines := make([]string, len(header))
		for i, l := range header {
			lines[i] = draw.Fit(l, s.width)
		}
		return lines
	}
	return []string{draw.Header(s.width, s.cfg.Title, s.cfg.Subtitle, s.cfg.Version, s.cfg.Theme.Chrome)}
}

// paneCanvas draws every placed pane into a blank body-sized canvas.
func (s Shell) paneCanvas(panes map[string]Pane, body layout.Rect) string {
	blank := strings.Repeat(" ", body.W)
	lines := make([]string, body.H)
	for i := range lines {
		lines[i] = blank
	}
	canvas := strings.Join(lines, "\n")

	// Declaration order, not map order, so any overlap draws the same every frame.
	for _, name := range paneOrder(s.cfg.Layout) {
		r, placed := s.rects[name]
		p, ok := panes[name]
		if !placed || !ok || r.W <= 0 || r.H <= 0 {
			continue
		}
		st := s.cfg.Theme.Panel
		if name == s.Focused() {
			st = s.cfg.Theme.PanelFocused
		}
		if c := s.glowColor(name); c != nil {
			st.Border = st.Border.Foreground(c)
		}
		if p.Accent != nil {
			st.Border = st.Border.Foreground(p.Accent)
		}
		content := ""
		if p.Body != nil {
			content = p.Body(draw.InnerRect(r))
		}
		edges := draw.Edges{Title: p.Title, Right: p.Right, Bottom: p.Bottom, Scroll: p.Scroll}
		canvas = draw.Composite(draw.Panel(r, edges, content, st), canvas, r.X, r.Y-body.Y)
	}
	return canvas
}

// statusRow: the input outranks the banner, the banner outranks busy, and
// busy outranks the tab's own status text. The Toast style draws the banner
// as a toast instead.
func (s Shell) statusRow() string {
	th := &s.cfg.Theme
	switch {
	case s.statusLiner() != nil:
		return s.statusLiner().StatusLine(s.width)
	case s.typing:
		return draw.Fit(s.input.View(), s.width)
	case s.status != "" && s.cfg.StatusStyle != Toast:
		return draw.StatusLine(s.width, s.status, s.statusStyle())
	case s.busy != "":
		return draw.StatusLine(s.width, s.spin.View(*th)+" "+s.busy, th.Chrome.Status)
	case len(s.cfg.Tabs) > 0:
		return draw.StatusLine(s.width, s.cfg.Tabs[s.tab].Status(s.Context()), th.Chrome.Status)
	}
	return draw.StatusLine(s.width, "", th.Chrome.Status)
}

// statusStyle is how the status message shows at its level.
func (s Shell) statusStyle() lipgloss.Style {
	switch s.level {
	case OK:
		return s.cfg.Theme.StatusOK
	case Error:
		return s.cfg.Theme.StatusErr
	}
	return s.cfg.Theme.Chrome.Status
}

// statusLiner is the top overlay when it draws the status row itself.
func (s Shell) statusLiner() overlay.StatusLiner {
	sl, _ := s.overlays.Top().(overlay.StatusLiner)
	return sl
}

// tabNames is what the tab strip shows, one per tab.
func (s Shell) tabNames() []string {
	names := make([]string, len(s.cfg.Tabs))
	for i, t := range s.cfg.Tabs {
		names[i] = t.Name()
	}
	return names
}
