package effect

import (
	"image/color"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
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
	e, res := e.Update(e.FrameAt(start.Add(time.Hour)))
	require.Nil(t, res.Cmd)
	require.False(t, e.Active())
	return e
}

func TestFadeInRisesFromItsColour(t *testing.T) {
	e := FadeIn(black, 100*ms).WithEase(animation.Linear)
	require.Equal(t, black, toRGBA(cells.Parse(e.Apply(whiteText))[0][0].Fg), "before its first frame it is all from")
	mid := cells.Parse(halfway(e).Apply(whiteText))[0][0].Fg
	require.Equal(t, grey, toRGBA(mid))
	require.Equal(t, cells.Parse(whiteText), cells.Parse(landed(t, e).Apply(whiteText)))
}

func TestFadeOutSinksIntoItsColour(t *testing.T) {
	require.Equal(t, black, toRGBA(cells.Parse(landed(t, FadeOut(black, 100*ms)).Apply(whiteText))[0][0].Fg))
}

// A sweep pattern dissolves the left half first; at the end only blanks are
// left, the block's size kept.
func TestDissolveFollowsItsPattern(t *testing.T) {
	block := "abcdefgh"
	e := Dissolve(100 * ms).WithEase(animation.Linear).WithPattern(Sweep(Right))
	mid := halfway(e).Apply(block)
	require.True(t, strings.HasPrefix(mid, "   "), "%q", mid)
	require.True(t, strings.HasSuffix(mid, "fgh"), "%q", mid)
	require.Equal(t, "        ", landed(t, e).Apply(block))
}

func TestCoalesceGathersTheBlock(t *testing.T) {
	block := "abcdefghij\nklmnopqrst"
	e := Coalesce(100 * ms).WithEase(animation.Linear)
	require.Equal(t, strings.Repeat(" ", 10)+"\n"+strings.Repeat(" ", 10), e.Apply(block))
	shown := strings.Count(halfway(e).Apply(block), " ")
	require.True(t, shown > 0 && shown < 20, "half gathered: %d blanks", shown)
	require.Equal(t, block, landed(t, e).Apply(block))
}

func TestSweepInRevealsBehindAGlow(t *testing.T) {
	e := SweepIn(black, 100*ms).WithEase(animation.Linear)
	mid := cells.Parse(halfway(e).Apply("\x1b[38;2;255;255;255m" + strings.Repeat("x", 20) + "\x1b[m"))[0]
	require.Equal(t, white, toRGBA(mid[0].Fg), "behind the glow the text is its own")
	require.Equal(t, " ", mid[19].Text, "ahead of the glow nothing shows yet")
	tinted := false
	for _, c := range mid {
		if c.Text == "x" && toRGBA(c.Fg) != white {
			tinted = true
		}
	}
	require.True(t, tinted, "the front is tinted")
}

func TestSlideInComesFromItsEdge(t *testing.T) {
	e := SlideIn(Right, 100*ms).WithEase(animation.Linear)
	require.Equal(t, "    ", e.Apply("abcd"))
	require.Equal(t, "cd  ", halfway(e).Apply("abcd"), "half way in from the left")
	require.Equal(t, "abcd", landed(t, e).Apply("abcd"))
	require.Equal(t, "cd\n  ", halfway(SlideIn(Down, 100*ms).WithEase(animation.Linear)).Apply("ab\ncd"), "the bottom half shows first")
	require.Equal(t, "    ", landed(t, SlideOut(Left, 100*ms)).Apply("abcd"))
}

func TestEvolveStepsThroughShades(t *testing.T) {
	e := Evolve(100 * ms).WithEase(animation.Linear).WithPattern(nil)
	mid := halfway(e).Apply("ab cd")
	require.Equal(t, 5, ansi.StringWidth(mid))
	require.Contains(t, "░▒▓█", string([]rune(mid)[0]))
	require.Equal(t, " ", string([]rune(mid)[2]), "a blank stays blank")
	require.Equal(t, "ab cd", landed(t, e).Apply("ab cd"))
}

// A staggered part starts a gap after the one before and the run grows to
// let the last one finish: halfway through, the first row is in, the last not.
func TestWithStaggerStartsEachPartLater(t *testing.T) {
	e := Coalesce(100*ms).WithEase(animation.Linear).WithPattern(nil).OnRows(0, 1, 2).WithStagger(50 * ms)
	require.Equal(t, 200*ms, e.run.Dur)
	block := "ab\ncd\nef"
	out := strings.Split(ansi.Strip(at(e, 0.6).Apply(block)), "\n")
	require.Equal(t, []string{"ab", "cd", "  "}, out)
	require.Equal(t, block, ansi.Strip(landed(t, e).Apply(block)))
}

