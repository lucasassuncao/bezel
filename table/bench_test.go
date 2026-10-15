package table

import (
	"fmt"
	"testing"
)

func BenchmarkFit(b *testing.B) {
	rows := make([][]string, 10_000)
	for i := range rows {
		rows[i] = []string{fmt.Sprint("Application number ", i), "winget", fmt.Sprintf("1.%d.0", i)}
	}
	for b.Loop() {
		_ = Fit(cols, rows, 60)
	}
}
