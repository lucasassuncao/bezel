package shell

import (
	"image/color"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// focusGlow is the focus colour leaving one pane for the focused one.
type focusGlow struct {
	from string // the pane it leaves
	p    float64
	run  frames.Run
}

// newGlow is the colour leaving from over d; nothing to play when d is zero.
func newGlow(from string, d time.Duration) focusGlow {
	if d <= 0 {
		return focusGlow{}
	}
	return focusGlow{from: from, run: frames.NewRun(d)}
}

// glowColor is the border colour of pane name while the focus colour moves:
// the focused pane warming up to the accent, the one left cooling down to
// the border, nil for every other pane and once it settled.
func (s Shell) glowColor(name string) color.Color {
	if !s.glow.run.Active() {
		return nil
	}
	accent, border := lipgloss.Color(s.cfg.Theme.Colors.Accent), lipgloss.Color(s.cfg.Theme.Colors.Border)
	q := frames.Or(s.glow.run.Ease)(s.glow.p)
	switch name {
	case s.Focus():
		return animation.Blend(border, accent, q)
	case s.glow.from:
		return animation.Blend(accent, border, q)
	}
	return nil
}
