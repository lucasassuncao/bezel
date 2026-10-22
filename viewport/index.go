package viewport

import (
	"sort"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/internal/edit"
)

// index maps rows to the content's lines: each line's cells, the longest,
// and while wrapping the rows before each line.
type index struct {
	lines   []string // the content's, shared and never written
	widths  []int
	narrow  []bool // the plain text is printable ASCII, so rows are arithmetic
	before  []int  // nil when not wrapping; else len(widths)+1 running sums
	longest int
}

// newIndex measures every line once.
func newIndex(lines []string) index {
	x := index{lines: lines, widths: make([]int, len(lines)), narrow: make([]bool, len(lines))}
	for i, l := range lines {
		x.widths[i] = ansi.StringWidth(l)
		x.narrow[i] = isNarrow(plainOf(l))
		x.longest = max(x.longest, x.widths[i])
	}
	return x
}

// wrapped counts each line's rows at a text width; below 1 is 1.
func (x index) wrapped(width int) index {
	width = max(width, 1)
	x.before = make([]int, len(x.widths)+1)
	for i, w := range x.widths {
		n := max(1, (w+width-1)/width)
		if !x.narrow[i] {
			n = len(breaks(x.lines[i], width))
		}
		x.before[i+1] = x.before[i] + n
	}
	return x
}

// unwrapped is a row a line again.
func (x index) unwrapped() index {
	x.before = nil
	return x
}

// total is how many rows the content takes.
func (x index) total() int {
	if x.before == nil {
		return len(x.widths)
	}
	return x.before[len(x.widths)]
}

// row is the first row of line.
func (x index) row(line int) int {
	line = min(max(line, 0), len(x.widths))
	if x.before == nil {
		return line
	}
	return x.before[line]
}

// at is the line holding row and which of its rows that is; past the end it
// is the line count. A binary search, so it costs the same for any length.
func (x index) at(row int) (line, part int) {
	row = max(row, 0)
	if row >= x.total() {
		return len(x.widths), 0
	}
	if x.before == nil {
		return row, 0
	}
	line = sort.Search(len(x.widths), func(i int) bool { return x.before[i+1] > row })
	return line, row - x.before[line]
}

// breaks is where each of line's rows starts, in cells, wrapped at width:
// never inside a grapheme, so a wide one at the edge opens the next row.
func breaks(line string, width int) []int {
	width = max(width, 1)
	plain := plainOf(line)
	if isNarrow(plain) {
		out := make([]int, max(1, (len(plain)+width-1)/width))
		for i := range out {
			out[i] = i * width
		}
		return out
	}
	out := []int{0}
	pos, used := 0, 0
	for _, g := range edit.Graphemes(plain) {
		c := edit.Cells(g)
		if used > 0 && used+c > width {
			out = append(out, pos)
			used = 0
		}
		pos += c
		used += c
	}
	return out
}

// plainOf is line without its escape sequences; most lines have none.
func plainOf(line string) string {
	if strings.IndexByte(line, '\x1b') < 0 {
		return line
	}
	return ansi.Strip(line)
}

// isNarrow reports whether every byte of plain is a printable ASCII
// character, one cell each.
func isNarrow(plain string) bool {
	for i := 0; i < len(plain); i++ {
		if plain[i] < 0x20 || plain[i] > 0x7e {
			return false
		}
	}
	return true
}
