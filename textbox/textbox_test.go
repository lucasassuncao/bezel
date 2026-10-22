package textbox

import (
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

func th() theme.Resolved { return theme.Resolve(theme.Plain, true) }

func typed(s string) tea.KeyPressMsg { return tea.KeyPressMsg{Code: []rune(s)[0], Text: s} }

func press(code rune, mod tea.KeyMod) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code, Mod: mod} }

func focused(value string) Model {
	m, _ := New(th()).WithValue(value).Focus()
	return m
}

// run sends each message in turn, dropping the blink commands.
func run(m Model, msgs ...tea.Msg) Model {
	for _, msg := range msgs {
		m, _ = m.Update(msg)
	}
	return m
}

func viewRows(m Model) []string { return strings.Split(ansi.Strip(m.View()), "\n") }

func TestALongValueIsKeptWhole(t *testing.T) {
	long := strings.Repeat("0123456789\n", 150) + strings.Repeat("x", 1000)
	require.Equal(t, long, New(th()).SetValue(long).Value())
}

func TestSetValueFoldsEveryLineBreak(t *testing.T) {
	m := New(th()).SetValue("a\r\nb\rc\nd\te")
	require.Equal(t, "a\nb\nc\nd    e", m.Value())
	require.Equal(t, 4, m.LineCount())
}

func TestRowsCarryTheSharedGutter(t *testing.T) {
	m := New(th()).SetWidth(40).SetHeight(3).SetValue("first")
	rows := viewRows(m)
	require.Len(t, rows, 3)
	require.True(t, strings.HasPrefix(rows[0], ansi.Strip(draw.Gutter(1, th().Muted))+"first"), rows[0])
	require.Equal(t, "   ~ │", strings.TrimRight(rows[1], " "), "rows past the end are marked")
	for _, r := range rows {
		require.Equal(t, 40, ansi.StringWidth(r))
	}
}

func TestAPasteKeepsOneLinePerLine(t *testing.T) {
	m := run(focused(""), tea.PasteMsg{Content: "a\r\nb\r\nc"})
	require.Equal(t, "a\nb\nc", m.Value())
	require.Equal(t, 2, m.Line())
	require.Equal(t, 1, m.Column())
}

func TestTypingAndEnterEditAtTheCaret(t *testing.T) {
	m := run(focused("helo"), press(tea.KeyLeft, 0), typed("l"), press(tea.KeyEnd, 0), press(tea.KeyEnter, 0), typed("x"))
	require.Equal(t, "hello\nx", m.Value())
	m = run(m, press(tea.KeyUp, 0), press(tea.KeyHome, 0), press(tea.KeyRight, 0), press(tea.KeyEnter, 0))
	require.Equal(t, "h\nello\nx", m.Value(), "enter splits the line at the caret")
	require.Equal(t, []int{1, 0}, []int{m.Line(), m.Column()})
}

func TestBackspaceAndDeleteJoinLines(t *testing.T) {
	m := run(focused("ab\ncd").MoveTo(1, 0), press(tea.KeyBackspace, 0))
	require.Equal(t, "abcd", m.Value())
	require.Equal(t, []int{0, 2}, []int{m.Line(), m.Column()}, "the caret where they meet")

	m = run(focused("ab\ncd").MoveTo(0, 2), press(tea.KeyDelete, 0))
	require.Equal(t, "abcd", m.Value())
	m = run(focused("ab\ncd").MoveTo(1, 0), press('w', tea.ModCtrl))
	require.Equal(t, "abcd", m.Value(), "a word kill at a line's start joins it too")
}

func TestArrowsCrossLineEnds(t *testing.T) {
	m := run(focused("ab\ncd").MoveTo(1, 0), press(tea.KeyLeft, 0))
	require.Equal(t, []int{0, 2}, []int{m.Line(), m.Column()})
	m = run(m, press(tea.KeyRight, 0))
	require.Equal(t, []int{1, 0}, []int{m.Line(), m.Column()})
}

