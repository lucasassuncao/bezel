package effect

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/internal/cells"
)

const (
	oldScreen = "aaaaaa\naaaaaa\naaaaaa\naaaaaa"
	newScreen = "bbbbbb\nbbbbbb\nbbbbbb\nbbbbbb"
)

// swapAt is t played to p of its run, over the two screens.
func swapAt(t Transition, p float64) string {
	start := time.Unix(10, 0)
	t, _ = t.Advance(start)
	t, _ = t.Advance(start.Add(time.Duration(p * float64(t.run.Dur))))
	return ansi.Strip(t.Apply(oldScreen, newScreen))
}

func TestIrisClosesTheOldAndOpensTheNew(t *testing.T) {
	tr := Iris(400 * ms)
	closing := swapAt(tr, 0.25)
	assert.Contains(t, closing, "a")
	assert.NotContains(t, closing, "b")
	opening := swapAt(tr, 0.75)
	assert.Contains(t, opening, "b")
	assert.NotContains(t, opening, "a")
	assert.Equal(t, newScreen, swapAt(tr, 1))
}

func TestBlindsTurnEverySlatAtOnce(t *testing.T) {
	rows := strings.Split(swapAt(Blinds(2, 400*ms), 0.5), "\n")
	assert.Equal(t, []string{"bbbbbb", "aaaaaa", "bbbbbb", "aaaaaa"}, rows, "the top row of each slat has turned")
	assert.Equal(t, newScreen, swapAt(Blinds(2, 400*ms), 1))
}

func TestStackSinksTheOldBlockUnderTheNew(t *testing.T) {
	tr := Stack(Left, 400*ms)
	start := time.Unix(10, 0)
	tr, _ = tr.Advance(start)
	tr, _ = tr.Advance(start.Add(200 * ms))
	lit := "\x1b[38;2;200;200;200m" + oldScreen + "\x1b[m"
	g := cells.Parse(tr.Apply(lit, newScreen))
	assert.Equal(t, "a", g[0][0].Text, "the old block still shows where the new has not reached")
	assert.Less(t, toRGBA(g[0][0].Fg).R, uint8(200), "and it has sunk")
	assert.Equal(t, "b", g[0][5].Text)
}

func TestStillSwapsShowTheNewBlock(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	for _, tr := range []Transition{Iris(time.Second), Blinds(3, time.Second), Stack(Up, time.Second)} {
		assert.Equal(t, newScreen, ansi.Strip(tr.Apply(oldScreen, newScreen)))
	}
}
