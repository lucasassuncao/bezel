package input

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

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
