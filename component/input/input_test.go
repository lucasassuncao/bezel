package input

import (
	"charm.land/lipgloss/v2"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/cursor"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Mint, true)

func typed(s string) tea.KeyPressMsg { return tea.KeyPressMsg{Code: []rune(s)[0], Text: s} }

func press(code rune, mod tea.KeyMod) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Mod: mod}
}

func focused() Model {
	m, _ := New(th).Focus()
	return m
}

// run types each key in turn, dropping the blink commands.
func run(m Model, msgs ...tea.Msg) Model {
	for _, msg := range msgs {
		m, _ = m.Update(msg)
	}
	return m
}

func TestTypingInsertsAtTheCaret(t *testing.T) {
	m := run(focused(), typed("h"), typed("i"), press(tea.KeyLeft, 0), typed("é"))
	require.Equal(t, "héi", m.Value())
	require.Equal(t, 2, m.Cursor())
}

func TestAnUnfocusedFieldTakesNothing(t *testing.T) {
	m, cmd := New(th).Update(typed("a"))
	require.Empty(t, m.Value())
	require.Nil(t, cmd)
	require.Empty(t, run(focused().Blur(), typed("a")).Value())
}

func TestTheReadlineKeysEdit(t *testing.T) {
	m := focused().WithValue("git commit -m")
	m = run(m, press('w', tea.ModCtrl))
	require.Equal(t, "git commit ", m.Value(), "ctrl+w eats the word before")
	m = run(m, press('a', tea.ModCtrl), press('f', tea.ModAlt), press('k', tea.ModCtrl))
	require.Equal(t, "git", m.Value(), "home, a word right, then kill to the end")
	m = run(m, press(tea.KeyLeft, 0), press('u', tea.ModCtrl))
	require.Equal(t, "t", m.Value())
	m = run(m, press(tea.KeyBackspace, 0), press(tea.KeyDelete, 0))
	require.Empty(t, m.Value(), "nothing before the caret, then the t after it")
}

func TestAKeyItDoesNotKnowIsLeftToTheHost(t *testing.T) {
	m := focused().WithValue("x")
	for _, k := range []tea.KeyPressMsg{press(tea.KeyEnter, 0), press(tea.KeyUp, 0), {Code: 'x', Text: "x", Mod: tea.ModCtrl}} {
		got, cmd := m.Update(k)
		require.Equal(t, "x", got.Value(), "%v", k)
		require.Nil(t, cmd, "%v", k)
	}
}

func TestAPasteIsOneLine(t *testing.T) {
	m := run(focused(), tea.PasteMsg{Content: "one\r\ntwo\tthree\x07"})
	require.Equal(t, "one two three", m.Value())
}

func TestTheMaskHidesTheTextAndItsWords(t *testing.T) {
	m := focused().WithMask('•').WithValue("hunter two")
	require.Equal(t, "••••••••••", ansi.Strip(m.Blur().View()))
	m = run(m, press(tea.KeyBackspace, tea.ModAlt))
	require.Empty(t, m.Value(), "a word kill runs to the start: no boundary shows")
}

func TestTheViewIsTheWidthSet(t *testing.T) {
	m := focused().WithWidth(8).WithValue("ab")
	require.Equal(t, "ab      ", ansi.Strip(m.View()))
	require.Contains(t, m.View(), th.Caret.Render(" "), "the caret on the cell after the text")
	require.Equal(t, "ab", ansi.Strip(m.Blur().WithWidth(0).View()), "no width, no caret: just the text")
}

func TestALongValueScrollsToKeepTheCaret(t *testing.T) {
	m := focused().WithWidth(5).WithValue("abcdefgh")
	require.Equal(t, "efgh ", ansi.Strip(m.View()), "the end and the caret's cell")

	m = run(m, press(tea.KeyHome, 0))
	require.Equal(t, "abcde", ansi.Strip(m.View()))
	m = run(m, press(tea.KeyRight, 0), press(tea.KeyRight, 0))
	require.Equal(t, "abcde", ansi.Strip(m.View()), "the window stays while the caret is in it")

	m = run(m, press(tea.KeyEnd, 0), press(tea.KeyBackspace, 0), press(tea.KeyBackspace, 0))
	require.Equal(t, "cdef ", ansi.Strip(m.View()), "a deletion pulls the window back")
}