// Up and down aim for the column the move began at, through shorter lines.
func TestUpAndDownKeepTheColumn(t *testing.T) {
	m := focused("long line\nab\nlong line").MoveTo(0, 7)
	m = run(m, press(tea.KeyDown, 0))
	require.Equal(t, []int{1, 2}, []int{m.Line(), m.Column()}, "a short line: its end")
	m = run(m, press(tea.KeyDown, 0))
	require.Equal(t, []int{2, 7}, []int{m.Line(), m.Column()}, "back to the column aimed for")
	m = run(m, typed("x"), press(tea.KeyUp, 0), press(tea.KeyUp, 0))
	require.Equal(t, []int{0, 8}, []int{m.Line(), m.Column()}, "typing set a new aim")
}

func TestALongLineWrapsAtASpace(t *testing.T) {
	m := New(th()).SetWidth(draw.GutterWidth + 11).SetValue("alpha beta gamma")
	rows := viewRows(m)
	require.Len(t, rows, 2)
	require.Equal(t, "   1 │ alpha beta", strings.TrimRight(rows[0], " "))
	require.Equal(t, "     │ gamma", strings.TrimRight(rows[1], " "), "a continuation has no number")
}

func TestDownWalksTheWrappedRows(t *testing.T) {
	m := focused("alpha beta gamma\nend").SetWidth(draw.GutterWidth+11).MoveTo(0, 2)
	m = run(m, press(tea.KeyDown, 0))
	require.Equal(t, []int{0, 13}, []int{m.Line(), m.Column()}, "the wrapped row, same column")
	m = run(m, press(tea.KeyDown, 0))
	require.Equal(t, []int{1, 2}, []int{m.Line(), m.Column()})
}

func TestTheWindowFollowsTheCaret(t *testing.T) {
	m := focused(strings.Repeat("line\n", 19) + "last").SetHeight(5)
	require.Contains(t, viewRows(m)[4], "last", "the caret at the end shows")
	m = m.MoveTo(0, 0)
	require.Contains(t, viewRows(m)[0], "   1 │")
	m = run(m, press(tea.KeyPgDown, 0))
	require.Equal(t, 4, m.Line(), "a page is the height less one row")
	m = m.SetValue("short")
	require.Contains(t, viewRows(m)[0], "short", "a shorter text never leaves the window past its end")
}

func TestHeightIsLeftToTheCaller(t *testing.T) {
	m := New(th()).SetHeight(7).SetValue(strings.Repeat("line\n", 20))
	require.Equal(t, 7, m.Height())
	require.Len(t, viewRows(m), 7)
	require.Len(t, viewRows(m.SetHeight(0)), 21, "no height: every row")
}

func TestTheCaretShowsOnlyWhileFocused(t *testing.T) {
	m := focused("ab").MoveTo(0, 1)
	require.Contains(t, m.View(), th().Caret.Render("b"))
	require.NotContains(t, m.Blur().View(), th().Caret.Render("b"))
	require.Contains(t, m.Blur().View(), th().Dim.Render("ab"), "a blurred field is dimmed")
}

func TestAnUnfocusedFieldTakesNothing(t *testing.T) {
	m, cmd := New(th()).Update(typed("a"))
	require.Empty(t, m.Value())
	require.Nil(t, cmd)
}

func TestTheCaretBlinksAtTheThemesPace(t *testing.T) {
	short := th()
	short.CaretBlink = time.Nanosecond
	m, cmd := New(short).Focus()
	m, next := m.Update(cmd())
	require.NotNil(t, next)
	require.NotContains(t, m.View(), short.Caret.Render(" "), "the dark half")
	require.Nil(t, m.Blur().Blink(), "a blurred field does not blink")
	require.NotNil(t, m.Blink())
}

func TestEditsLeaveTheOldFieldAlone(t *testing.T) {
	before := focused("ab\ncd")
	after := run(before, press(tea.KeyBackspace, 0), press(tea.KeyEnter, 0))
	require.Equal(t, "ab\ncd", before.Value())
	require.NotEqual(t, before.Value(), after.Value())
}

