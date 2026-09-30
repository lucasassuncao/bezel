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
