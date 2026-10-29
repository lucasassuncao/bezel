package palette

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/input"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

type ran struct{ name, arg string }

func echo(name string) func(string) tea.Cmd {
	return func(arg string) tea.Cmd { return func() tea.Msg { return ran{name, arg} } }
}

func items() []Item {
	return []Item{
		{Name: "goto", Arg: "path", Desc: "jump to a path", Available: true, Run: echo("goto")},
		{Name: "copy", Key: "y", Desc: "copy the secret", Available: true, Run: echo("copy")},
		{Name: "copy-path", Desc: "copy the path", Available: true, Run: echo("copy-path")},
		{Name: "create", Key: "a", Desc: "create a secret", Available: true, Run: echo("create")},
		{Name: "delete", Key: "d", Desc: "delete the secret", Available: false, Run: echo("delete")},
	}
}

func open() Model { return New(items(), theme.Resolve(theme.Plain, true)) }

func typed(m Model, s string) Model {
	for _, r := range s {
		o, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = o.(Model)
	}
	return m
}

func press(m Model, code rune) (Model, tea.Cmd) {
	o, cmd := m.Update(tea.KeyPressMsg{Code: code})
	return o.(Model), cmd
}

func done(t *testing.T, cmd tea.Cmd) DoneMsg {
	t.Helper()
	require.NotNil(t, cmd)
	d, ok := cmd().(DoneMsg)
	require.True(t, ok, "enter finishes with a DoneMsg")
	return d
}

func TestExactNameBeatsALongerPrefix(t *testing.T) {
	_, cmd := press(typed(open(), "copy"), tea.KeyEnter)
	require.Equal(t, ran{"copy", ""}, done(t, cmd).Cmd())
}

func TestAUniquePrefixRuns(t *testing.T) {
	_, cmd := press(typed(open(), "cr"), tea.KeyEnter)
	require.Equal(t, ran{"create", ""}, done(t, cmd).Cmd())
}

func TestAnAmbiguousLineStaysOpenAndSaysSo(t *testing.T) {
	m, cmd := press(typed(open(), "c"), tea.KeyEnter)
	require.Nil(t, cmd)
	require.Contains(t, view(m), "c is ambiguous: copy, copy-path, create")
}

func TestUnavailableIsNotUnknown(t *testing.T) {
	_, cmd := press(typed(open(), "delete"), tea.KeyEnter)
	require.Equal(t, "delete is not available here", done(t, cmd).Err)
	_, cmd = press(typed(open(), "nope"), tea.KeyEnter)
	require.Equal(t, "unknown command: nope", done(t, cmd).Err)
}

func TestAnEmptyLineJustCloses(t *testing.T) {
	_, cmd := press(open(), tea.KeyEnter)
	d := done(t, cmd)
	require.Empty(t, d.Err)
	require.Nil(t, d.Cmd)
}

func TestACommandThatTakesNoArgumentRefusesOne(t *testing.T) {
	_, cmd := press(typed(open(), "copy x"), tea.KeyEnter)
	require.Equal(t, ":copy takes no argument", done(t, cmd).Err)
}

func TestTheArgumentReachesTheCommand(t *testing.T) {
	_, cmd := press(typed(open(), "goto kv/app"), tea.KeyEnter)
	require.Equal(t, ran{"goto", "kv/app"}, done(t, cmd).Cmd())
}

func TestArrowsPickAndEnterRunsThePick(t *testing.T) {
	m, _ := press(open(), tea.KeyDown) // copy
	m, _ = press(m, tea.KeyDown)       // copy-path
	_, cmd := press(m, tea.KeyEnter)
	require.Equal(t, ran{"copy-path", ""}, done(t, cmd).Cmd())
}

func TestUpFromNothingPicksTheLast(t *testing.T) {
	m, _ := press(open(), tea.KeyUp)
	_, cmd := press(m, tea.KeyEnter)
	require.Equal(t, ran{"goto", ""}, done(t, cmd).Cmd())
}

func TestTypingClearsThePick(t *testing.T) {
	m, _ := press(open(), tea.KeyDown)
	m = typed(m, "c")
	require.Equal(t, -1, m.cursor)
}

