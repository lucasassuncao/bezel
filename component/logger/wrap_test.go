package logger

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapKeepsEveryCharacter(t *testing.T) {
	for _, s := range []string{
		`C:\Users\lucas\Downloads\Installers\exe\tor-browser-windows-x86_64-portable-15.0.26.exe`,
		strings.Repeat("abcdefghij", 9),
		"one two three four five six seven eight nine ten",
		"ação → 名前 名前 名前 名前 名前 名前 名前 名前",
	} {
		lines := wrap(s, 16)
		require.Equal(t, s, strings.Join(lines, ""))
		for _, l := range lines {
			require.LessOrEqual(t, len([]rune(l)), 16, l)
		}
	}
}

func TestWrapPrefersAPathSeparator(t *testing.T) {
	require.Equal(t, []string{`/usr/local/`, `bin/bezel`}, wrap(`/usr/local/bin/bezel`, 14))
}

func TestWrapFallsBackToAWeakSeparator(t *testing.T) {
	require.Equal(t, []string{"go1.27.3-", "windows"}, wrap("go1.27.3-windows", 12))
}

func TestWrapCutsWhereNothingBreaks(t *testing.T) {
	require.Equal(t, []string{"abcdefghijkl", "mno"}, wrap("abcdefghijklmno", 12))
}

// A separator in the first third would leave a sliver, so the cut is at room.
func TestWrapSkipsAnEarlySeparator(t *testing.T) {
	require.Equal(t, []string{"a/bcdefghijklmn", "op"}, wrap("a/bcdefghijklmnop", 15))
}

func TestWrapHoldsAMinimumRoom(t *testing.T) {
	require.Equal(t, []string{"abcdefghijkl", "m"}, wrap("abcdefghijklm", 3))
}

func TestWrapSplitsNewlinesWithoutRoom(t *testing.T) {
	require.Equal(t, []string{"a", "b"}, wrap("a\nb", 0))
}

func TestWrapCountsWideRunes(t *testing.T) {
	lines := wrap(strings.Repeat("名", 10), 12)
	require.Equal(t, []string{strings.Repeat("名", 6), strings.Repeat("名", 4)}, lines)
}