func TestAWideCharacterNeverSplitsAtTheEdge(t *testing.T) {
	m := focused().WithWidth(4).WithValue("a中中")
	got := ansi.Strip(m.View())
	require.Equal(t, 4, ansi.StringWidth(got))
	require.Equal(t, "中  ", got, "a would need half of 中's room: the window starts at 中")
}

func TestThePlaceholderShowsWhileEmpty(t *testing.T) {
	m := focused().WithPlaceholder("name")
	require.Equal(t, "name", ansi.Strip(m.View()))
	require.Contains(t, m.View(), th.Caret.Render("n"), "the caret on its first letter")
	require.Equal(t, th.Muted.Render("name"), m.Blur().View())
	require.Equal(t, "x ", ansi.Strip(run(m, typed("x")).View()))
}

func TestThePromptLeads(t *testing.T) {
	require.Equal(t, ":cmd", ansi.Strip(New(th).WithPrompt(":").WithValue("cmd").View()))
}

func TestTheCaretBlinksOnItsOwnTicksOnly(t *testing.T) {
	short := th
	short.CaretBlink = time.Millisecond
	m, cmd := New(short).Focus()
	tick := cmd()
	require.IsType(t, BlinkMsg{}, tick)

	m, next := m.Update(tick)
	require.NotNil(t, next, "the blink runs on")
	require.NotContains(t, m.View(), short.Caret.Render(" "), "the dark half")

	m, _ = m.Update(tick)
	require.NotContains(t, m.View(), short.Caret.Render(" "), "a spent tick changes nothing")

	other, otherCmd := New(short).Focus()
	_ = other
	m, _ = m.Update(otherCmd())
	require.NotContains(t, m.View(), short.Caret.Render(" "), "another input's tick is not ours")

	m, _ = m.Update(typed("a"))
	require.Contains(t, m.View(), short.Caret.Render(" "), "typing lights the caret again")
}

func TestEditsLeaveTheOldFieldAlone(t *testing.T) {
	before := focused().WithValue("abc")
	after := run(before, press(tea.KeyBackspace, 0))
	require.Equal(t, "abc", before.Value())
	require.Equal(t, "ab", after.Value())
}

func FuzzInputNeverOutgrowsItsWidth(f *testing.F) {
	f.Add("hello 中👍🏽 world", 6, 3)
	f.Fuzz(func(t *testing.T, s string, width, left int) {
		width = (width%40+40)%40 + 1
		m := focused().WithWidth(width).WithValue(s)
		for range (left%50 + 50) % 50 {
			m, _ = m.Update(press(tea.KeyLeft, 0))
		}
		if got := ansi.StringWidth(m.View()); got != width {
			t.Fatalf("view is %d cells, want %d: %q", got, width, ansi.Strip(m.View()))
		}
	})
}

func TestSetWidthAndSetThemeChangeALiveField(t *testing.T) {
	m := focused().WithValue("abcdefgh").SetWidth(4)
	require.Equal(t, "fgh ", ansi.Strip(m.View()), "narrowed, the caret stays in view")

	plain := theme.Resolve(theme.Plain, true)
	m = New(theme.Resolved{}).WithPlaceholder("x").SetTheme(plain)
	require.Equal(t, plain.Muted.Render("x"), m.View(), "drawn in the theme learned late")
}

func TestBlinkRestartsAFocusedCaretOnly(t *testing.T) {
	require.Nil(t, New(th).Blink(), "blurred: nothing to blink")
	short := th
	short.CaretBlink = time.Nanosecond
	m, _ := New(short).Focus()
	m, next := m.Update(m.Blink()())
	require.NotNil(t, next, "the tick is the field's own")
	require.NotContains(t, m.View(), short.Caret.Render(" "))
}

