package animation

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

const (
	oldCard = "AAAA\nAAAA"
	newCard = "BBBB\nBBBB"
)

// tHalf is t after its first frame and one at half its 100ms.
func tHalf(t Transition) Transition {
	start := time.Unix(1000, 0)
	t, _ = t.Advance(start)
	t, _ = t.Advance(start.Add(50 * ms))
	return t
}

func tEnd(t Transition) Transition {
	start := time.Unix(1000, 0)
	t, _ = t.Advance(start)
	t, _ = t.Advance(start.Add(time.Hour))
	return t
}

// Every transition starts on the old block and ends on the new one.
func TestTransitionsRunFromOldToNew(t *testing.T) {
	for name, tr := range map[string]Transition{
		"push": Push(Left, 100*ms), "cover": Cover(Up, 100*ms), "reveal": Reveal(Right, 100*ms),
		"crossfade": CrossFade(100 * ms), "wipe": Wipe(Radial(), 100*ms), "zoom": Zoom(100 * ms), "shuffle": Shuffle(100 * ms),
	} {
		assert.Equal(t, oldCard, ansi.Strip(tr.Apply(oldCard, newCard)), name)
		end := tEnd(tr)
		assert.False(t, end.Active(), name)
		assert.Equal(t, newCard, ansi.Strip(end.Apply(oldCard, newCard)), name)
	}
}

func TestPushShowsBothHalfway(t *testing.T) {
	mid := ansi.Strip(tHalf(Push(Left, 100*ms).WithEase(Linear)).Apply(oldCard, newCard))
	assert.Equal(t, "AABB\nAABB", mid)
}

func TestCoverKeepsTheOldBlockStill(t *testing.T) {
	mid := ansi.Strip(tHalf(Cover(Right, 100*ms).WithEase(Linear)).Apply(oldCard, newCard))
	assert.Equal(t, "BBAA\nBBAA", mid)
}

func TestWipeFollowsItsPattern(t *testing.T) {
	mid := ansi.Strip(tHalf(Wipe(Sweep(Right), 100*ms).WithEase(Linear)).Apply("AAAAAA", "BBBBBB"))
	assert.True(t, strings.HasPrefix(mid, "BB") && strings.HasSuffix(mid, "AA"), mid)
}

func TestCrossFadeBlendsColours(t *testing.T) {
	from := "\x1b[38;2;0;0;0mab\x1b[m"
	to := "\x1b[38;2;255;255;255mcd\x1b[m"
	mid := parse(tHalf(CrossFade(100*ms).WithEase(Linear)).Apply(from, to))[0][0]
	assert.Equal(t, "c", mid.text, "the text changes over at halfway")
	assert.Equal(t, grey, toRGBA(mid.fg))
}

// A smaller block is padded, so the frame keeps the larger one's size.
func TestTransitionPadsToTheLargerBlock(t *testing.T) {
	out := ansi.Strip(tHalf(Zoom(100*ms)).Apply("A", "BBB\nBBB"))
	lines := strings.Split(out, "\n")
	assert.Len(t, lines, 2)
	assert.Equal(t, 3, len(lines[0]))
}

func TestTransitionTicksOnItsOwnFrames(t *testing.T) {
	tr := Shuffle(100 * ms)
	assert.True(t, tr.Owns(tr.FrameAt(time.Unix(1, 0))))
	tr, cmd := tr.Update(tr.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, cmd)
	assert.False(t, tr.Owns(FrameMsg{}))
}
