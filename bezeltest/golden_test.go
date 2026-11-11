package bezeltest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// recorder is a testing.TB that keeps what failed instead of failing.
type recorder struct {
	testing.TB
	name   string
	errors []string
	fatal  bool
}

func (r *recorder) Name() string        { return r.name }
func (r *recorder) Helper()             {}
func (r *recorder) Logf(string, ...any) {}
func (r *recorder) Errorf(f string, a ...any) {
	r.errors = append(r.errors, fmt.Sprintf(f, a...))
}

// Fatalf stops the helper the way t.Fatalf would, through a panic run catches.
func (r *recorder) Fatalf(f string, a ...any) {
	r.errors = append(r.errors, fmt.Sprintf(f, a...))
	r.fatal = true
	panic(r)
}

func run(r *recorder, f func()) {
	defer func() {
		if p := recover(); p != nil && p != r {
			panic(p)
		}
	}()
	f()
}

func inTemp(t *testing.T) {
	t.Chdir(t.TempDir())
}

func TestGoldenWritesThenMatches(t *testing.T) {
	inTemp(t)
	r := &recorder{TB: t, name: "TestScreen/narrow"}
	run(r, func() { golden(r, "a\nb", ".golden", true) })
	require.Empty(t, r.errors)
	data, err := os.ReadFile(filepath.Join("testdata", "TestScreen", "narrow.golden"))
	require.NoError(t, err)
	require.Equal(t, "a\nb", string(data))

	run(r, func() { golden(r, "a\nb", ".golden", false) })
	require.Empty(t, r.errors)
}

func TestGoldenNamesTheLineThatChanged(t *testing.T) {
	inTemp(t)
	r := &recorder{TB: t, name: "TestScreen"}
	run(r, func() { golden(r, "a\nb\nc", ".golden", true) })
	run(r, func() { golden(r, "a\nB\nc", ".golden", false) })
	require.Len(t, r.errors, 1)
	require.Contains(t, r.errors[0], "line 2")
	require.Contains(t, r.errors[0], `"b"`)
	require.Contains(t, r.errors[0], `"B"`)
	require.NotContains(t, r.errors[0], "line 1")
}

func TestGoldenWithoutAFileSaysHowToWriteIt(t *testing.T) {
	inTemp(t)
	r := &recorder{TB: t, name: "TestNothing"}
	run(r, func() { golden(r, "x", ".golden", false) })
	require.True(t, r.fatal)
	require.Contains(t, r.errors[0], "-update")
}

func TestGoldenReadsACRLFCheckoutAsTheSameScreen(t *testing.T) {
	inTemp(t)
	require.NoError(t, os.MkdirAll("testdata", 0o750))
	require.NoError(t, os.WriteFile(filepath.Join("testdata", "TestCRLF.golden"), []byte("a\r\nb"), 0o600))
	r := &recorder{TB: t, name: "TestCRLF"}
	run(r, func() { golden(r, "a\nb", ".golden", false) })
	require.Empty(t, r.errors)
}

func TestGoldenStripsColoursAndGoldenANSIKeepsThem(t *testing.T) {
	inTemp(t)
	red := "\x1b[31mred\x1b[0m"
	r := &recorder{TB: t, name: "TestColour"}
	run(r, func() { golden(r, red, ".ansi.golden", true) })
	data, err := os.ReadFile(filepath.Join("testdata", "TestColour.ansi.golden"))
	require.NoError(t, err)
	require.Equal(t, red, string(data))

	// Golden is the public way in, and it strips before comparing.
	require.NoError(t, os.WriteFile(filepath.Join("testdata", "TestGoldenStripsColoursAndGoldenANSIKeepsThem.golden"), []byte("red"), 0o600))
	Golden(t, red)
}

func TestGoldenPathKeepsOddNamesInsideTestdata(t *testing.T) {
	p := goldenPath(`TestX/a b:c?<d>`, ".golden")
	require.Equal(t, filepath.Join("testdata", "TestX", "a_b_c__d_.golden"), p)
	require.False(t, strings.Contains(p, ":"))
}

func TestLineDiffOfTheSameTextIsEmpty(t *testing.T) {
	require.Empty(t, lineDiff("a\nb", "a\nb"))
	require.Contains(t, lineDiff("a\n", "a"), "line 2")
}