// ctrl+v reads the clipboard off the loop and comes back as a paste, which
// the field then takes as one line.
func TestCtrlVPastesTheClipboard(t *testing.T) {
	m := focused().WithClipboard(func() (string, error) { return "from\nthe clipboard", nil })
	m, cmd := m.Update(press('v', tea.ModCtrl))
	require.Empty(t, m.Value(), "nothing until the read answers")
	paste := cmd()
	require.Equal(t, tea.PasteMsg{Content: "from\nthe clipboard"}, paste)
	require.Equal(t, "from the clipboard", run(m, paste).Value())

	_, cmd = focused().Blur().WithClipboard(func() (string, error) { return "x", nil }).Update(press('v', tea.ModCtrl))
	require.Nil(t, cmd, "a blurred field reads nothing")
}

// Home and End on a long value: the window is found from the caret, so the
// cost follows the field's width, not the value's length.
func BenchmarkJumpingAcrossALongValue(b *testing.B) {
	m := focused().WithWidth(40).WithValue(strings.Repeat("a", 40_000))
	home, end := press(tea.KeyHome, 0), press(tea.KeyEnd, 0)
	for b.Loop() {
		m, _ = m.Update(home)
		m, _ = m.Update(end)
	}
}

func TestASuggestionShowsTheRestAndTabTakesIt(t *testing.T) {
	m := focused().WithSuggestions([]string{"commit", "Checkout", "cherry-pick"})
	m = run(m, typed("c"))
	require.Equal(t, []string{"commit", "Checkout", "cherry-pick"}, m.MatchedSuggestions())
	require.Equal(t, "commit", ansi.Strip(m.View()), "the rest muted after the caret")
	require.Contains(t, m.View(), th.Muted.Render("m"))

	m = run(m, typed("h"))
	require.Equal(t, "Checkout", m.CurrentSuggestion(), "matched whatever the case")
	m = run(m, press('n', tea.ModCtrl))
	require.Equal(t, "cherry-pick", m.CurrentSuggestion())
	m = run(m, press('n', tea.ModCtrl))
	require.Equal(t, "Checkout", m.CurrentSuggestion(), "the turn wraps")
	m = run(m, press('p', tea.ModCtrl), press(tea.KeyTab, 0))
	require.Equal(t, "cherry-pick", m.Value())
	require.Empty(t, m.MatchedSuggestions(), "a whole suggestion has nothing to add")
}

func TestWithoutAMatchTabIsLeftToTheHost(t *testing.T) {
	m := focused().WithSuggestions([]string{"commit"}).WithValue("x")
	got, cmd := m.Update(press(tea.KeyTab, 0))
	require.Equal(t, "x", got.Value())
	require.Nil(t, cmd)
	require.Empty(t, focused().WithSuggestions([]string{"a"}).CurrentSuggestion(), "an empty field suggests nothing")
}

func TestASuggestionKeepsToTheWidth(t *testing.T) {
	m := focused().WithWidth(4).WithSuggestions([]string{"abcdefgh"}).WithValue("ab")
	require.Equal(t, "abcd", ansi.Strip(m.View()))
	m = run(m, press(tea.KeyLeft, 0))
	require.Equal(t, "ab  ", ansi.Strip(m.View()), "only with the caret at the end")
}

func TestTheCharLimitStopsTypingAndCutsAPaste(t *testing.T) {
	m := focused().WithCharLimit(4).WithValue("abcdef")
	require.Equal(t, "abcd", m.Value(), "a value is cut too")
	got, cmd := m.Update(typed("x"))
	require.Equal(t, "abcd", got.Value())
	require.Nil(t, cmd, "a key that types nothing is left to the host")

	m = run(m, press(tea.KeyBackspace, 0), press(tea.KeyBackspace, 0), tea.PasteMsg{Content: "中xyz"})
	require.Equal(t, "ab中x", m.Value(), "characters, not bytes")
	require.Equal(t, 4, m.CharLimit())
	require.Equal(t, "abcdef", focused().WithCharLimit(-1).WithValue("abcdef").Value(), "below one is no cap")
}