func TestTabCompletesTheOnlyCandidate(t *testing.T) {
	m, _ := press(typed(open(), "cr"), tea.KeyTab)
	require.Equal(t, "create", m.input.Value())
}

func TestTabLeavesRoomForAnArgument(t *testing.T) {
	m, _ := press(typed(open(), "go"), tea.KeyTab)
	require.Equal(t, "goto ", m.input.Value())
	m, _ = press(typed(open(), "go kv"), tea.KeyTab)
	require.Equal(t, "goto kv", m.input.Value())
}

// An exact name that is refused must not fall through to a prefix sibling.
func TestARefusedExactNameDoesNotRunItsPrefixSibling(t *testing.T) {
	its := []Item{
		{Name: "copy", Available: false, Run: echo("copy")},
		{Name: "copy-path", Available: true, Run: echo("copy-path")},
	}
	m := typed(New(its, theme.Resolve(theme.Plain, true)), "copy")
	m, cmd := press(m, tea.KeyEnter)
	// Refused, but its siblings are one keystroke away: stay open, run nothing.
	require.Nil(t, cmd)
	require.Contains(t, view(m), "copy is not available here")
	require.Contains(t, view(m), ":copy-path")
}

// A path pasted after ":goto " lands in the line like typed text.
func TestPasteLandsInTheLine(t *testing.T) {
	m := typed(open(), "goto ")
	o, _ := m.Update(tea.PasteMsg{Content: "kv/app/prod"})
	require.Equal(t, "goto kv/app/prod", o.(Model).input.Value())
}

func TestValueIsTheTypedLine(t *testing.T) {
	require.Equal(t, "goto kv", typed(open(), "goto kv").Value())
}

func TestSetItemsKeepsTheLineAndClampsThePick(t *testing.T) {
	m, _ := press(typed(open(), "copy"), tea.KeyDown)
	m = m.SetItems([]Item{{Name: "copy", Available: true}})
	require.Equal(t, "copy", m.Value())
	require.Contains(t, view(m), ":copy")
	require.NotPanics(t, func() { m.View(layout.Rect{W: 80, H: 20}) })
}

// The caret's blink comes back bare, at the theme's pace; the shell routes it.
func TestTheCaretBlinksAtTheThemesPace(t *testing.T) {
	_, cmd := New(items(), theme.Resolved{CaretBlink: time.Nanosecond}).Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	require.IsType(t, input.BlinkMsg{}, cmd())
}

func TestTheLegendNamesThePalettesKeys(t *testing.T) {
	line := ansi.Strip(legend.HintLine(open().Legend(), legend.Style{}))
	require.Equal(t, "[enter] run · [tab] complete · [↑/↓] choose · [esc] cancel", line)
}

func TestThePaletteAnswersToTheKeysItWasGiven(t *testing.T) {
	keys := DefaultKeys()
	keys.Up = key.New("ctrl+p").WithHelp("ctrl+p/n", "choose")
	keys.Down = key.New("ctrl+n")
	m := open().WithKeys(keys)

	m, _ = press(m, tea.KeyDown)
	require.Equal(t, -1, m.cursor, "the arrow is no longer bound")
	o, _ := m.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	require.Equal(t, 0, o.(Model).cursor)
	require.Contains(t, ansi.Strip(legend.HintLine(m.Legend(), legend.Style{})), "[ctrl+p/n] choose")
}

// Whatever is typed, the palette resolves to an available item the name
// leads to, or to an error: never to a refused item, never to a panic.
func FuzzResolveOnlyRunsWhatTheNameLeadsTo(f *testing.F) {
	for _, s := range []string{"", ":", "goto /a b", "  :c  ", "copy", "delete", "co", "::x", "\t:goto\t"} {
		f.Add(s)
	}
	m := open()
	f.Fuzz(func(t *testing.T, line string) {
		name, arg := splitLine(line)
		if strings.Contains(name, " ") || arg != strings.TrimSpace(arg) {
			t.Fatalf("splitLine(%q) = %q, %q", line, name, arg)
		}
		it, gotArg, err := m.resolve(line)
		if err != nil {
			return
		}
		if !it.Available || !strings.HasPrefix(it.Name, name) || gotArg != arg {
			t.Fatalf("resolve(%q) = %q (available %v), arg %q", line, it.Name, it.Available, gotArg)
		}
	})
}
