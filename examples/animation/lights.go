package main

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/theme"
)

type (
	lightMsg struct{} // e on the light tab: the next light
	shineMsg struct{} // r on the light tab: the same again
	diffMsg  struct{} // f: the diff plays
	undoMsg  struct{} // u: undo restores the line
)

// lightKinds are the lights and presets the card cycles through.
var lightKinds = []struct {
	name string
	make func(th theme.Resolved) animation.Effect
}{
	{"tint", func(th theme.Resolved) animation.Effect { return animation.Tint(th.Success.GetForeground(), play) }},
	{"pulse", func(th theme.Resolved) animation.Effect {
		return animation.Pulse(th.Accent.GetForeground(), 800*time.Millisecond).WithRepeat(2)
	}},
	{"press", func(theme.Resolved) animation.Effect { return animation.Press(400 * time.Millisecond) }},
	{"flash", func(th theme.Resolved) animation.Effect { return animation.Flash(th.Warning.GetForeground(), play) }},
	{"invert", func(theme.Resolved) animation.Effect { return animation.Invert(700 * time.Millisecond) }},
	{"spotlight", func(theme.Resolved) animation.Effect {
		return animation.Spotlight(2, 1, 26, 1, black, 600*time.Millisecond).WithYoyo().WithDelay(0).WithRepeat(1)
	}},
	{"vignette", func(theme.Resolved) animation.Effect { return animation.Vignette(black, play).WithYoyo().WithRepeat(1) }},
	{"contrast", func(theme.Resolved) animation.Effect { return animation.Contrast(0.2, play).WithYoyo().WithRepeat(1) }},
	{"hover", func(th theme.Resolved) animation.Effect {
		return animation.Hover(black, animation.Blend(black, th.Accent.GetForeground(), 0.3), 300*time.Millisecond).
			WithYoyo().WithRepeat(1).WithDelay(0)
	}},
	{"melt", func(theme.Resolved) animation.Effect { return animation.Melt(nil, play).WithYoyo().WithRepeat(1) }},
	{"enter", func(theme.Resolved) animation.Effect { return animation.Enter(black) }},
	{"exit", func(theme.Resolved) animation.Effect { return animation.Exit(black).WithYoyo().WithRepeat(1) }},
	{"attention", func(th theme.Resolved) animation.Effect { return animation.Attention(th.Accent.GetForeground()) }},
	{"success", func(th theme.Resolved) animation.Effect { return animation.Success(th.Success.GetForeground()) }},
	{"failure", func(th theme.Resolved) animation.Effect { return animation.Failure(th.Danger.GetForeground()) }},
}

// lights is the sixth tab: one light at a time on a card, a diff whose lines
// come and go, and the ghost of a dragged file.
type lights struct {
	kind     int
	effect   animation.Effect
	removed  animation.Effect // the diff's line going out
	added    animation.Effect // the diff's line coming in
	restored animation.Effect // undo's trail over the line it brought back
	undone   bool
}

// newLights is the card's first light, and a diff at rest: its effects
// already over, so the rows read as they are.
func newLights(th theme.Resolved) lights {
	quiet := func(row int) animation.Effect { return animation.Flash(black, 0).OnRows(row) }
	l := lights{removed: quiet(1), added: quiet(2), restored: quiet(1)}
	return l.again(th)
}

func (l lights) again(th theme.Resolved) lights {
	l.effect = lightKinds[l.kind].make(th)
	return l
}

func (l lights) next(th theme.Resolved) lights {
	l.kind = (l.kind + 1) % len(lightKinds)
	return l.again(th)
}

// diff plays the change: the old line melts red, the new one flashes green.
func (l lights) diff(th theme.Resolved) lights {
	l.undone = false
	l.removed = animation.Melt(th.Danger.GetForeground(), play).OnRows(1)
	l.added = animation.Flash(th.Success.GetForeground(), play).OnRows(2)
	return l
}

// undo brings the old line back, a soft light over it for a moment.
func (l lights) undo(th theme.Resolved) lights {
	l.undone = true
	l.restored = animation.Flash(th.Accent.GetForeground(), play).WithStrength(0.5).OnRows(1)
	return l
}

func (l lights) advance(at time.Time) lights {
	l.effect, _ = l.effect.Advance(at)
	l.removed, _ = l.removed.Advance(at)
	l.added, _ = l.added.Advance(at)
	l.restored, _ = l.restored.Advance(at)
	return l
}

func (l lights) moving() bool {
	return l.effect.Active() || l.removed.Active() || l.added.Active() || l.restored.Active()
}

// rows is the tab's body: the light and its card, the diff, the ghost.
func (l lights) rows(th theme.Resolved, label func(kind, variant string) string) string {
	rows := []string{label("light", lightKinds[l.kind].name) + th.Muted.Render("e next · r again"), ""}
	indent := label("", "")
	for _, line := range strings.Split(l.effect.Apply(lightCard(th)), "\n") {
		rows = append(rows, indent+line)
	}
	rows = append(rows, "")
	code := l.diffBlock(th)
	for i, line := range strings.Split(code, "\n") {
		kind, variant := "", ""
		switch i {
		case 0:
			kind, variant = "diff", "f plays"
		case 1:
			variant = "u undoes"
		}
		rows = append(rows, label(kind, variant)+line)
	}
	file := th.Info.Render("▤ report.pdf")
	rows = append(rows, label("ghost", "dragged")+file+"   "+animation.Ghost(file, black, 0.35))
	return strings.Join(rows, "\n")
}

// diffBlock is the change as a unified diff, or the code put back by undo.
func (l lights) diffBlock(th theme.Resolved) string {
	if l.undone {
		code := strings.Join([]string{
			"  func sum(a, b int) int {",
			"      return a - b",
			"  }",
		}, "\n")
		return l.restored.Apply(code)
	}
	code := strings.Join([]string{
		"  func sum(a, b int) int {",
		th.Danger.Render("-     return a - b"),
		th.Success.Render("+     return a + b"),
		"  }",
	}, "\n")
	return l.added.Apply(l.removed.Apply(code))
}

// lightCard is what the lights play on.
func lightCard(th theme.Resolved) string {
	body := strings.Join([]string{
		th.Accent.Bold(true).Render("bezel ✦ light"),
		"Every cell keeps its place;",
		"only its colour moves.",
	}, "\n")
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(th.Accent.GetForeground()).
		Padding(0, 1).Render(body)
}
