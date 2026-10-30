package main

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/effect"
	"github.com/lucasassuncao/bezel/theme"
)

type (
	retroMsg   struct{} // e on the retro tab: the next effect
	rewindMsg  struct{} // r on the retro tab: the same again
	loadingMsg struct{} // l: the panel is loading, or done
)

// retroKinds are the effects the card plays through, one at a time.
var retroKinds = []struct {
	name string
	make func(th theme.Resolved) effect.Effect
}{
	{"glitch", func(theme.Resolved) effect.Effect { return effect.Glitch(700 * time.Millisecond) }},
	{"power off", func(th theme.Resolved) effect.Effect { return effect.PowerOff(th.Bold.GetForeground(), time.Second) }},
	{"power on", func(th theme.Resolved) effect.Effect { return effect.PowerOn(th.Bold.GetForeground(), time.Second) }},
	{"scanlines", func(theme.Resolved) effect.Effect { return effect.Scanlines(black, play).WithYoyo(true).WithRepeat(1) }},
	{"pixelate", func(theme.Resolved) effect.Effect { return effect.Pixelate(time.Second) }},
	{"iris in", func(theme.Resolved) effect.Effect { return effect.IrisIn(play) }},
	{"iris out", func(theme.Resolved) effect.Effect { return effect.IrisOut(play).WithYoyo(true).WithRepeat(1) }},
}

// retros is the fourth tab: one effect at a time on a card, and a panel
// that is loading, shown as a skeleton with a light running round its edge.
type retros struct {
	kind    int
	effect  effect.Effect
	loading bool
	trace   effect.Effect
	shape   effect.Effect
}

func newRetros(th theme.Resolved) retros { return retros{loading: true}.again(th).load(th) }

func (r retros) again(th theme.Resolved) retros {
	r.effect = retroKinds[r.kind].make(th)
	return r
}

func (r retros) next(th theme.Resolved) retros {
	r.kind = (r.kind + 1) % len(retroKinds)
	return r.again(th)
}

// load starts the skeleton and the trace, or settles them when loading ends.
func (r retros) load(th theme.Resolved) retros {
	r.trace = effect.Trace(th.Accent.GetForeground(), 1500*time.Millisecond)
	r.shape = effect.Skeleton(th.Muted.GetForeground(), th.Bold.GetForeground(), time.Second)
	return r
}

func (r retros) toggleLoading(th theme.Resolved) (retros, tea.Cmd) {
	r.loading = !r.loading
	r = r.load(th)
	return r, r.ticks()
}

// start sets the loading panel going on its own frames.
func (r retros) start() (retros, tea.Cmd) { return r, r.ticks() }

// ticks is the loading panel's first frames, none once it is done.
func (r retros) ticks() tea.Cmd {
	if !r.loading {
		return nil
	}
	return tea.Batch(r.trace.Tick(), r.shape.Tick())
}

// update takes the skeleton's and the trace's own frames.
func (r retros) update(msg tea.Msg) (retros, tea.Cmd) {
	var a, b bezel.Result
	r.trace, a = r.trace.Update(msg)
	r.shape, b = r.shape.Update(msg)
	return r, tea.Batch(a.Cmd, b.Cmd)
}

func (r retros) advance(at time.Time) retros {
	r.effect, _ = r.effect.Advance(at)
	return r
}

func (r retros) moving() bool { return r.effect.Active() }

// rows is the tab's body: the effect over its card, then the loading panel.
func (r retros) rows(th theme.Resolved, label func(kind, variant string) string) string {
	rows := []string{label("retro", retroKinds[r.kind].name) + th.Muted.Render("e next · r again"), ""}
	indent := label("", "")
	for _, line := range strings.Split(r.effect.Apply(retroCard(th)), "\n") {
		rows = append(rows, indent+line)
	}
	rows = append(rows, "")
	panel := report(th)
	if r.loading {
		panel = r.trace.Apply(r.shape.Within(2, 1, 26, 3).Apply(panel))
	}
	for i, line := range strings.Split(panel, "\n") {
		kind, variant := "", ""
		switch i {
		case 0:
			kind, variant = "skeleton", "and a trace"
		case 1:
			variant = "l loads"
		}
		rows = append(rows, label(kind, variant)+line)
	}
	return strings.Join(rows, "\n")
}

// retroCard is what the retro effects play on.
func retroCard(th theme.Resolved) string {
	body := strings.Join([]string{
		th.Accent.Bold(true).Render("bezel ✦ retro"),
		"Old screens had character:",
		th.Success.Render("> READY."),
	}, "\n")
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(th.Accent.GetForeground()).
		Padding(0, 1).Render(body)
}

// report is the panel that loads: its rows are bars while it does.
func report(th theme.Resolved) string {
	body := strings.Join([]string{
		th.Bold.Render("Weekly report"),
		"requests  1,204,331",
		"errors    0.02%",
	}, "\n")
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(th.Muted.GetForeground()).
		Padding(0, 1).Width(30).Render(body)
}
