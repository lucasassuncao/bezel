package effect

import (
	"image/color"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/cells"
)

// whiteText is "abcd" drawn white, for the effects that change colour.
var whiteText = "\x1b[38;2;255;255;255mabcd\x1b[m"

// halfway is e after its first frame and a frame at half its 100ms.
func halfway(e Effect) Effect {
	start := time.Unix(1000, 0)
	e, _ = e.Update(e.FrameAt(start))
	e, _ = e.Update(e.FrameAt(start.Add(50 * ms)))
	return e
}

// landed is e after a frame long past its end.
func landed(t *testing.T, e Effect) Effect {
	t.Helper()
	start := time.Unix(1000, 0)
	e, _ = e.Update(e.FrameAt(start))
	e, cmd := e.Update(e.FrameAt(start.Add(time.Hour)))
	require.Nil(t, cmd)
	require.False(t, e.Active())
	return e
}

func TestFadeInRisesFromItsColour(t *testing.T) {
	e := FadeIn(black, 100*ms).WithEase(animation.Linear)
	assert.Equal(t, black, toRGBA(cells.Parse(e.Apply(whiteText))[0][0].Fg), "before its first frame it is all from")
	mid := cells.Parse(halfway(e).Apply(whiteText))[0][0].Fg
	assert.Equal(t, grey, toRGBA(mid))
	assert.Equal(t, cells.Parse(whiteText), cells.Parse(landed(t, e).Apply(whiteText)))
}

func TestFadeOutSinksIntoItsColour(t *testing.T) {
	assert.Equal(t, black, toRGBA(cells.Parse(landed(t, FadeOut(black, 100*ms)).Apply(whiteText))[0][0].Fg))
}

// A sweep pattern dissolves the left half first; at the end only blanks are
// left, the block's size kept.
func TestDissolveFollowsItsPattern(t *testing.T) {
	block := "abcdefgh"
	e := Dissolve(100 * ms).WithEase(animation.Linear).WithPattern(Sweep(Right))
	mid := halfway(e).Apply(block)
	assert.True(t, strings.HasPrefix(mid, "   "), "%q", mid)
	assert.True(t, strings.HasSuffix(mid, "fgh"), "%q", mid)
	assert.Equal(t, "        ", landed(t, e).Apply(block))
}

func TestCoalesceGathersTheBlock(t *testing.T) {
	block := "abcdefghij\nklmnopqrst"
	e := Coalesce(100 * ms).WithEase(animation.Linear)
	assert.Equal(t, strings.Repeat(" ", 10)+"\n"+strings.Repeat(" ", 10), e.Apply(block))
	shown := strings.Count(halfway(e).Apply(block), " ")
	assert.True(t, shown > 0 && shown < 20, "half gathered: %d blanks", shown)
	assert.Equal(t, block, landed(t, e).Apply(block))
}

func TestSweepInRevealsBehindAGlow(t *testing.T) {
	e := SweepIn(black, 100*ms).WithEase(animation.Linear)
	mid := cells.Parse(halfway(e).Apply("\x1b[38;2;255;255;255m" + strings.Repeat("x", 20) + "\x1b[m"))[0]
	assert.Equal(t, white, toRGBA(mid[0].Fg), "behind the glow the text is its own")
	assert.Equal(t, " ", mid[19].Text, "ahead of the glow nothing shows yet")
	tinted := false
	for _, c := range mid {
		if c.Text == "x" && toRGBA(c.Fg) != white {
			tinted = true
		}
	}
	assert.True(t, tinted, "the front is tinted")
}

func TestSlideInComesFromItsEdge(t *testing.T) {
	e := SlideIn(Right, 100*ms).WithEase(animation.Linear)
	assert.Equal(t, "    ", e.Apply("abcd"))
	assert.Equal(t, "cd  ", halfway(e).Apply("abcd"), "half way in from the left")
	assert.Equal(t, "abcd", landed(t, e).Apply("abcd"))
	assert.Equal(t, "cd\n  ", halfway(SlideIn(Down, 100*ms).WithEase(animation.Linear)).Apply("ab\ncd"), "the bottom half shows first")
	assert.Equal(t, "    ", landed(t, SlideOut(Left, 100*ms)).Apply("abcd"))
}

func TestEvolveStepsThroughShades(t *testing.T) {
	e := Evolve(100 * ms).WithEase(animation.Linear).WithPattern(nil)
	mid := halfway(e).Apply("ab cd")
	assert.Equal(t, 5, ansi.StringWidth(mid))
	assert.Contains(t, "░▒▓█", string([]rune(mid)[0]))
	assert.Equal(t, " ", string([]rune(mid)[2]), "a blank stays blank")
	assert.Equal(t, "ab cd", landed(t, e).Apply("ab cd"))
}

// Whatever the effect and the moment, a block keeps its size.
func TestEveryEffectKeepsTheBlocksSize(t *testing.T) {
	block := "\x1b[1mbold\x1b[m 日本\nsecond"
	start := time.Unix(1000, 0)
	for _, e := range []Effect{
		FadeIn(black, 100*ms), FadeOut(black, 100*ms), Dissolve(100 * ms), Coalesce(100 * ms),
		SweepIn(white, 100*ms), SlideIn(Left, 100*ms), SlideOut(Up, 100*ms), Evolve(100 * ms),
	} {
		e, _ = e.Update(e.FrameAt(start))
		for d := time.Duration(0); d <= 100*ms; d += 10 * ms {
			e, _ = e.Advance(start.Add(d))
			lines := strings.Split(e.Apply(block), "\n")
			require.Len(t, lines, 2)
			for _, l := range lines {
				assert.LessOrEqual(t, ansi.StringWidth(l), 9)
			}
		}
	}
}

func TestStillEffectsShowTheirEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.Equal(t, "abcd", Coalesce(time.Second).Apply("abcd"))
	assert.Equal(t, "    ", Dissolve(time.Second).Apply("abcd"))
	assert.False(t, SlideIn(Right, time.Second).Active())
}

// toRGBA is c as 8-bit RGBA, so colours read back from SGR compare by value.
func toRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

const ms = time.Millisecond

var (
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
	grey  = color.RGBA{128, 128, 128, 255}
)

// TestMain clears the motion switches, so a CI with TERM=dumb still tests motion.
func TestMain(m *testing.M) {
	_ = os.Setenv("BEZEL_NO_MOTION", "")
	_ = os.Setenv("BEZEL_MOTION", "")
	_ = os.Setenv("TERM", "xterm-256color")
	os.Exit(m.Run())
}
