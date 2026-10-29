package scene

import (
	"image/color"

	"github.com/lucasassuncao/bezel/internal/cells"
	"github.com/lucasassuncao/bezel/internal/noise"
)

// lifeRate is how many generations a second Life steps through, and
// lifeSpan how many before it seeds a new board: a board can settle.
const (
	lifeRate = 8
	lifeSpan = 160
)

// Life is Conway's Game of Life in c, on a board as large as the scene
// that wraps at its edges, seeded afresh every twenty seconds.
func Life(c color.Color) Scene { return newScene(life, c) }

// life draws the board at the generation the time has reached, worked out
// from its seed: the same time draws the same board at the same size.
func (a Scene) life(cv cells.Canvas, w, h int) {
	gen := int(a.t * lifeRate)
	board := seedBoard(uint64(gen/lifeSpan), w, h) // #nosec G115 -- a non-negative count
	for range gen % lifeSpan {
		board = step(board, w, h)
	}
	c := a.color(0)
	for y, row := range board {
		for x, alive := range row {
			if alive {
				cv.Put(x, y, cells.Cell{Text: "●", Width: 1, Fg: c})
			}
		}
	}
}

// seedBoard is a w×h board with about a third of its cells alive.
func seedBoard(seed uint64, w, h int) [][]bool {
	b := make([][]bool, h)
	for y := range b {
		b[y] = make([]bool, w)
		for x := range b[y] {
			b[y][x] = noise.Unit(noise.Hash(101+seed, uint64(y*w+x))) < 0.33 // #nosec G115 -- a hash key
		}
	}
	return b
}

// step is the next generation of b: a cell lives on with two or three live
// neighbours, and one is born with exactly three.
func step(b [][]bool, w, h int) [][]bool {
	next := make([][]bool, h)
	for y := range next {
		next[y] = make([]bool, w)
		for x := range next[y] {
			n := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if (dx != 0 || dy != 0) && b[(y+dy+h)%h][(x+dx+w)%w] {
						n++
					}
				}
			}
			next[y][x] = n == 3 || (n == 2 && b[y][x])
		}
	}
	return next
}
