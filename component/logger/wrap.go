package logger

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// minRoom keeps a deep tree on a narrow terminal from cutting a value into
// slivers; its lines run past the width instead.
const minRoom = 12

// wrap cuts s into lines of at most room cells, keeping every character. A
// line ends after a path separator or space if one is far enough in, else
// after - _ . , ; :, else at room. room 0 only splits s at its own newlines.
func wrap(s string, room int) []string {
	var out []string
	for line := range strings.SplitSeq(s, "\n") {
		if room <= 0 {
			out = append(out, line)
			continue
		}
		out = append(out, cut(line, max(room, minRoom))...)
	}
	return out
}

func cut(s string, room int) []string {
	var out []string
	for ansi.StringWidth(s) > room {
		end := fit(s, room)
		if at := breakAt(s[:end], room); at > 0 {
			end = at
		}
		out = append(out, s[:end])
		s = s[end:]
	}
	return append(out, s)
}

// fit is the byte length of the longest prefix of s within room cells.
func fit(s string, room int) int {
	w := 0
	for i, r := range s {
		w += ansi.StringWidth(string(r))
		if w > room {
			return max(i, len(string(r)))
		}
	}
	return len(s)
}

// breakAt is where line should end: just after its last strong separator in
// the second two thirds of room, else its last weak one, else 0.
func breakAt(line string, room int) int {
	floor := room / 3
	for _, set := range []string{`/\ `, "-_.,;:"} {
		if i := strings.LastIndexAny(line, set); i >= 0 && ansi.StringWidth(line[:i+1]) > floor {
			return i + 1
		}
	}
	return 0
}
