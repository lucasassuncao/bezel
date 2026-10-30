package cells

// Canvas is a w×h block of columns that cells are put into, whole, then read
// back as a grid with every empty column blank.
type Canvas struct {
	w, h  int
	slots [][]slot
}

type slot struct {
	c        Cell
	used     bool // a cell starts here
	covering int  // a wide cell starting this many columns to the left covers it
}

// blank is what an empty column reads back as.
var blank = Cell{Text: " ", Width: 1}

func NewCanvas(w, h int) Canvas {
	cv := Canvas{w: w, h: h, slots: make([][]slot, h)}
	all := make([]slot, w*h) // one allocation for every row
	for y := range cv.slots {
		cv.slots[y] = all[y*w : (y+1)*w : (y+1)*w]
	}
	return cv
}

// Put places c at x, y, clearing any cell it lands on; one that would not fit
// whole is left out.
func (cv Canvas) Put(x, y int, c Cell) {
	if y < 0 || y >= cv.h || x < 0 || x+c.Width > cv.w {
		return
	}
	row := cv.slots[y]
	for i := x; i < x+c.Width; i++ {
		if back := row[i].covering; back > 0 {
			cv.clear(row, i-back)
		}
		if row[i].used {
			cv.clear(row, i)
		}
	}
	row[x] = slot{c: c, used: true}
	for i := 1; i < c.Width; i++ {
		row[x+i] = slot{covering: i}
	}
}

// clear empties the cell starting at x and the columns it covers.
func (cv Canvas) clear(row []slot, x int) {
	width := max(row[x].c.Width, 1)
	for i := x; i < x+width && i < len(row); i++ {
		row[i] = slot{}
	}
}

// each calls put for every cell of row y, a blank for every empty column.
func (cv Canvas) each(y int, put func(Cell)) {
	row := cv.slots[y]
	for x := 0; x < cv.w; {
		if row[x].used {
			put(row[x].c)
			x += row[x].c.Width
			continue
		}
		put(blank)
		x++
	}
}

// Grid reads the canvas back, a blank for every column no cell holds.
func (cv Canvas) Grid() Grid {
	g := make(Grid, cv.h)
	for y := range g {
		cells := make([]Cell, 0, cv.w)
		cv.each(y, func(c Cell) { cells = append(cells, c) })
		g[y] = cells
	}
	return g
}

// String is the canvas as text, as Grid().String() is, without the grid.
func (cv Canvas) String() string {
	var p painter
	for y := range cv.h {
		if y > 0 {
			p.newline()
		}
		cv.each(y, p.cell)
	}
	return p.done()
}
