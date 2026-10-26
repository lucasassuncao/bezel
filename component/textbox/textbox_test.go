package textbox

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/cursor"
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

// A keystroke is an Update and the frame it causes. Each loop types and
// deletes, so the text is the same on every pass whatever b.N is.
func BenchmarkTypingInALongText(b *testing.B) {
	m := benchField(5000)
	back := press(tea.KeyBackspace, 0)
	for b.Loop() {
		m, _ = m.Update(typed("x"))
		_ = m.View()
		m, _ = m.Update(back)
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

// keysBold makes every YAML-ish key bold, counting its own calls.
func keysBold(calls *int) func(string) [][]Span {
	return func(v string) [][]Span {
		*calls++
		var out [][]Span
		for _, l := range strings.Split(v, "\n") {
			var spans []Span
			if i := strings.Index(l, ":"); i > 0 {
				spans = append(spans, Span{Start: 0, End: i, Style: lipgloss.NewStyle().Bold(true)})
			}
			out = append(out, spans)
		}
		return out
	}
}

func TestHighlightPaintsTheSpansItIsGiven(t *testing.T) {
	bold := lipgloss.NewStyle().Bold(true)
	calls := 0
	m, _ := New(th()).WithHighlight(keysBold(&calls)).WithValue("café: x\nkey: y").Focus()
	require.Contains(t, m.View(), bold.Render("café"), "spans are bytes; é is two")
	require.Contains(t, m.View(), bold.Render("key"))

	before := calls
	m = run(m, press(tea.KeyUp, 0), press(tea.KeyHome, 0))
	require.Equal(t, before, calls, "moving the caret does not run the highlighter")
	require.Contains(t, m.View(), bold.Render("afé"), "the caret takes its cell over the span")

	m = run(m, typed("x"))
	require.Equal(t, before+1, calls, "an edit runs it again")
	require.Contains(t, m.View(), bold.Render("x"), "the new text is painted too")
	require.Equal(t, "xcafé: x\nkey: y", m.Value())
}

func TestABlurredFieldIsDimmedNotHighlighted(t *testing.T) {
	calls := 0
	m := New(th()).WithHighlight(keysBold(&calls)).WithValue("key: y")
	require.NotContains(t, m.View(), lipgloss.NewStyle().Bold(true).Render("key"))
}

// A field never built with New is an empty one, as an input's zero value is.
func TestTheZeroFieldIsAnEmptyOne(t *testing.T) {
	var m Model
	require.Equal(t, 1, m.LineCount())
	require.Equal(t, 0, m.Column())
	require.Equal(t, 1, m.Rows())
	m = m.SetWidth(20).SetHeight(3).MoveTo(4, 4)
	require.NotPanics(t, func() { _ = m.View() })

	m, _ = Model{}.Focus()
	m = run(m, typed("a"), press(tea.KeyEnter, 0), typed("b"), press(tea.KeyUp, 0))
	require.Equal(t, "a\nb", m.Value())
	require.Equal(t, 0, m.Line())
}

func TestShiftMovesSelect(t *testing.T) {
	m := focused("one two\nthree")
	m = run(m, press(tea.KeyUp, 0), press(tea.KeyHome, 0), press(tea.KeyRight, tea.ModShift), press(tea.KeyRight, tea.ModShift))
	require.True(t, m.HasSelection())
	require.Equal(t, "on", m.SelectedText())

	m = run(m, press(tea.KeyDown, tea.ModShift))
	require.Equal(t, "one two\nth", m.SelectedText(), "down carries the selection to the next line")
	start, end := m.Selection()
	require.Equal(t, Position{0, 0}, start)
	require.Equal(t, Position{1, 2}, end)

	m = run(m, press(tea.KeyLeft, 0))
	require.False(t, m.HasSelection(), "a plain move drops it")
}

func TestTypingReplacesTheSelectionAndDeletingTakesIt(t *testing.T) {
	m := focused("hello world")
	m = run(m, press(tea.KeyLeft, tea.ModCtrl|tea.ModShift))
	require.Equal(t, "world", m.SelectedText())
	m = run(m, typed("m"), typed("e"))
	require.Equal(t, "hello me", m.Value())

	m = run(m, press('a', tea.ModAlt))
	require.Equal(t, "hello me", m.SelectedText(), "alt+a takes everything")
	m = run(m, press(tea.KeyBackspace, 0))
	require.Empty(t, m.Value())

	m = run(focused("ab\ncd"), press('a', tea.ModAlt), tea.PasteMsg{Content: "x"})
	require.Equal(t, "x", m.Value(), "a paste replaces it too")
}

func TestSelectAndClearSelectionFromTheApp(t *testing.T) {
	m := focused("abc\ndef").Select(Position{0, 1}, Position{1, 1})
	require.Equal(t, "bc\nd", m.SelectedText())
	require.Equal(t, 1, m.Line())
	require.Contains(t, m.View(), th().Selected.Render("bc"))
	require.False(t, m.ClearSelection().HasSelection())
	require.Empty(t, m.SetValue("x").SelectedText(), "a new value has nothing selected")
	require.Empty(t, Model{}.SelectedText(), "the zero field neither")
}

func TestAltWCopiesTheSelectionAndAltXCutsIt(t *testing.T) {
	var board string
	m := focused("copy me").WithClipboardWrite(func(s string) error { board = s; return nil })
	_, cmd := m.Update(press('w', tea.ModAlt))
	require.Nil(t, cmd, "nothing selected, nothing copied")

	m = run(m, press(tea.KeyLeft, tea.ModCtrl|tea.ModShift))
	_, cmd = m.Update(press('w', tea.ModAlt))
	require.Equal(t, CopiedMsg{Text: "me"}, cmd())
	require.Equal(t, "me", board)

	board = ""
	m, cmd = m.Update(press('x', tea.ModAlt))
	require.Equal(t, "copy ", m.Value())
	for _, c := range cmd().(tea.BatchMsg) {
		if c, ok := c().(CopiedMsg); ok {
			require.Equal(t, "me", c.Text)
		}
	}
	require.Equal(t, "me", board)
}

func TestTheCharLimitCountsLineBreaks(t *testing.T) {
	m := focused("").WithCharLimit(5)
	m = run(m, typed("a"), typed("b"), press(tea.KeyEnter, 0), typed("c"), typed("d"), typed("e"))
	require.Equal(t, "ab\ncd", m.Value())
	m = run(m, press(tea.KeyBackspace, 0), tea.PasteMsg{Content: "中xyz"})
	require.Equal(t, "ab\nc中", m.Value(), "characters, not bytes")
	require.Equal(t, "abcde", New(th()).WithCharLimit(5).WithValue("abcdefgh").Value())
	require.Equal(t, 5, m.CharLimit())
}

func TestMaxLinesStopsNewLines(t *testing.T) {
	m := focused("one").WithMaxLines(2)
	m = run(m, press(tea.KeyEnter, 0), typed("x"), press(tea.KeyEnter, 0), typed("y"))
	require.Equal(t, "one\nxy", m.Value())
	m = run(m, tea.PasteMsg{Content: "z\nlost"})
	require.Equal(t, "one\nxyz", m.Value(), "a paste is cut at the last line that fits")
	require.Equal(t, "a\nb", New(th()).WithMaxLines(2).WithValue("a\nb\nc").Value())
	require.Equal(t, 2, m.MaxLines())
}

func TestADynamicHeightFollowsTheText(t *testing.T) {
	m := focused("").SetWidth(20).WithDynamicHeight(2, 4)
	require.Equal(t, 2, m.Height(), "never under the least")
	require.Len(t, viewRows(m), 2)
	m = run(m, press(tea.KeyEnter, 0), press(tea.KeyEnter, 0))
	require.Equal(t, 3, m.Height())
	m = run(m, press(tea.KeyEnter, 0), press(tea.KeyEnter, 0), press(tea.KeyEnter, 0))
	require.Equal(t, 4, m.Height(), "nor over the most: it scrolls")
	require.Len(t, viewRows(m), 4)
	m = run(m, press('a', tea.ModAlt), press(tea.KeyBackspace, 0))
	require.Equal(t, 2, m.Height(), "it shrinks back")
	require.Equal(t, 7, m.SetHeight(7).Height(), "SetHeight ends it")
	require.Equal(t, 0, m.WithDynamicHeight(0, 9).SetHeight(0).Height())
}

func TestThePlaceholderShowsWhileEmpty(t *testing.T) {
	m := focused("").SetWidth(20).SetHeight(2).WithPlaceholder("write here")
	rows := viewRows(m)
	require.Contains(t, rows[0], "write here")
	require.Equal(t, 20, ansi.StringWidth(rows[0]))
	require.Equal(t, "write here", m.Placeholder())
	require.NotContains(t, viewRows(run(m, typed("x")))[0], "write here")
}

func TestThePromptLeadsEveryLineAndTheNumbersCanGo(t *testing.T) {
	m := New(th()).SetWidth(12).WithLineNumbers(false).WithPrompt("> ").SetValue("ab\ncd")
	require.Equal(t, []string{"> ab        ", "> cd        "}, viewRows(m))

	m = m.WithPromptFunc(3, func(i int) string { return fmt.Sprint(i+1, ":") }).SetValue("one two three")
	rows := viewRows(m)
	require.Equal(t, "1: one two", strings.TrimRight(rows[0], " "), "the text wraps in what the prompt leaves")
	require.True(t, strings.HasPrefix(rows[1], "   three"), "a wrapped row gets a blank prompt")
	for _, r := range rows {
		require.Equal(t, 12, ansi.StringWidth(r))
	}
}

func TestWordKeysChangeCaseAndCtrlTTransposes(t *testing.T) {
	m := run(focused("make it loud"), press(tea.KeyHome, 0), press('u', tea.ModAlt), press('c', tea.ModAlt))
	require.Equal(t, "MAKE It loud", m.Value())
	m = run(m, press('l', tea.ModAlt))
	require.Equal(t, "MAKE It loud", m.Value(), "the next word, loud, was lower already")
	m = run(m, press(tea.KeyHome, 0), press('l', tea.ModAlt))
	require.Equal(t, "make It loud", m.Value())
	require.Equal(t, "ba", run(focused("ab"), press('t', tea.ModCtrl)).Value())
}

func TestInsertStringAndRuneTypeAtTheCaret(t *testing.T) {
	m := focused("ad").MoveTo(0, 1).InsertString("b\nc").InsertRune('!')
	require.Equal(t, "ab\nc!d", m.Value())
	require.Equal(t, 1, m.Line())
	m = m.Select(Position{0, 0}, Position{1, 0}).InsertString("x")
	require.Equal(t, "xc!d", m.Value(), "in place of the selection")
	require.Equal(t, "xc", New(th()).WithCharLimit(2).InsertString("xcz").Value())
	require.Equal(t, "xc!d", m.MoveTo(0, 2).Word())
}

func TestCursorModesAndTheTerminalCursor(t *testing.T) {
	m, cmd := New(th()).WithCursorMode(cursor.Steady).Focus()
	require.Nil(t, cmd)
	require.Nil(t, m.Blink())
	m, cmd = m.SetCursorMode(cursor.Blinking)
	require.NotNil(t, cmd)
	require.Equal(t, cursor.Blinking, m.CursorMode())

	m, cmd = New(th()).SetWidth(20).SetHeight(2).WithTerminalCursor().SetValue("a\nb\nxyz").MoveTo(2, 1).Focus()
	require.Nil(t, cmd, "the terminal blinks its own")
	c := m.TerminalCursor()
	require.Equal(t, draw.GutterWidth+1, c.X)
	require.Equal(t, 1, c.Y, "the window scrolled one row down")
	require.Nil(t, m.Blur().TerminalCursor())
	require.Nil(t, focused("x").TerminalCursor())
}
