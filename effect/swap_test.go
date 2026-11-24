package effect

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/internal/cells"
)

const (
	oldScreen = "aaaaaa\naaaaaa\naaaaaa\naaaaaa"
	newScreen = "bbbbbb\nbbbbbb\nbbbbbb\nbbbbbb"
)

// swapAt is t played to p of its run, over the two screens.
func swapAt(t Transition, p float64) string {
	return ansi.Strip(swapAtOf(t, p, oldScreen, newScreen))
}

// swapAtOf is t played to p of its run, over from and to.
func swapAtOf(t Transition, p float64, from, to string) string {
	start := time.Unix(10, 0)
	t, _ = t.Advance(start)
	t, _ = t.Advance(start.Add(time.Duration(p * float64(t.run.Dur))))
	return t.Apply(from, to)
}

func TestIrisClosesTheOldAndOpensTheNew(t *testing.T) {
	tr := Iris(400 * ms)
	closing := swapAt(tr, 0.25)
	require.Contains(t, closing, "a")
	require.NotContains(t, closing, "b")
	opening := swapAt(tr, 0.75)
	require.Contains(t, opening, "b")
	require.NotContains(t, opening, "a")
	require.Equal(t, newScreen, swapAt(tr, 1))
}

func TestBlindsTurnEverySlatAtOnce(t *testing.T) {
	rows := strings.Split(swapAt(Blinds(2, 400*ms), 0.5), "\n")
	require.Equal(t, []string{"bbbbbb", "aaaaaa", "bbbbbb", "aaaaaa"}, rows, "the top row of each slat has turned")
	require.Equal(t, newScreen, swapAt(Blinds(2, 400*ms), 1))
}

func TestStackSinksTheOldBlockUnderTheNew(t *testing.T) {
	tr := Stack(Left, 400*ms)
	start := time.Unix(10, 0)
	tr, _ = tr.Advance(start)
	tr, _ = tr.Advance(start.Add(200 * ms))
	lit := "\x1b[38;2;200;200;200m" + oldScreen + "\x1b[m"
	g := cells.Parse(tr.Apply(lit, newScreen))
	require.Equal(t, "a", g[0][0].Text, "the old block still shows where the new has not reached")
	require.Less(t, toRGBA(g[0][0].Fg).R, uint8(200), "and it has sunk")
	require.Equal(t, "b", g[0][5].Text)
}

func TestStillSwapsShowTheNewBlock(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	for _, tr := range []Transition{Iris(time.Second), Blinds(3, time.Second), Stack(Up, time.Second)} {
		require.Equal(t, newScreen, ansi.Strip(tr.Apply(oldScreen, newScreen)))
	}
}

func TestDoorsPartTheOldBlockDownTheMiddle(t *testing.T) {
	require.Equal(t, "abbbba", strings.Split(swapAt(Doors(400*ms), 0.5), "\n")[0])
	require.Equal(t, newScreen, swapAt(Doors(400*ms), 1))
}

func TestShutClosesTheNewBlockOverTheOld(t *testing.T) {
	require.Equal(t, "baaaab", strings.Split(swapAt(Shut(400*ms), 0.5), "\n")[0])
	require.Equal(t, newScreen, swapAt(Shut(400*ms), 1))
}

func TestCombSlidesEveryOtherRowFromTheOtherSide(t *testing.T) {
	rows := strings.Split(swapAt(Comb(400*ms), 0.5), "\n")
	require.Equal(t, []string{"aaabbb", "bbbaaa", "aaabbb", "bbbaaa"}, rows)
	require.Equal(t, newScreen, swapAt(Comb(400*ms), 1))
}

func TestStaticSnowsTheOldAwayAndSettlesIntoTheNew(t *testing.T) {
	early, mid, late := swapAt(Static(400*ms), 0.25), swapAt(Static(400*ms), 0.5), swapAt(Static(400*ms), 0.75)
	require.Contains(t, early, "a")
	require.NotContains(t, early, "b")
	require.NotContains(t, mid, "a", "halfway it is all snow")
	require.NotContains(t, mid, "b")
	require.Contains(t, late, "b")
	require.NotContains(t, late, "a")
	require.Equal(t, newScreen, swapAt(Static(400*ms), 1))
}

func TestStillNewSwapsShowTheNewBlock(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	for _, tr := range []Transition{Doors(time.Second), Shut(time.Second), Comb(time.Second), Static(time.Second),
		Turn(time.Second), Cube(Left, time.Second), Peel(time.Second)} {
		require.Equal(t, newScreen, ansi.Strip(tr.Apply(oldScreen, newScreen)))
	}
}
