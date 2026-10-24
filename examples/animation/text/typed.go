package main

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/animation/text"
	"github.com/lucasassuncao/bezel/theme"
)

const (
	headline = "Copying 1,204 files to the backup drive…"
	longPath = "~/Downloads/Media/Images/wallhaven-2026/landscapes/aurora-over-the-fjord.png"
	password = "[ password: •••••• ]"
)

type (
	typeMsg  struct{} // t: the headline is typed again
	shakeMsg struct{} // s: the password field refuses
)

// typed is the first tab: a headline typed out, a path too long for its
// room scrolled round on its own ticks, and a field that shakes its head.
type typed struct {
	typewriter text.Typewriter
	marquee    text.Marquee
	shake      text.Shake
}

func newTyped() typed {
	return typed{
		typewriter: text.NewTypewriter(headline, 0),
		marquee:    text.NewMarquee(longPath, 30),
		shake:      text.NewShake(3, 0),
	}
}

// start sets the marquee going on its own ticks.
func (y typed) start() (typed, tea.Cmd) {
	var cmd tea.Cmd
	y.marquee, cmd = y.marquee.Start()
	return y, cmd
}

func (y typed) retype() typed {
	y.typewriter = text.NewTypewriter(headline, 35*time.Millisecond)
	return y
}

func (y typed) refuse() typed {
	y.shake = text.NewShake(3, 450*time.Millisecond)
	return y
}

// update takes the marquee's own ticks.
func (y typed) update(msg tea.Msg) (typed, tea.Cmd) {
	var cmd tea.Cmd
	y.marquee, cmd = y.marquee.Update(msg)
	return y, cmd
}

func (y typed) advance(at time.Time) typed {
	y.typewriter, _ = y.typewriter.Advance(at)
	y.shake, _ = y.shake.Advance(at)
	return y
}

func (y typed) moving() bool { return y.typewriter.Active() || y.shake.Active() }

// rows is the tab's body, a row for each.
func (y typed) rows(th theme.Resolved, label func(kind, variant string) string) string {
	return strings.Join([]string{
		label("typewriter", "t types") + th.Bold.Render(y.typewriter.View()),
		label("marquee", "30 cells") + th.Muted.Render("[") + y.marquee.View() + th.Muted.Render("]"),
		label("shake", "s refuses") + th.Danger.Render(y.shake.View(password)),
	}, "\n")
}
