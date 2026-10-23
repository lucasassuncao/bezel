package animation

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sized reports whether block is h lines of at most w columns.
func sized(t *testing.T, block string, w, h int) {
	t.Helper()
	lines := strings.Split(block, "\n")
	require.Len(t, lines, h)
	for _, l := range lines {
		assert.LessOrEqual(t, ansi.StringWidth(l), w, "%q", l)
	}
}

func TestExplodeFliesTheCellsApart(t *testing.T) {
	block := "abcdefgh\nijklmnop\nqrstuvwx"
	e := Explode(100 * ms).WithEase(Linear)
	assert.Equal(t, block, e.Apply(block), "before its first frame nothing has moved")
	mid := halfway(e).Apply(block)
	sized(t, mid, 8, 3)
	assert.NotEqual(t, block, mid)
	assert.NotEqual(t, "", strings.TrimSpace(mid), "half way, pieces are still in the air")
	end := landed(t, e).Apply(block)
	sized(t, end, 8, 3)
	assert.Equal(t, "", strings.TrimSpace(end), "at the end every piece has gone")
}

func TestExpandOpensFromTheMiddle(t *testing.T) {
	e := Expand(white, 100*ms).WithEase(Linear)
	assert.Equal(t, "", strings.Trim(ansi.Strip(e.Apply("abcdefgh")), " ▐▌"), "it starts closed: only its seam shows")
	mid := ansi.Strip(halfway(e).Apply("abcdefgh"))
	assert.Equal(t, " ▐cdef▌ ", mid)
	assert.Equal(t, "abcdefgh", landed(t, e).Apply("abcdefgh"))
}

func TestStretchFillsFromAnEdge(t *testing.T) {
	e := Stretch(Right, white, 100*ms).WithEase(Linear)
	assert.Equal(t, "abcd    ", ansi.Strip(halfway(e).Apply("abcdefgh")))
	start := time.Unix(1000, 0)
	e, _ = e.Update(e.FrameAt(start))
	e, _ = e.Update(e.FrameAt(start.Add(55 * ms)))
	assert.Equal(t, "abcd▍   ", ansi.Strip(e.Apply("abcdefgh")), "the front cell is filled in eighths")
	assert.Equal(t, "    efgh", ansi.Strip(halfway(Stretch(Left, white, 100*ms).WithEase(Linear)).Apply("abcdefgh")))
	assert.Equal(t, "  \ncd", ansi.Strip(halfway(Stretch(Up, white, 100*ms).WithEase(Linear)).Apply("ab\ncd")))
	assert.Equal(t, "ab\n  ", ansi.Strip(halfway(Stretch(Down, white, 100*ms).WithEase(Linear)).Apply("ab\ncd")))
	assert.Equal(t, "abcdefgh", landed(t, e).Apply("abcdefgh"))
}

func TestStillGeometryShowsItsEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.Equal(t, "    ", Explode(time.Second).Apply("abcd"))
	assert.Equal(t, "abcd", Expand(white, time.Second).Apply("abcd"))
	assert.Equal(t, "abcd", Stretch(Right, white, time.Second).Apply("abcd"))
}
