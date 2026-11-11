package draw

import (
	"hash/fnv"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// Link is text in style, clickable to url in a terminal that knows OSC 8 and
// plain text in one that does not; empty text shows the url itself.
func Link(text, url string, style lipgloss.Style) string {
	url = strings.Map(dropUnsafe, url)
	if text == "" {
		text = url
	}
	lines := SanitizeLines(text)
	if url == "" {
		return style.Render(strings.Join(lines, "\n"))
	}
	// Each line is a link of its own, so a border drawn after one is never
	// inside it; the shared id lets the terminal treat them as one.
	h := fnv.New32a()
	_, _ = h.Write([]byte(url))
	linked := style.Hyperlink(url, "id=bezel-"+strconv.FormatUint(uint64(h.Sum32()), 36))
	for i, l := range lines {
		lines[i] = linked.Render(l)
	}
	return strings.Join(lines, "\n")
}

// dropUnsafe keeps a url's runes but those that would end the link early or
// act on the terminal.
func dropUnsafe(r rune) rune {
	if unsafeToDraw(r) {
		return -1
	}
	return r
}
