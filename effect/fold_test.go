package effect

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/internal/cells"
)

func TestTurnNarrowsTheOldAndWidensTheNew(t *testing.T) {
	require.Equal(t, " aaaa ", strings.Split(swapAt(Turn(400*ms), 0.25), "\n")[0])
	require.Equal(t, " bbbb ", strings.Split(swapAt(Turn(400*ms), 0.75), "\n")[0])
	require.Equal(t, newScreen, swapAt(Turn(400*ms), 1))
}

func TestCubeSharesTheBoxBetweenTheTwo(t *testing.T) {
	require.Equal(t, "aaabbb", strings.Split(swapAt(Cube(Left, 400*ms), 0.5), "\n")[0])
	require.Equal(t, "bbbaaa", strings.Split(swapAt(Cube(Right, 400*ms), 0.5), "\n")[0])
	require.Equal(t, []string{"aaaaaa", "aaaaaa", "bbbbbb", "bbbbbb"}, strings.Split(swapAt(Cube(Up, 400*ms), 0.5), "\n"))
	require.Equal(t, []string{"bbbbbb", "bbbbbb", "aaaaaa", "aaaaaa"}, strings.Split(swapAt(Cube(Down, 400*ms), 0.5), "\n"))
	require.Equal(t, newScreen, swapAt(Cube(Left, 400*ms), 1))
}

// The face going round the far side darkens as it turns away.
func TestCubeDarkensTheFaceTurningAway(t *testing.T) {
	tr := Cube(Left, 400*ms)
	start := time.Unix(10, 0)
	tr, _ = tr.Advance(start)
	tr, _ = tr.Advance(start.Add(300 * ms))
	g := cells.Parse(tr.Apply("\x1b[38;2;200;200;200m"+oldScreen+"\x1b[m", newScreen))
	require.Equal(t, "a", g[0][0].Text)
	require.Less(t, toRGBA(g[0][0].Fg).R, uint8(200))
}

func TestPeelFoldsTheOldBackFromItsCorner(t *testing.T) {
	early := strings.Split(swapAt(Peel(400*ms), 0.25), "\n")
	require.Equal(t, "aaaaaa", early[0], "the far corner is still flat")
	require.True(t, strings.HasSuffix(early[3], "bb"), "the near corner has lifted: %q", early[3])
	require.Contains(t, strings.Join(early, "\n"), "░", "and lies folded back over the page")
	require.NotContains(t, swapAt(Peel(400*ms), 0.5), "a", "halfway the page is folded in two")
	require.Equal(t, newScreen, swapAt(Peel(400*ms), 1))
}

// Whatever the transition and the moment, the frame keeps the blocks' size.
func TestNewTransitionsKeepTheSize(t *testing.T) {
	from, to := "\x1b[1mbold\x1b[m 日本\nsecond", "日本語 x\nthird\nfourth"
	for _, tr := range []Transition{Doors(400 * ms), Shut(400 * ms), Comb(400 * ms), Static(400 * ms),
		Turn(400 * ms), Cube(Up, 400*ms), Cube(Right, 400*ms), Peel(400 * ms)} {
		for p := 0.0; p <= 1; p += 0.1 {
			sized(t, swapAtOf(tr, p, from, to), 9, 3)
		}
	}
}
