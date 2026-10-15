package table

import (
	"strings"
	"testing"

	"github.com/lucasassuncao/bezel/draw"
)

// A row never runs past the width it was fitted to, so a narrow terminal
// truncates cells instead of wrapping them onto the next line.
func FuzzRowFitsTheWidth(f *testing.F) {
	f.Add("Google Chrome", "winget", "126.0.6478.127", 30)
	f.Add("", "", "", 0)
	f.Add("日本語のアプリ", "scoop", "1", 9)
	f.Fuzz(func(t *testing.T, name, source, version string, width int) {
		width = (width%200 + 200) % 200
		cells := []string{draw.Sanitize(name), draw.Sanitize(source), draw.Sanitize(version)}
		if strings.ContainsAny(strings.Join(cells, ""), "\n") {
			t.Skip()
		}
		widths := Fit(cols, [][]string{cells}, width)
		for _, w := range widths {
			if w < 0 {
				t.Fatalf("Fit gave a negative width: %v", widths)
			}
		}
		total := 0
		for _, w := range widths {
			total += w
		}
		total += Gap * (len(widths) - 1)
		if got := draw.Width(Row(widths, cells)); got > total {
			t.Fatalf("Row(%v, %q) is %d wide, columns add to %d", widths, cells, got, total)
		}
	})
}
