package fuzzy

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchFindsTheLettersInOrderWhateverTheCase(t *testing.T) {
	_, at, ok := Match("gc", "Git Commit")
	require.True(t, ok)
	require.Equal(t, []int{0, 4}, at)
	_, _, ok = Match("cg", "Git Commit")
	require.False(t, ok, "out of order")
	s, at, ok := Match("", "anything")
	require.True(t, ok)
	require.Zero(t, s)
	require.Nil(t, at)
}

func TestMatchTakesTheTightestRun(t *testing.T) {
	_, at, _ := Match("abc", "a_b_abc")
	require.Equal(t, []int{4, 5, 6}, at, "the run at the end, not the scattered letters")
}

func TestRunsAndWordStartsRankFirst(t *testing.T) {
	texts := []string{"carousel", "cherry-pick", "checkout", "commit"}
	require.Equal(t, []int{2, 1}, Ranked("chk", texts)[:2], "a run beats scattered letters")
	require.Equal(t, []int{1, 0}, Ranked("cp", []string{"scapegoat", "cherry-pick"}), "a word start beats the middle")
	require.Equal(t, []int{0, 1}, Ranked("x", []string{"x1", "x2"}), "ties keep their order")
	require.Empty(t, Ranked("zz", texts))
}

func TestCamelCaseStartsAWord(t *testing.T) {
	a, _, _ := Match("fb", "fooBar")
	b, _, _ := Match("fb", "foobar")
	require.Greater(t, a, b)
}

func BenchmarkRankingTenThousandTexts(b *testing.B) {
	texts := make([]string, 10_000)
	for i := range texts {
		texts[i] = fmt.Sprintf("package/path/item-%05d/file_name_%d.go", i, i%97)
	}
	for b.Loop() {
		Ranked("itemfile", texts)
	}
}
