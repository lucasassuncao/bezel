package cells

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestACanvasDrawsBlanksWhereNoCellWasPut(t *testing.T) {
	cv := NewCanvas(3, 2)
	cv.Put(1, 0, Cell{Text: "x", Width: 1})
	assert.Equal(t, " x \n   ", cv.Grid().String())
}

func TestACellPutOverAWideOneClearsIt(t *testing.T) {
	cv := NewCanvas(3, 1)
	cv.Put(0, 0, Cell{Text: "界", Width: 2})
	cv.Put(1, 0, Cell{Text: "x", Width: 1})
	assert.Equal(t, " x ", cv.Grid().String())
}

func TestACellThatWouldNotFitIsLeftOut(t *testing.T) {
	cv := NewCanvas(2, 1)
	cv.Put(1, 0, Cell{Text: "界", Width: 2})
	assert.Equal(t, "  ", cv.Grid().String())
}
