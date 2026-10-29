package viewport

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Mint, true)

func numbered(n int) string {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	return strings.Join(lines, "\n")
}

func sized(content string, w, h int) Model { return New().SetWidth(w).SetHeight(h).SetContent(content) }

func rows(m Model) []string { return strings.Split(ansi.Strip(m.View(th)), "\n") }

func TestANewViewportDrawsNothingUntilSized(t *testing.T) {
	require.Equal(t, "", New().SetContent("a").View(th))
}

func TestTheViewIsExactlyItsSize(t *testing.T) {
	r := rows(sized("short\n"+strings.Repeat("x", 30), 10, 4))
	require.Equal(t, []string{"short     ", "xxxxxxxxxx", "          ", "          "}, r)
}

func TestScrollingMovesByLinePageAndHalfPage(t *testing.T) {
	m := sized(numbered(40), 20, 10)
	require.Equal(t, 1, m.ScrollDown(1).YOffset())
	require.Equal(t, 10, m.PageDown().YOffset())
	require.Equal(t, 5, m.HalfPageDown().YOffset())
	require.Equal(t, 30, m.ToBottom().YOffset())
	require.True(t, m.ToBottom().AtBottom())
	require.Equal(t, 30, m.ToBottom().ScrollDown(5).YOffset(), "past the end does nothing")
	require.Equal(t, 29, m.ToBottom().ScrollUp(1).YOffset())
	require.True(t, m.AtTop())
}

func TestTheKeysScroll(t *testing.T) {
	m := sized(numbered(40), 20, 10)
	for k, want := range map[string]int{"down": 1, "pgdown": 10, "end": 30} {
		got, _ := m.Update(bezeltest.Key(k))
		require.Equal(t, want, got.YOffset(), k)
	}
	m = m.ToBottom()
	for k, want := range map[string]int{"up": 29, "pgup": 20, "home": 0} {
		got, _ := m.Update(bezeltest.Key(k))
		require.Equal(t, want, got.YOffset(), k)
	}
}

func TestLettersDoNotScroll(t *testing.T) {
	m := sized(numbered(40)+"\n"+strings.Repeat("x", 50), 20, 10).SetYOffset(5)
	for _, k := range []string{"j", "k", "h", "l", "g", "G", "space", "f", "b", "d", "u", "ctrl+d", "ctrl+u"} {
		got, _ := m.Update(bezeltest.Key(k))
		require.Equal(t, []int{5, 0}, []int{got.YOffset(), got.XOffset()}, k)
	}
}

func TestHalfPageKeysCanBeBound(t *testing.T) {
	k := DefaultKeys()
	k.HalfPageDown = key.New("ctrl+d")
	got, _ := sized(numbered(40), 20, 10).WithKeys(k).Update(bezeltest.Key("ctrl+d"))
	require.Equal(t, 5, got.YOffset())
}

func TestPanningCutsLongLinesAtTheOffset(t *testing.T) {
	m := sized("0123456789abcdef", 6, 1)
	require.Equal(t, "012345", rows(m)[0])
	m, _ = m.Update(bezeltest.Key("right"))
	require.Equal(t, "6789ab", rows(m)[0], "a pan is six cells by default")
	require.Equal(t, "abcdef", rows(m.SetXOffset(99))[0], "the pan stops where the longest line ends")
	require.Equal(t, 0, m.SetWrap(true).XOffset(), "wrapping leaves nothing to pan")
	still, _ := sized("0123456789abcdef", 6, 1).WithStep(0).Update(bezeltest.Key("right"))
	require.Equal(t, 0, still.XOffset(), "a step of zero stops the keys panning")
}

func TestWrapFoldsLongLinesIntoRows(t *testing.T) {
	m := sized("0123456789\nab", 4, 5).SetWrap(true)
	require.Equal(t, []string{"0123", "4567", "89  ", "ab  ", "    "}, rows(m))
	require.Equal(t, 4, m.TotalRows())
	require.Equal(t, 4, m.VisibleRows())
}

func TestSetWrapKeepsTheTopLineInView(t *testing.T) {
	m := sized(strings.Repeat("0123456789\n", 10), 5, 3).SetWrap(true).SetYOffset(6)
	require.Equal(t, 3, m.SetWrap(false).YOffset(), "row 6 was line 3's first row")
}

func TestAResizeLeavesTheOffsetPastTheBottom(t *testing.T) {
	m := sized(numbered(40), 20, 10).ToBottom().SetHeight(20)
	require.True(t, m.PastBottom())
	require.Equal(t, 20, m.YOffset(), "reads clamp to the last full page")
	require.Equal(t, 19, m.ScrollUp(1).YOffset(), "the first key back moves at once")
}

func TestSetContentKeepsTheOffsetWithinTheText(t *testing.T) {
	m := sized(numbered(40), 20, 10).ToBottom().SetContent(numbered(12))
	require.Equal(t, 2, m.YOffset())
	require.False(t, m.PastBottom())
}

