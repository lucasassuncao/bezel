package effect

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
)

// sized reports whether block is h lines of at most w columns.
func sized(t *testing.T, block string, w, h int) {
	t.Helper()
	lines := strings.Split(block, "\n")
	require.Len(t, lines, h)
	for _, l := range lines {
		require.LessOrEqual(t, ansi.StringWidth(l), w, "%q", l)
	}
}

func TestExplodeFliesTheCellsApart(t *testing.T) {
	block := "abcdefgh\nijklmnop\nqrstuvwx"
	e := Explode(100 * ms).WithEase(animation.Linear)
	require.Equal(t, block, e.Apply(block), "before its first frame nothing has moved")
	mid := halfway(e).Apply(block)
	sized(t, mid, 8, 3)
	require.NotEqual(t, block, mid)
	require.NotEqual(t, "", strings.TrimSpace(mid), "half way, pieces are still in the air")
	end := landed(t, e).Apply(block)
	sized(t, end, 8, 3)
	require.Equal(t, "", strings.TrimSpace(end), "at the end every piece has gone")
}

func TestExpandOpensFromTheMiddle(t *testing.T) {
	e := Expand(white, 100*ms).WithEase(animation.Linear)
	require.Equal(t, "", strings.Trim(ansi.Strip(e.Apply("abcdefgh")), " ▐▌"), "it starts closed: only its seam shows")
	mid := ansi.Strip(halfway(e).Apply("abcdefgh"))
	require.Equal(t, " ▐cdef▌ ", mid)
	require.Equal(t, "abcdefgh", landed(t, e).Apply("abcdefgh"))
}

func TestStretchFillsFromAnEdge(t *testing.T) {
	e := Stretch(Right, white, 100*ms).WithEase(animation.Linear)
	require.Equal(t, "abcd    ", ansi.Strip(halfway(e).Apply("abcdefgh")))
	start := time.Unix(1000, 0)
	e, _ = e.Update(e.FrameAt(start))
	e, _ = e.Update(e.FrameAt(start.Add(55 * ms)))
	require.Equal(t, "abcd▍   ", ansi.Strip(e.Apply("abcdefgh")), "the front cell is filled in eighths")
	require.Equal(t, "    efgh", ansi.Strip(halfway(Stretch(Left, white, 100*ms).WithEase(animation.Linear)).Apply("abcdefgh")))
	require.Equal(t, "  \ncd", ansi.Strip(halfway(Stretch(Up, white, 100*ms).WithEase(animation.Linear)).Apply("ab\ncd")))
	require.Equal(t, "ab\n  ", ansi.Strip(halfway(Stretch(Down, white, 100*ms).WithEase(animation.Linear)).Apply("ab\ncd")))
	require.Equal(t, "abcdefgh", landed(t, e).Apply("abcdefgh"))
}

func TestStillGeometryShowsItsEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	require.Equal(t, "    ", Explode(time.Second).Apply("abcd"))
	require.Equal(t, "abcd", Expand(white, time.Second).Apply("abcd"))
	require.Equal(t, "abcd", Stretch(Right, white, time.Second).Apply("abcd"))
}

func TestPourFillsFromTheBottom(t *testing.T) {
	e := Pour(white, 400*ms).WithEase(animation.Linear)
	require.Equal(t, strings.Repeat(" ", 8), strings.Split(ansi.Strip(e.Apply(screen)), "\n")[4], "nothing has risen yet")
	rows := strings.Split(ansi.Strip(at(e, 0.5).Apply(screen)), "\n")
	require.Equal(t, "6789ABCD", rows[4], "the bottom row is under the surface")
	require.Equal(t, strings.Repeat(" ", 8), rows[0], "the top one not yet")
	require.Equal(t, screen, ansi.Strip(landed(t, e).Apply(screen)))
}