func TestStaggerFitsMasksGivenAfterIt(t *testing.T) {
	e := FadeIn(black, 100*ms).WithStagger(30*ms).OnRows(0, 1, 2, 3)
	require.Equal(t, 190*ms, e.run.Dur)
}

func TestStillStaggerShowsEveryPartAtItsEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	e := Coalesce(time.Second).OnRows(0, 1).WithStagger(time.Second)
	require.Equal(t, "ab\ncd", ansi.Strip(e.Apply("ab\ncd")))
}

// Whatever the effect and the moment, a block keeps its size.
func TestEveryEffectKeepsTheBlocksSize(t *testing.T) {
	block := "\x1b[1mbold\x1b[m 日本\nsecond"
	start := time.Unix(1000, 0)
	for _, e := range []Effect{
		FadeIn(black, 100*ms), FadeOut(black, 100*ms), Dissolve(100 * ms), Coalesce(100 * ms),
		SweepIn(white, 100*ms), SlideIn(Left, 100*ms), SlideOut(Up, 100*ms), Evolve(100 * ms),
		Ripple(3, 1, white, 100*ms), Shine(white, 100*ms), Duotone(black, white, 100*ms), Decode(white, 100*ms),
		Flicker(black, 100*ms), Warp(100 * ms), Marching(white, 100*ms), Pour(white, 100*ms),
		SlideIn(Right, 60*ms).OnRows(0, 1).WithStagger(40 * ms),
	} {
		e, _ = e.Update(e.FrameAt(start))
		for d := time.Duration(0); d <= 100*ms; d += 10 * ms {
			e, _ = e.Advance(start.Add(d))
			lines := strings.Split(e.Apply(block), "\n")
			require.Len(t, lines, 2)
			for _, l := range lines {
				require.LessOrEqual(t, ansi.StringWidth(l), 9)
			}
		}
	}
}

func TestStillEffectsShowTheirEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	require.Equal(t, "abcd", Coalesce(time.Second).Apply("abcd"))
	require.Equal(t, "    ", Dissolve(time.Second).Apply("abcd"))
	require.False(t, SlideIn(Right, time.Second).Active())
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

func TestOwnFrameIsHandledAndAsksForTheNext(t *testing.T) {
	x := FadeIn(black, 100*ms)
	_, r := x.Update(x.FrameAt(time.Unix(1000, 0)))
	require.True(t, r.Handled)
	require.NotNil(t, r.Cmd)
}

func TestForeignFrameIsNotHandled(t *testing.T) {
	a, b := FadeIn(black, 100*ms), FadeIn(black, 100*ms)
	_, r := a.Update(b.FrameAt(time.Unix(1000, 0)))
	require.Equal(t, bezel.Result{}, r)
}

// fullScreen is a 120x40 terminal of coloured text, what a frame redraws.
func fullScreen() string {
	rows := make([]string, 40)
	for i := range rows {
		rows[i] = "\x1b[38;2;80;200;120m" + strings.Repeat("panel text ", 6) + "\x1b[m\x1b[38;2;200;120;80m" + strings.Repeat("·", 54) + "\x1b[m"
	}
	return strings.Join(rows, "\n")
}

// BenchmarkEffectFrame is one frame of each effect over a whole screen: the
// budget is 2ms, an eighth of a 60fps frame, so a host can play several.
func BenchmarkEffectFrame(b *testing.B) {
	block := fullScreen()
	for name, e := range map[string]Effect{
		"fade": FadeIn(black, 100*ms), "dissolve": Dissolve(100 * ms), "sweep": SweepIn(white, 100*ms),
		"slide": SlideIn(Left, 100*ms), "evolve": Evolve(100 * ms), "explode": Explode(100 * ms),
		"melt": Melt(white, 100*ms), "tint": Tint(white, 100*ms), "spotlight": Spotlight(10, 5, 40, 10, black, 100*ms),
		"glitch": Glitch(100 * ms), "powerOff": PowerOff(white, 100*ms), "iris": IrisIn(100 * ms), "pixelate": Pixelate(100 * ms),
		"skeleton": Skeleton(black, white, 100*ms), "wave": Dissolve(100 * ms).WithPattern(Wave()),
		"ripple": Ripple(60, 20, white, 100*ms), "shine": Shine(white, 100*ms), "decode": Decode(white, 100*ms),
		"warp": Warp(100 * ms), "pour": Pour(white, 100*ms), "duotone": Duotone(black, white, 100*ms),
	} {
		mid := halfway(e)
		b.Run(name, func(b *testing.B) {
			for b.Loop() {
				mid.Apply(block)
			}
		})
	}
}