func TestContentFoldsLineEndingsAndTabs(t *testing.T) {
	m := New().SetLines([]string{"a\r\nb", "c\rd", "x\ty", "1234\tz"})
	require.Equal(t, []string{"a", "b", "c", "d", "x   y", "1234    z"}, m.Lines())
	require.Equal(t, "a\nb\nc\nd\nx   y\n1234    z", m.Content())
}

func TestScrollPercent(t *testing.T) {
	m := sized(numbered(40), 20, 10)
	require.Equal(t, 0.0, m.ScrollPercent())
	require.Equal(t, 0.5, m.SetYOffset(15).ScrollPercent())
	require.Equal(t, 1.0, sized("a", 20, 10).ScrollPercent(), "everything shows")
	require.Equal(t, 0.5, sized("0123456789", 6, 1).SetXOffset(2).HorizontalScrollPercent())
}

func TestTheWheelScrollsThreeLinesAndShiftPans(t *testing.T) {
	m := sized(numbered(40)+"\n"+strings.Repeat("x", 50), 20, 10)
	got, _ := m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	require.Equal(t, 3, got.YOffset())
	got, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown, Mod: tea.ModShift})
	require.Equal(t, 6, got.XOffset())
	got, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelRight})
	require.Equal(t, 6, got.XOffset())
	got, _ = m.WithWheel(0).Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	require.Equal(t, 0, got.YOffset(), "a wheel of no lines is off")
}

func TestEnsureVisibleBringsALineAndItsColumnsIn(t *testing.T) {
	m := sized(numbered(40)+"\n"+strings.Repeat("x", 100), 20, 10).EnsureVisible(40, 60, 70)
	require.Equal(t, 31, m.YOffset())
	require.Equal(t, 54, m.XOffset(), "a step left of the match")
}

func TestScrollIsWhereTheWindowSits(t *testing.T) {
	require.Equal(t, layout.Scroll{Offset: 10, Visible: 10, Total: 40}, sized(numbered(40), 20, 10).PageDown().Scroll())
}

func TestLegendLeavesPanOutWhileWrapping(t *testing.T) {
	keys := func(m Model) (out []string) {
		for _, e := range m.Legend() {
			out = append(out, e.Help().Key)
		}
		return out
	}
	require.Equal(t, []string{"↑/↓", "←/→", "pgup/pgdn"}, keys(New()))
	require.Equal(t, []string{"↑/↓", "pgup/pgdn"}, keys(New().SetWrap(true)))
}

func TestStyledContentIsCutByItsCells(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("0123456789")
	require.Equal(t, "0123", rows(sized(red, 4, 1))[0])
	require.Equal(t, []string{"0123", "4567", "89  "}, rows(sized(red, 4, 3).SetWrap(true)))
}

func TestAWideGraphemeAtTheEdgeKeepsTheWidth(t *testing.T) {
	require.Equal(t, "ab ", rows(sized("ab🌍cd", 3, 1))[0])
	require.Equal(t, []string{"全 ", "角 ", "全 "}, rows(sized("全角全", 3, 3).SetWrap(true)))
}

func TestEmptyAndTinyViewportsNeverPanic(t *testing.T) {
	require.NotPanics(t, func() {
		for _, m := range []Model{New(), sized("", 5, 2), sized("\n", 5, 2), sized("abc", -1, -1), sized("abc", 1, 1).SetWrap(true)} {
			_ = m.View(th)
			_ = m.PageDown().ToBottom().ScrollRight(9).Scroll()
		}
	})
}

func TestWithStylesHoldsUnderAnotherTheme(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m := New().WithStyles(func(s Styles) Styles { s.Match = red; return s })
	require.Equal(t, red, m.styles(theme.Resolve(theme.Plain, true)).Match)
}

func FuzzViewportStaysInItsBox(f *testing.F) {
	f.Add("line one\nline two", 10, 3, false, 0, 0)
	f.Add("🌍ab\tc\x1b[31mred\x1b[0m", 4, 2, true, 1, 0)
	f.Add("全角全角全角\nab", 5, 4, true, 0, 0)
	f.Add("0\x1b", 4, 2, true, 0, 0) // a stray ESC swallowed the padding after it
	f.Add("\x1bX0", 4, 2, true, 0, 0)
	f.Add("0\xe800\xe5", 5, 4, false, 0, 0)
	f.Add("\x1b \x02", 31, 9, true, 1, 0)
	f.Fuzz(func(t *testing.T, content string, w, h int, wrap bool, y, x int) {
		w, h = w%40, h%20
		m := sized(content, w, h).SetWrap(wrap).SetYOffset(y).SetXOffset(x)
		view := m.View(th)
		if w <= 0 || h <= 0 {
			require.Equal(t, "", view)
			return
		}
		lines := strings.Split(view, "\n")
		require.Len(t, lines, h)
		for i, l := range lines {
			require.Equal(t, w, ansi.StringWidth(l), "row %d of %q", i, view)
		}
	})
}

