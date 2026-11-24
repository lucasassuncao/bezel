package cells

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestACanvasDrawsBlanksWhereNoCellWasPut(t *testing.T) {
	cv := NewCanvas(3, 2)
	cv.Put(1, 0, Cell{Text: "x", Width: 1})
	require.Equal(t, " x \n   ", cv.Grid().String())
}

func TestACellPutOverAWideOneClearsIt(t *testing.T) {
	cv := NewCanvas(3, 1)
	cv.Put(0, 0, Cell{Text: "界", Width: 2})
	cv.Put(1, 0, Cell{Text: "x", Width: 1})
	require.Equal(t, " x ", cv.Grid().String())
}

func TestACellThatWouldNotFitIsLeftOut(t *testing.T) {
	cv := NewCanvas(2, 1)
	cv.Put(1, 0, Cell{Text: "界", Width: 2})
	require.Equal(t, "  ", cv.Grid().String())
}

// A column past any canvas, as int(math.Inf(1)) may give, is left out too:
// x plus the cell's width must not wrap round to fit.
func TestACellFarPastTheEdgeIsLeftOut(t *testing.T) {
	cv := NewCanvas(2, 1)
	cv.Put(math.MaxInt, 0, Cell{Text: "x", Width: 1})
	cv.Put(math.MaxInt, 0, Cell{Text: "界", Width: 2})
	require.Equal(t, "  ", cv.Grid().String())
}

func TestCanvasStringIsItsGridsString(t *testing.T) {
	cv := NewCanvas(5, 2)
	cv.Put(1, 0, Cell{Text: "日", Width: 2, Attrs: Bold})
	cv.Put(0, 1, Cell{Text: "x", Width: 1})
	require.Equal(t, cv.Grid().String(), cv.String())
}