func FuzzTextboxStaysInItsBox(f *testing.F) {
	f.Add("alpha beta\n中文 text\n", 20, 4, "jkhl")
	f.Fuzz(func(t *testing.T, s string, width, height int, keys string) {
		width = (width%60+60)%60 + 1
		height = (height%12+12)%12 + 1
		m := focused(s).SetWidth(width).SetHeight(height)
		moves := map[rune]tea.KeyPressMsg{
			'j': press(tea.KeyDown, 0), 'k': press(tea.KeyUp, 0), 'h': press(tea.KeyLeft, 0),
			'l': press(tea.KeyRight, 0), 'b': press(tea.KeyBackspace, 0), 'n': press(tea.KeyEnter, 0),
		}
		for _, r := range keys {
			if k, ok := moves[r]; ok {
				m, _ = m.Update(k)
			} else {
				m, _ = m.Update(typed(string(r)))
			}
		}
		rows := strings.Split(m.View(), "\n")
		if len(rows) != height {
			t.Fatalf("%d rows, want %d", len(rows), height)
		}
		if width > draw.GutterWidth {
			for _, r := range rows {
				if w := ansi.StringWidth(r); w != width {
					t.Fatalf("row is %d cells, want %d: %q", w, width, ansi.Strip(r))
				}
			}
		}
		if m.Line() < 0 || m.Line() >= m.LineCount() {
			t.Fatalf("caret on line %d of %d", m.Line(), m.LineCount())
		}
		for i, l := range m.lines {
			if want := segments(l.Line, m.textWidth()); !slices.Equal(l.starts, want) {
				t.Fatalf("line %d is cached as wrapping at %v, wraps at %v", i, l.starts, want)
			}
		}
	})
}

// A form grows the field from Rows: wrapped lines count, as the field draws them.
func TestRowsCountsTheWrappedRows(t *testing.T) {
	m := New(th()).SetValue("alpha beta gamma\nend")
	require.Equal(t, 2, m.Rows(), "no width: one row a line")
	require.Equal(t, 3, m.SetWidth(draw.GutterWidth+11).Rows())
}

// longText is lines lines of prose, every third long enough to wrap at 80.
func longText(lines int) string {
	var b strings.Builder
	for i := range lines {
		b.WriteString("key: value with some words in it")
		if i%3 == 0 {
			b.WriteString(" and a tail long enough to wrap past eighty columns of text")
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// benchField is a focused field of lines lines, 80 by 40, the caret mid-text.
func benchField(lines int) Model {
	m, _ := New(th()).SetWidth(80).SetHeight(40).SetValue(longText(lines)).Focus()
	return m.MoveTo(lines/2, 10)
}

// A keystroke is an Update and the frame it causes.
func BenchmarkTypingInALongText(b *testing.B) {
	m := benchField(5000)
	for b.Loop() {
		m, _ = m.Update(typed("x"))
		_ = m.View()
	}
}

func BenchmarkMovingDownALongText(b *testing.B) {
	m := benchField(5000)
	down := press(tea.KeyDown, 0)
	for b.Loop() {
		m, _ = m.Update(down)
		_ = m.View()
		if m.Line() == m.LineCount()-1 {
			m = m.MoveTo(0, 10)
		}
	}
}

// ctrl+v reads the clipboard off the loop and comes back as a paste, its
// lines kept.
func TestCtrlVPastesTheClipboard(t *testing.T) {
	m := focused("").WithClipboard(func() (string, error) { return "a\r\nb", nil })
	m, cmd := m.Update(press('v', tea.ModCtrl))
	require.Empty(t, m.Value(), "nothing until the read answers")
	m = run(m, cmd())
	require.Equal(t, "a\nb", m.Value())
	require.Equal(t, []int{1, 1}, []int{m.Line(), m.Column()})
}
