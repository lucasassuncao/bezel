package list

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/theme"
)

func bigList(n int) Model {
	rows := make([]Row, n)
	for i := range rows {
		rows[i] = Row{Label: fmt.Sprintf("package-%05d", i), Value: i}
	}
	return New(rows, 40)
}

func BenchmarkUpdateDown(b *testing.B) {
	down := tea.KeyPressMsg{Code: tea.KeyDown}
	m := bigList(100_000)
	for b.Loop() {
		m, _ = m.Update(down)
	}
}

// The filter re-matches every row on each keystroke and each frame.
func BenchmarkFilter(b *testing.B) {
	th := theme.Resolve(theme.Plain, true)
	m := bigList(100_000)
	m, _ = m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m, _ = m.Update(tea.KeyPressMsg{Code: '7', Text: "7"})
	down := tea.KeyPressMsg{Code: tea.KeyDown}
	b.Run("down", func(b *testing.B) {
		for b.Loop() {
			_, _ = m.Update(down)
		}
	})
	b.Run("view", func(b *testing.B) {
		for b.Loop() {
			_ = m.View(th)
		}
	})
}
