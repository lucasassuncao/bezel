package effect

import (
	"image/color"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/internal/cells"
)

func TestSkeletonDrawsBarsWhereTheTextIs(t *testing.T) {
	e := Skeleton(grey, white, 400*ms)
	out := ansi.Strip(at(e, 0.5).Apply("ab cd\n  e  "))
	assert.Equal(t, "▒▒ ▒▒\n  ▒  ", out)
	assert.True(t, at(e, 0.5).Active(), "it sweeps until it is no longer applied")
}

func TestSkeletonsGlowMoves(t *testing.T) {
	e := Skeleton(grey, white, 400*ms)
	row := strings.Repeat("x", 30)
	early := cells.Parse(at(e, 0.3).Apply(row))[0]
	late := cells.Parse(at(e, 0.7).Apply(row))[0]
	brightest := func(cs []cells.Cell) int {
		best, at := uint8(0), 0
		for i, c := range cs {
			if r := toRGBA(c.Fg).R; r > best {
				best, at = r, i
			}
		}
		return at
	}
	assert.Less(t, brightest(early), brightest(late))
}

func TestTraceLightsOnlyTheEdge(t *testing.T) {
	box := "\x1b[38;2;80;80;80m┌──┐\n│ab│\n└──┘\x1b[m"
	g := cells.Parse(at(Trace(color.RGBA{0, 255, 0, 255}, 400*ms), 0.05).Apply(box))
	assert.Greater(t, toRGBA(g[0][0].Fg).G, uint8(80), "the corner the light starts from")
	assert.Equal(t, uint8(80), toRGBA(g[1][1].Fg).G, "inside is never lit")
	sized(t, Trace(white, 400*ms).Apply(box), 4, 3)
}

func TestEdgeIndexGoesClockwise(t *testing.T) {
	order := []struct{ x, y int }{{0, 0}, {3, 0}, {3, 2}, {0, 2}}
	last := -1
	for _, p := range order {
		i, ok := edgeIndex(p.x, p.y, 4, 3)
		assert.True(t, ok)
		assert.Greater(t, i, last)
		last = i
	}
	_, inside := edgeIndex(1, 1, 4, 3)
	assert.False(t, inside)
}

func TestStillLoadingEffectsHoldStill(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.Equal(t, "▒▒", ansi.Strip(Skeleton(grey, white, time.Second).Apply("ab")))
	assert.Equal(t, "ab", Trace(white, time.Second).Apply("ab"))
}
