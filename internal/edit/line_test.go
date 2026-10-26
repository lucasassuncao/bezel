package edit

import (
	"math"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

// at is s with the caret before grapheme i.
func at(s string, i int) Line { return NewLine(s).SetCursor(i) }

func TestGraphemesKeepWhatReadsAsOneCharacterTogether(t *testing.T) {
	require.Equal(t, []string{"a", "é", "👍🏽", "🇧🇷", "中"}, Graphemes("aé👍🏽🇧🇷中"))
	require.Empty(t, Graphemes(""))
}

func TestInsertAndDeleteMoveTheCaret(t *testing.T) {
	l := at("helo", 3).Insert("l")
	require.Equal(t, "hello", l.String())
	require.Equal(t, 4, l.Cursor())

	l = l.DeleteBack()
	require.Equal(t, "helo", l.String())
	require.Equal(t, 3, l.Cursor())

	l = l.DeleteForward()
	require.Equal(t, "hel", l.String())
	require.Equal(t, 3, l.Cursor())

	require.Equal(t, "x", at("x", 0).DeleteBack().String(), "nothing before the caret")
	require.Equal(t, "x", at("x", 1).DeleteForward().String(), "nothing after it")
}

func TestAWideGraphemeIsOneStep(t *testing.T) {
	l := NewLine("a👍🏽b").SetCursor(2).DeleteBack()
	require.Equal(t, "ab", l.String())
	require.Equal(t, 3, NewLine("中a").Width(0, 2))
	require.Equal(t, 0, NewLine("中a").IndexAt(1), "the second cell of a wide grapheme is still that grapheme")
	require.Equal(t, 1, NewLine("中a").IndexAt(2))
	require.Equal(t, 2, NewLine("中a").IndexAt(9), "past the end")
}

func TestWordsAreRunsBetweenSpaces(t *testing.T) {
	l := at("git commit  -m", 14)
	require.Equal(t, 12, l.WordLeft().Cursor())
	require.Equal(t, 4, l.WordLeft().WordLeft().Cursor())
	require.Equal(t, 3, at("git commit", 0).WordRight().Cursor())
	require.Equal(t, 10, at("git commit", 3).WordRight().Cursor(), "past the space, then the word")

	require.Equal(t, "git ", at("git commit", 10).DeleteWordBack().String())
	require.Equal(t, "git", at("git commit", 3).DeleteWordForward().String())
}

func TestKillsCutToTheEnds(t *testing.T) {
	require.Equal(t, "lo", at("hello", 3).KillToStart().String())
	require.Equal(t, "hel", at("hello", 3).KillToEnd().String())
}

func TestSplitAndJoinAreInverse(t *testing.T) {
	before, after := at("hello", 2).Split()
	require.Equal(t, "he", before.String())
	require.Equal(t, "llo", after.String())
	require.Equal(t, 0, after.Cursor())

	joined := before.Join(after)
	require.Equal(t, "hello", joined.String())
	require.Equal(t, 2, joined.Cursor(), "the caret where the two meet")
}

// An edit never writes into the slice an older copy still reads.
func TestEditsLeaveOlderCopiesAlone(t *testing.T) {
	old := at("abcd", 2)
	_ = old.Insert("X")
	_ = old.DeleteForward()
	_ = old.DeleteBack()
	require.Equal(t, "abcd", old.String())
	require.Equal(t, 2, old.Cursor())
}

// A combining mark typed after a letter becomes part of it: the line is one
// grapheme, as a terminal draws it, and deleting takes both.
func TestAnEditJoinsWhatNowReadsAsOne(t *testing.T) {
	l := NewLine("e").Insert("\u0301")
	require.Equal(t, 1, l.Len())
	require.Equal(t, 1, l.Cursor())
	require.Equal(t, "", l.DeleteBack().String())
}

func TestPrintableKeepsTheJoinersEmojiNeed(t *testing.T) {
	require.Equal(t, "ab", Printable("a\x07\u202eb\u0601"), "a bell, a bidi override, a prepended sign")
	family := "👩\u200d👩\u200d👧"
	require.Equal(t, family, Printable(family))
}

func TestSliceAndAtStayInRange(t *testing.T) {
	l := NewLine("abc")
	require.Equal(t, "bc", l.Slice(1, 99))
	require.Equal(t, "", l.Slice(2, 1))
	require.Equal(t, "", l.At(3))
	require.Equal(t, 0, l.SetCursor(-5).Cursor())
}

func FuzzLineEditsNeverPanic(f *testing.F) {
	f.Add("hello world", 3, "x")
	f.Add("中👍🏽 a", 1, "é")
	f.Fuzz(func(t *testing.T, s string, cursor int, ins string) {
		l := NewLine(s).SetCursor(cursor)
		for _, op := range []func(Line) Line{
			Line.DeleteBack, Line.DeleteForward, Line.WordLeft, Line.WordRight,
			Line.DeleteWordBack, Line.DeleteWordForward, Line.KillToStart, Line.KillToEnd,
		} {
			got := op(l.Insert(ins))
			if got.Cursor() < 0 || got.Cursor() > got.Len() {
				t.Fatalf("caret %d outside 0..%d", got.Cursor(), got.Len())
			}
			// The offsets the line keeps are the graphemes its text splits into.
			g := Graphemes(got.String())
			if got.Len() != len(g) {
				t.Fatalf("%d graphemes kept, the text has %d", got.Len(), len(g))
			}
			for i, c := range g {
				if got.At(i) != c {
					t.Fatalf("grapheme %d is %q, the text's is %q", i, got.At(i), c)
				}
			}
		}
	})
}

// Split's halves are each a whole line of their own: joined, they are the
// line again, and an edit to one leaves the other as it was.
func TestSplitHalvesStandAlone(t *testing.T) {
	before, after := at("héllo wörld", 5).Split()
	require.Equal(t, "héllo", before.String())
	require.Equal(t, " wörld", after.String())
	require.Equal(t, "héllo!", before.Insert("!").String())
	require.Equal(t, " wörld", after.String())
	require.Equal(t, 5, NewLine("").Join(before).Join(after).Cursor(), "the caret where the last two met")
}

// The ASCII shortcut answers as the full measure does.
func TestCellsAgreesWithTheFullMeasure(t *testing.T) {
	for _, g := range []string{"a", " ", "~", "中", "👍🏽", "é", "\x7f", ""} {
		require.Equal(t, ansi.StringWidth(g), Cells(g), "%q", g)
	}
}

// An offset is kept in four bytes: one outside them is clamped, never wrapped.
func TestAt32ClampsWhatFourBytesCannotHold(t *testing.T) {
	require.Equal(t, int32(7), at32(7))
	require.Equal(t, int32(0), at32(-1))
	require.Equal(t, int32(math.MaxInt32), at32(math.MaxInt32+1))
}

func TestMapWordChangesTheWordAfterTheCaret(t *testing.T) {
	l := NewLine("hello  big world").SetCursor(5)
	got := l.MapWord(strings.ToUpper)
	require.Equal(t, "hello  BIG world", got.String())
	require.Equal(t, 10, got.Cursor(), "after the word")
	require.Equal(t, "end", NewLine("end").MapWord(strings.ToUpper).String(), "no word after the caret")
}

func TestTransposeSwapsAroundTheCaret(t *testing.T) {
	require.Equal(t, "bac", NewLine("abc").SetCursor(1).Transpose().String())
	l := NewLine("abc").Transpose()
	require.Equal(t, "acb", l.String(), "at the end, the last two")
	require.Equal(t, 3, l.Cursor())
	require.Equal(t, "a", NewLine("a").Transpose().String())
	require.Equal(t, "ab", NewLine("ab").SetCursor(0).Transpose().String())
}

func TestWordAtIsTheWordUnderTheCaret(t *testing.T) {
	require.Equal(t, "big", NewLine("the big dog").SetCursor(5).WordAt())
	require.Equal(t, "the", NewLine("the big dog").SetCursor(3).WordAt(), "touching its end")
	require.Empty(t, NewLine("a  b").SetCursor(2).WordAt())
}
