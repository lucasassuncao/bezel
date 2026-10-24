package effect

import (
	"image/color"
	"time"

	"github.com/lucasassuncao/bezel/animation"
)

// The presets name what the app means rather than how it looks, so a screen
// asks for Attention and the package decides what that is.

// Enter brings a block in, gathering it out of bg from the middle outward.
func Enter(bg color.Color) Effect {
	return FadeIn(bg, 300*time.Millisecond).WithPattern(Radial()).WithEase(animation.EaseOut)
}

// Exit takes a block away, sinking it into bg from the edges inward.
func Exit(bg color.Color) Effect {
	return FadeOut(bg, 250*time.Millisecond).WithPattern(Inverted(Radial())).WithEase(animation.EaseIn)
}

// Attention draws the eye to a block with two pulses of c.
func Attention(c color.Color) Effect { return Pulse(c, 450*time.Millisecond).WithRepeat(1) }

// Success acknowledges a block with one wash of c, a green say.
func Success(c color.Color) Effect { return Tint(c, 500*time.Millisecond).WithStrength(0.5) }

// Failure lights a block in c at once and lets it fade, a red say.
func Failure(c color.Color) Effect { return Flash(c, 450*time.Millisecond).WithStrength(0.7) }