func hundredThousand() string {
	line := strings.Repeat("0123456789", 12)
	lines := make([]string, 100_000)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func BenchmarkViewOverAHundredThousandLines(b *testing.B) {
	content := hundredThousand()
	for _, wrap := range []bool{false, true} {
		name := map[bool]string{false: "unwrapped", true: "wrapped"}[wrap]
		b.Run(name, func(b *testing.B) {
			m := sized(content, 120, 40).SetWrap(wrap).ToBottom()
			b.ReportAllocs()
			for b.Loop() {
				_ = m.View(th)
			}
		})
	}
}

func BenchmarkScrollingAHundredThousandLines(b *testing.B) {
	m := sized(hundredThousand(), 50, 40).SetWrap(true)
	b.ReportAllocs()
	for b.Loop() {
		if m.AtBottom() {
			m = m.ToTop()
		}
		m = m.PageDown()
	}
}

func TestStrayControlCharactersAreDropped(t *testing.T) {
	m := New().SetLines([]string{"0\x1b", "b\x07c\x00", "\x1b[31mred\x1b[0m", "é\u0301"})
	require.Equal(t, []string{"0", "bc", "\x1b[31mred\x1b[0m", "é\u0301"}, m.Lines())
}

func TestAnEscapeSequenceCutOffByTheLineIsDropped(t *testing.T) {
	require.Equal(t, []string{"", "a"}, New().SetLines([]string{"\x1bX0", "a\x1b[31"}).Lines())
}

func TestInvalidUTF8BecomesTheReplacementCharacter(t *testing.T) {
	require.Equal(t, []string{"0\uFFFD00\uFFFD"}, New().SetLines([]string{"0\xe800\xe5"}).Lines())
}

func TestAnEscapeWithoutItsFinalByteIsDropped(t *testing.T) {
	require.Equal(t, []string{""}, New().SetLines([]string{"\x1b \x02"}).Lines())
}

func numbers(r GutterRow) string {
	switch {
	case r.Line >= r.Lines:
		return "  ~ "
	case r.Continued:
		return "    "
	}
	return fmt.Sprintf("%3d ", r.Line+1)
}

func TestTheGutterNumbersRowsAndStaysWhilePanning(t *testing.T) {
	m := sized("0123456789\nab", 8, 2).WithGutter(numbers)
	require.Equal(t, []string{"  1 0123", "  2 ab  "}, rows(m))
	require.Equal(t, "  1 6789", rows(m.ScrollRight(6))[0])
}

func TestAWrappedLinesRowsCarryTheContinuedGutter(t *testing.T) {
	m := sized("0123456789", 8, 3).WithGutter(numbers).SetWrap(true)
	require.Equal(t, []string{"  1 0123", "    4567", "    89  "}, rows(m))
}

func TestFillPutsTheGutterOnRowsPastTheEnd(t *testing.T) {
	m := sized("a", 6, 3).WithGutter(numbers)
	require.Equal(t, []string{"  1 a ", "      ", "      "}, rows(m))
	require.Equal(t, []string{"  1 a ", "  ~   ", "  ~   "}, rows(m.WithFill(true)))
}

func TestAGutterWiderThanTheViewportLeavesNoText(t *testing.T) {
	require.Equal(t, []string{"  1 "}, rows(sized("abc", 4, 1).WithGutter(numbers)))
}

func TestLineStylePaintsWholeLines(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m := sized("a\nb", 1, 2).WithLineStyle(func(i int) lipgloss.Style {
		if i == 1 {
			return red
		}
		return lipgloss.NewStyle()
	})
	require.Equal(t, "a\n"+red.Render("b"), m.View(th))
}

func TestEnsureVisibleWaitsForASize(t *testing.T) {
	m := New().SetContent(numbered(40)+"\n"+strings.Repeat("x", 100)).EnsureVisible(40, 60, 70)
	m = m.SetWidth(20).SetHeight(10)
	require.Equal(t, []int{0, 0}, []int{m.YOffset(), m.XOffset()}, "nothing shows yet, so nothing scrolls")
}

// Letters scroll only once the app asks for vim's keys.
func TestVimKeysAddTheLetters(t *testing.T) {
	m := sized(numbered(40), 20, 10)
	got, _ := m.Update(bezeltest.Key("j"))
	require.Zero(t, got.YOffset(), "no letters by default")
	m = m.WithKeys(VimKeys())
	for k, want := range map[string]int{"j": 1, "ctrl+f": 10, "ctrl+d": 5, "G": 30, "down": 1} {
		got, _ := m.Update(bezeltest.Key(k))
		require.Equal(t, want, got.YOffset(), k)
	}
}
