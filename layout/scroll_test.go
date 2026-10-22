package layout

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClampScrollKeepsTheCursorInTheWindow(t *testing.T) {
	require.Equal(t, 0, ClampScroll(0, 0, 10))
	require.Equal(t, 3, ClampScroll(3, 5, 10))  // cursor above: scroll up to it
	require.Equal(t, 6, ClampScroll(15, 0, 10)) // cursor below: last row shows it
	require.Equal(t, 5, ClampScroll(7, 5, 10))  // inside: untouched
	require.Equal(t, 0, ClampScroll(2, -4, 0))  // no height: only the floor
}

func TestScrollStartCentresTheCursorWithinTheList(t *testing.T) {
	require.Equal(t, 0, ScrollStart(3, 5, 10))    // the whole list fits
	require.Equal(t, 0, ScrollStart(3, 50, 0))    // no height
	require.Equal(t, 0, ScrollStart(2, 50, 10))   // near the top: no negative start
	require.Equal(t, 15, ScrollStart(20, 50, 10)) // centred on the cursor
	require.Equal(t, 40, ScrollStart(48, 50, 10)) // near the end: the last page
}

func TestWindowClampKeepsTheCursorClearOfTheMoreRow(t *testing.T) {
	// The last row says "↓ N more" while rows overflow, so the cursor stops one short.
	require.Equal(t, Window{Cursor: 9, Offset: 1, Height: 10}, Window{Cursor: 9, Height: 10}.Clamp(30))
	require.Equal(t, Window{Cursor: 29, Offset: 20, Height: 10}, Window{Cursor: 29, Height: 10}.Clamp(30))
	require.Equal(t, Window{Cursor: 5, Offset: 3, Height: 10}, Window{Cursor: 5, Offset: 3, Height: 10}.Clamp(30))
}

func TestWindowClampFollowsRowsThatShrank(t *testing.T) {
	require.Equal(t, Window{Cursor: 4, Height: 10}, Window{Cursor: 4, Offset: 25, Height: 10}.Clamp(8),
		"a short list scrolled past its end shows from the top, not blank rows")
	require.Equal(t, Window{Cursor: 29, Offset: 20, Height: 10}, Window{Cursor: 50, Height: 10}.Clamp(30))
	require.Equal(t, Window{Height: 10}, Window{Cursor: 3, Offset: 2, Height: 10}.Clamp(0))
}

func TestWindowBoundsTradesTheLastRowForTheCount(t *testing.T) {
	start, end, below := Window{Offset: 1, Height: 10}.Bounds(30)
	require.Equal(t, []int{1, 10, 20}, []int{start, end, below})

	start, end, below = Window{Offset: 20, Height: 10}.Bounds(30)
	require.Equal(t, []int{20, 30, 0}, []int{start, end, below}, "the last page needs no count")

	start, end, below = Window{Height: 1}.Bounds(30)
	require.Equal(t, []int{0, 1, 0}, []int{start, end, below}, "a one-row window keeps its row")

	start, end, below = Window{Height: 10}.Bounds(4)
	require.Equal(t, []int{0, 4, 0}, []int{start, end, below})
}

func TestWindowMovePassesOverSkippedRows(t *testing.T) {
	heading := func(i int) bool { return i == 0 || i == 3 || i == 4 }
	w := Window{Cursor: 2, Height: 10}

	require.Equal(t, 5, w.Move(1, 8, heading).Cursor)
	require.Equal(t, 2, w.Move(1, 8, heading).Move(-1, 8, heading).Cursor)
	require.Equal(t, 1, Window{Cursor: 1, Height: 10}.Move(-1, 8, heading).Cursor, "a heading on top is no row to land on")
	require.Equal(t, 7, Window{Cursor: 1, Height: 10}.Move(5, 8, heading).Cursor, "a long step stops at the last row")
}

func TestWindowMoveScrolls(t *testing.T) {
	require.Equal(t, Window{Cursor: 2, Offset: 1, Height: 3}, Window{Cursor: 1, Height: 3}.Move(1, 5, nil))
	require.Equal(t, Window{Height: 3}, Window{Cursor: 2, Height: 3}.Move(1, 0, nil))
}

func TestWindowScrollLeavesTheMoreRowOut(t *testing.T) {
	require.Equal(t, Scroll{Offset: 1, Visible: 9, Total: 30}, Window{Offset: 1, Height: 10}.Scroll(30))
	require.Equal(t, Scroll{Offset: 20, Visible: 10, Total: 30}, Window{Offset: 20, Height: 10}.Scroll(30))
	require.Equal(t, Scroll{Visible: 4, Total: 4}, Window{Height: 10}.Scroll(4))
}

func TestScrollThumbSizesToTheShareShown(t *testing.T) {
	pos, size := Scroll{Visible: 10, Total: 100}.Thumb(10)
	require.Equal(t, []int{0, 1}, []int{pos, size})
	pos, size = Scroll{Visible: 50, Total: 100}.Thumb(10)
	require.Equal(t, []int{0, 5}, []int{pos, size})
	pos, size = Scroll{Visible: 29, Total: 30}.Thumb(10)
	require.Equal(t, []int{0, 9}, []int{pos, size}, "a full track would say nothing is hidden")
}

func TestScrollThumbTouchesAnEndOnlyAtThatEnd(t *testing.T) {
	thumb := func(offset int) int {
		pos, _ := Scroll{Offset: offset, Visible: 10, Total: 100}.Thumb(10)
		return pos
	}
	require.Equal(t, 0, thumb(0))
	require.Equal(t, 1, thumb(1), "scrolled at all: off the top")
	require.Equal(t, 5, thumb(50))
	require.Equal(t, 8, thumb(89), "not at the end: off the bottom")
	require.Equal(t, 9, thumb(90))
	require.Equal(t, 9, thumb(500), "an offset past the end reads as the end")
}

func TestScrollThumbIsNoneWhenEverythingShows(t *testing.T) {
	for _, s := range []Scroll{{}, {Visible: 10, Total: 10}, {Visible: 10, Total: 4}, {Total: 30}} {
		_, size := s.Thumb(10)
		require.Zero(t, size, "%+v", s)
	}
	_, size := Scroll{Visible: 10, Total: 100}.Thumb(0)
	require.Zero(t, size, "no track, no thumb")
	pos, size := Scroll{Offset: 5, Visible: 10, Total: 100}.Thumb(1)
	require.Equal(t, []int{0, 1}, []int{pos, size}, "a one-row track is all thumb")
}
