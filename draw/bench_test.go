package draw

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

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

func BenchmarkSanitize(b *testing.B) {
	clean := strings.Repeat("a plain line of server text ", 40)
	dirty := strings.Repeat("evil \x1b]52;c;aGk=\x07 text \u202e ", 40)
	b.Run("clean", func(b *testing.B) {
		for b.Loop() {
			_ = Sanitize(clean)
		}
	})
	b.Run("dirty", func(b *testing.B) {
		for b.Loop() {
			_ = Sanitize(dirty)
		}
	})
}

func BenchmarkFit(b *testing.B) {
	s := lipgloss.NewStyle().Bold(true).Render(strings.Repeat("wide text ", 30))
	for b.Loop() {
		_ = Fit(s, 80)
	}
}
