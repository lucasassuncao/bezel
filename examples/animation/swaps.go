package main

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/effect"
	"github.com/lucasassuncao/bezel/theme"
)

// swapTime is how long one transition takes.
const swapTime = 900 * time.Millisecond

type (
	swapMsg   struct{} // n on the swap tab: the next transition
	againSwap struct{} // r on the swap tab: back the other way
)

// swapKinds are the transitions the two cards swap with.
var swapKinds = []struct {
	name string
	make func() effect.Transition
}{
	{"push ←", func() effect.Transition { return effect.Push(effect.Left, swapTime) }},
	{"cover ↑", func() effect.Transition { return effect.Cover(effect.Up, swapTime) }},
	{"reveal →", func() effect.Transition { return effect.Reveal(effect.Right, swapTime) }},
	{"crossfade", func() effect.Transition { return effect.CrossFade(swapTime) }},
	{"wipe radial", func() effect.Transition { return effect.Wipe(effect.Radial(), swapTime) }},
	{"wipe diagonal", func() effect.Transition { return effect.Wipe(effect.Diagonal(), swapTime) }},
	{"zoom", func() effect.Transition { return effect.Zoom(swapTime) }},
	{"shuffle", func() effect.Transition { return effect.Shuffle(swapTime) }},
}

// swaps is the eighth tab: two cards, one carried into the other.
type swaps struct {
	kind    int
	t       effect.Transition
	flipped bool // the settings card is in front
}

func newSwaps() swaps { return swaps{t: effect.Push(effect.Left, 0)} }

// again swaps back the way it came, with the same transition.
func (s swaps) again() swaps {
	s.flipped = !s.flipped
	s.t = swapKinds[s.kind].make()
	return s
}

func (s swaps) next() swaps {
	s.kind = (s.kind + 1) % len(swapKinds)
	return s.again()
}

func (s swaps) advance(at time.Time) swaps {
	s.t, _ = s.t.Advance(at)
	return s
}

func (s swaps) moving() bool { return s.t.Active() }

func (s swaps) rows(th theme.Resolved, label func(kind, variant string) string) string {
	rows := []string{label("transition", swapKinds[s.kind].name) + th.Muted.Render("n next · r again"), ""}
	from, to := settingsCard(th), inboxCard(th)
	if s.flipped {
		from, to = to, from
	}
	indent := label("", "")
	for _, line := range strings.Split(s.t.Apply(from, to), "\n") {
		rows = append(rows, indent+line)
	}
	return strings.Join(rows, "\n")
}

// swapCard is a bordered card of one size, so the two swap cell for cell.
func swapCard(border lipgloss.Style, lines ...string) string {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border.GetForeground()).
		Width(32).Height(len(lines)).Padding(0, 1).Render(strings.Join(lines, "\n"))
}

func inboxCard(th theme.Resolved) string {
	return swapCard(th.Accent,
		th.Accent.Bold(true).Render("Inbox"),
		"● Build passed on main",
		"● 3 new comments",
		th.Muted.Render("○ Weekly digest"),
		th.Muted.Render("○ Release notes"),
	)
}

func settingsCard(th theme.Resolved) string {
	return swapCard(th.Info,
		th.Info.Bold(true).Render("Settings"),
		"Theme        "+th.Bold.Render("mint"),
		"Motion       "+th.Bold.Render("on"),
		"Confirm save "+th.Bold.Render("yes"),
		th.Muted.Render("esc closes"),
	)
}