func TestTheCharLimitCutsATakenSuggestion(t *testing.T) {
	m := focused().WithCharLimit(3).WithSuggestions([]string{"commit"})
	m = run(m, typed("c"), press(tea.KeyTab, 0))
	require.Equal(t, "com", m.Value())
}

func TestTheValidatorJudgesEveryChange(t *testing.T) {
	digits := func(s string) error {
		if strings.Trim(s, "0123456789") != "" {
			return errors.New("digits only")
		}
		return nil
	}
	m := focused().WithValidate(digits).WithValue("12")
	require.NoError(t, m.Err())
	m = run(m, typed("x"))
	require.EqualError(t, m.Err(), "digits only")
	require.Equal(t, "12x", m.Value(), "the text stays as typed")
	m = run(m, press(tea.KeyBackspace, 0))
	require.NoError(t, m.Err())
	require.NoError(t, m.WithValidate(nil).WithValue("x").Err(), "nil checks nothing")
}

func TestAHiddenFieldDrawsNoneOfItsText(t *testing.T) {
	m := focused().WithHidden().WithWidth(6).WithValue("secret words")
	require.Equal(t, "      ", ansi.Strip(m.View()))
	require.Contains(t, m.View(), th.Caret.Render(" "), "the caret on the first cell")
	require.Empty(t, ansi.Strip(m.Blur().WithWidth(0).View()))
	m = run(m, press(tea.KeyBackspace, tea.ModAlt))
	require.Empty(t, m.Value(), "no word boundary to give away")
}

func TestASteadyCaretNeverBlinksAndAHiddenOneNeverShows(t *testing.T) {
	m, cmd := New(th).WithCursorMode(cursor.Steady).Focus()
	require.Nil(t, cmd, "nothing to blink")
	require.Nil(t, m.Blink())
	require.Contains(t, m.View(), th.Caret.Render(" "))
	m, cmd = m.Update(typed("a"))
	require.Nil(t, cmd, "an edit arms no blink either")

	m, cmd = m.SetCursorMode(cursor.Hidden)
	require.Nil(t, cmd)
	require.Equal(t, cursor.Hidden, m.CursorMode())
	require.NotContains(t, m.View(), th.Caret.Render(" "))

	m, cmd = m.SetCursorMode(cursor.Blinking)
	require.NotNil(t, cmd, "back to blinking, the blink starts")
	require.Contains(t, m.View(), th.Caret.Render(" "))
}

func TestTheTerminalCursorStandsInForTheCaret(t *testing.T) {
	m, cmd := New(th).WithTerminalCursor().WithPrompt("> ").WithWidth(4).Focus()
	require.Nil(t, cmd, "the terminal blinks its own")
	m = run(m, typed("a"), typed("中"), typed("b"), typed("c"))
	require.NotContains(t, m.View(), th.Caret.Render(" "), "no painted caret")
	c := m.TerminalCursor()
	require.NotNil(t, c)
	require.Equal(t, 2+2, c.X, "the prompt, then bc: the scroll left 中 out")
	require.True(t, c.Blink)

	m = run(m, press(tea.KeyHome, 0))
	require.Equal(t, 2, m.TerminalCursor().X)
	require.Nil(t, m.Blur().TerminalCursor())
	require.Nil(t, m.WithCursorMode(cursor.Hidden).TerminalCursor())
	require.Nil(t, focused().TerminalCursor(), "a field painting its own caret has none")
}

// The edit survives a theme switch: SetTheme rebuilds the styles under it.
func TestWithStylesHoldsAcrossSetTheme(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m := New(theme.Resolve(theme.Plain, true)).WithPrompt("> ").WithStyles(func(s Styles) Styles { s.Prompt = red; return s })
	m = m.SetTheme(theme.Resolve(theme.Mint, true))
	require.True(t, strings.HasPrefix(m.View(), red.Render("> ")))
}

func TestLegendOffersCompletionOnlyWithASuggestion(t *testing.T) {
	m := New(theme.Resolved{}).WithSuggestions([]string{"deploy"})
	require.Len(t, m.Legend(), 1, "paste alone")
	m = m.WithValue("dep")
	require.Len(t, m.Legend(), 3, "complete, the next suggestion, paste")
}
