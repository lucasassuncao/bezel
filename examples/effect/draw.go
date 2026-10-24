package main

import (
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/effect"
	"github.com/lucasassuncao/bezel/theme"
)

// play is how long one effect takes on the card.
const play = 1200 * time.Millisecond

type (
	effectMsg  struct{} // e: the next effect
	patternMsg struct{} // p: the next pattern
	againMsg   struct{} // r on the draw tab: the same again
)

// black is the gallery's dark background, what fades start from and end in.
var black = color.RGBA{0, 0, 0, 255}

// effectKinds are the effects the card cycles through. The ones that take the
// card away come back the same way, so it is never left empty.
var effectKinds = []struct {
	name string
	make func(th theme.Resolved) effect.Effect
}{
	{"fade in", func(theme.Resolved) effect.Effect { return effect.FadeIn(black, play) }},
	{"fade out", func(theme.Resolved) effect.Effect { return effect.FadeOut(black, play).WithYoyo().WithRepeat(1) }},
	{"dissolve", func(theme.Resolved) effect.Effect { return effect.Dissolve(play).WithYoyo().WithRepeat(1) }},
	{"coalesce", func(theme.Resolved) effect.Effect { return effect.Coalesce(play) }},
	{"sweep in", func(th theme.Resolved) effect.Effect { return effect.SweepIn(th.Accent.GetForeground(), play) }},
	{"slide in", func(theme.Resolved) effect.Effect { return effect.SlideIn(effect.Right, play) }},
	{"slide out", func(theme.Resolved) effect.Effect {
		return effect.SlideOut(effect.Up, play).WithYoyo().WithRepeat(1)
	}},
	{"evolve", func(theme.Resolved) effect.Effect { return effect.Evolve(play) }},
	{"explode", func(theme.Resolved) effect.Effect { return effect.Explode(play).WithYoyo().WithRepeat(1) }},
	{"expand", func(th theme.Resolved) effect.Effect { return effect.Expand(th.Accent.GetForeground(), play) }},
	{"stretch", func(th theme.Resolved) effect.Effect {
		return effect.Stretch(effect.Right, th.Accent.GetForeground(), play)
	}},
	{"hue shift", func(theme.Resolved) effect.Effect { return effect.HueShift(180, play).WithYoyo().WithRepeat(1) }},
	{"desaturate", func(theme.Resolved) effect.Effect { return effect.Saturate(-1, play).WithYoyo().WithRepeat(1) }},
}

// patterns are the orders the effect can reach the card's cells in; the
// first leaves the effect its own.
var patterns = []struct {
	name    string
	pattern effect.Pattern
}{
	{"its own", nil},
	{"sweep →", effect.Sweep(effect.Right)},
	{"radial", effect.Radial()},
	{"diamond", effect.Diamond()},
	{"diagonal", effect.Diagonal()},
	{"checker", effect.Checker()},
	{"spiral", effect.Spiral(2)},
	{"wave", effect.Wave()},
	{"noise", effect.Noise(9)},
}

// drawing is the first tab: one effect at a time on a card drawn by lipgloss.
type drawing struct {
	kind, order int
	effect      effect.Effect
}

// again plays the current effect, in the current order, from the start.
func (d drawing) again(th theme.Resolved) drawing {
	d.effect = effectKinds[d.kind].make(th)
	if p := patterns[d.order]; p.pattern != nil {
		d.effect = d.effect.WithPattern(p.pattern)
	}
	return d
}

func (d drawing) nextEffect(th theme.Resolved) drawing {
	d.kind = (d.kind + 1) % len(effectKinds)
	return d.again(th)
}

func (d drawing) nextPattern(th theme.Resolved) drawing {
	d.order = (d.order + 1) % len(patterns)
	return d.again(th)
}

func (d drawing) advance(at time.Time) drawing {
	d.effect, _ = d.effect.Advance(at)
	return d
}

func (d drawing) moving() bool { return d.effect.Active() }

// rows is the tab's body: what plays, then the card as the effect shows it.
func (d drawing) rows(th theme.Resolved, label func(kind, variant string) string) string {
	order := patterns[d.order].name
	switch effectKinds[d.kind].name {
	case "slide in", "slide out", "expand", "stretch":
		order = "none: it moves whole"
	}
	rows := []string{
		label("effect", effectKinds[d.kind].name) + th.Muted.Render("e next · r again"),
		label("pattern", "") + th.Bold.Render(order) + th.Muted.Render("  p next"),
		"",
	}
	indent := label("", "")
	for _, line := range strings.Split(d.effect.Apply(card(th)), "\n") {
		rows = append(rows, indent+line)
	}
	return strings.Join(rows, "\n")
}

// card is what the effects play on: a bordered box of text in a few colours.
func card(th theme.Resolved) string {
	body := strings.Join([]string{
		th.Accent.Bold(true).Render("bezel ✦ effects"),
		"",
		"Every cell of this card is",
		"read back from its escape",
		"codes and drawn again.",
		"",
		th.Success.Render("● ") + th.Info.Render("13 effects · 9 orders"),
	}, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(th.Accent.GetForeground()).
		Padding(0, 1).
		Render(body)
}
