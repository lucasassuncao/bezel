package draw

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"
)

func TestCompositeReplacesOnlyTheCoveredCells(t *testing.T) {
	bg := "0123456789\n0123456789\n0123456789"
	fg := "XX\nXX"
	got := Composite(fg, bg, 3, 1)
	require.Equal(t, "0123456789\n012XX56789\n012XX56789", got)
}

func TestCompositePadsAShortBackgroundLine(t *testing.T) {
	got := Composite("XX", "ab", 5, 0)
	require.Equal(t, "ab   XX", got)
}

func TestCompositeCenterPlacesInTheMiddle(t *testing.T) {
	bg := strings.Repeat(".........\n", 4) + "........."
	got := CompositeCenter("XXX\nXXX", bg)
	lines := strings.Split(got, "\n")
	require.Equal(t, "...XXX...", lines[1])
	require.Equal(t, "...XXX...", lines[2])
	require.Equal(t, ".........", lines[0])
}

// A modal drawn over a screen changes the cells it covers and nothing else:
// the line count stays, and rows it does not reach are left byte for byte.
func FuzzCompositeKeepsTheBackground(f *testing.F) {
	f.Add("[ok]", 3, 1)
	f.Add("a\nb\nc", -2, -1)
	f.Add("日本", 9, 2)
	f.Fuzz(func(t *testing.T, fg string, x, y int) {
		fg = strings.Join(SanitizeLines(fg), "\n")
		x, y = x%20, y%20
		bg := strings.TrimSuffix(strings.Repeat(strings.Repeat(".", 10)+"\n", 6), "\n")
		got := strings.Split(Composite(fg, bg, x, y), "\n")
		bgLines := strings.Split(bg, "\n")
		if len(got) != len(bgLines) {
			t.Fatalf("Composite(%q, x=%d, y=%d) has %d lines, want %d", fg, x, y, len(got), len(bgLines))
		}
		fgRows := strings.Count(fg, "\n") + 1
		for i := range got {
			if (i < y || i >= y+fgRows) && got[i] != bgLines[i] {
				t.Fatalf("Composite(%q, x=%d, y=%d) changed row %d: %q", fg, x, y, i, got[i])
			}
		}
	})
}

// screen is a styled 200x60 frame, the size a modal floats over.
func screen() string {
	line := lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render(strings.Repeat("·", 200))
	return strings.TrimSuffix(strings.Repeat(line+"\n", 60), "\n")
}

func BenchmarkCompositeCenter(b *testing.B) {
	bg := screen()
	fg := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(60).Height(20).Render("modal")
	for b.Loop() {
		_ = CompositeCenter(fg, bg)
	}
}
