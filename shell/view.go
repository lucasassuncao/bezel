package shell

import (
	"strings"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/overlay"
)

// Body is the rect the panes share: the terminal minus header, tabs, status
// and legend. Pure in the shell's state, so Update and View agree on it.
func (s Shell) Body() layout.Rect {
	top, bottom, _ := s.chromeRows()
	return layout.Rect{X: 0, Y: top, W: s.width, H: max(1, s.height-top-bottom)}
}

// chromeRows measures the chrome without rendering the panes.
func (s Shell) chromeRows() (top, bottom int, legendLines []string) {
	top = s.headerRows() + 1 // a blank row always sits above the panes
	if len(s.cfg.Tabs) > 0 {
		top++ // the tab strip, above that blank row
	}
	// HelpKey names the key in the "+N in [?]" mark, only where Help() is
	// active and not while the input has the keys.
	st := s.cfg.Theme.Legend
	st.HelpKey = ""
	if help := s.helpAction(); help != nil && s.input == nil {
		st.HelpKey = help.Entry().Help().Key
	}
	legendLines = legend.Pack(s.legendEntries(), s.width, s.cfg.LegendLines, st)
	bottom = 1 + len(legendLines)
	return top, bottom, legendLines
}

func (s Shell) headerRows() int {
	if s.cfg.Header != nil {
		return len(s.cfg.Header(s.Context(), s.width))
	}
	return 1
}

// legendEntries is the screen's keys, filtered by Can. An overlay never
// replaces them: a modal draws its own keys inside its box.
func (s Shell) legendEntries() []legend.Entry {
	return legend.Filter(s.actionEntries(s.input == nil), s.cfg.Can)
}

// helpOverlay lists every key for where the user is, unpacked: the top
// overlay's when one is up, else the screen's. Help itself is the panel's key.
func (s Shell) helpOverlay() overlay.Overlay {
	keys := s.actionEntries(false)
	if top := s.overlays.Top(); top != nil {
		keys = top.Legend()
	}
	var sections []overlay.HelpSection
	if len(keys) > 0 {
		sections = append(sections, overlay.HelpSection{Name: s.sectionName(), Entries: keys})
	}
	return overlay.NewHelp("", sections, s.cfg.Theme.Modal, s.cfg.Theme.Legend)
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
	Title string
	Body  func(inner layout.Rect) string
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
	if s.width <= 0 || s.height <= 0 {
		return ""
	}
	s = s.withPresent(panes)
	th := s.cfg.Theme
	_, _, legendLines := s.chromeRows()
	body := s.Body()

	rows := make([]string, 0, s.height)
	rows = append(rows, s.headerLines()...)
	if len(s.cfg.Tabs) > 0 {
		names := make([]string, len(s.cfg.Tabs))
		for i, t := range s.cfg.Tabs {
			names[i] = t.Name()
		}
		rows = append(rows, draw.Tabs(s.width, names, s.tab, th.Chrome))
	}
	rows = append(rows, draw.Fit("", s.width))

	rows = append(rows, strings.Split(s.paneCanvas(panes, body), "\n")...)
	rows = append(rows, s.statusRow())
	for _, l := range legendLines {
		rows = append(rows, draw.Fit(l, s.width))
	}

	out := draw.FitBlock(strings.Join(rows, "\n"), s.width, s.height)
	if s.overlays.Len() > 0 {
		out = s.overlays.View(body, out)
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
	for _, name := range layout.Order(s.cfg.Layout) {
		r, placed := s.rects[name]
		p, ok := panes[name]
		if !placed || !ok || r.W <= 0 || r.H <= 0 {
			continue
		}
		st := s.cfg.Theme.Panel
		if name == s.Focus() {
			st = s.cfg.Theme.PanelFocused
		}
		content := ""
		if p.Body != nil {
			content = p.Body(draw.InnerRect(r))
		}
		canvas = draw.Composite(draw.Panel(r, p.Title, content, st), canvas, r.X, r.Y-body.Y)
	}
	return canvas
}

// statusRow: the input outranks the banner, the banner outranks busy, and
// busy outranks the tab's own status text.
func (s Shell) statusRow() string {
	th := s.cfg.Theme
	switch {
	case s.statusLiner() != nil:
		return s.statusLiner().StatusLine(s.width)
	case s.input != nil:
		return draw.Fit(s.input.View(), s.width)
	case s.status != "" && s.level == OK:
		return draw.StatusLine(s.width, s.status, th.StatusOK)
	case s.status != "" && s.level == Error:
		return draw.StatusLine(s.width, s.status, th.StatusErr)
	case s.status != "":
		return draw.StatusLine(s.width, s.status, th.Chrome.Status)
	case s.busy != "":
		return draw.StatusLine(s.width, th.Accent.Render(s.spin.View())+" "+s.busy, th.Chrome.Status)
	case len(s.cfg.Tabs) > 0:
		return draw.StatusLine(s.width, s.cfg.Tabs[s.tab].Status(s.Context()), th.Chrome.Status)
	}
	return draw.StatusLine(s.width, "", th.Chrome.Status)
}

// statusLiner is the top overlay when it draws the status row itself.
func (s Shell) statusLiner() overlay.StatusLiner {
	sl, _ := s.overlays.Top().(overlay.StatusLiner)
	return sl
}
