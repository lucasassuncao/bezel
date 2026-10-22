package shell

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/input"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/theme"
)

func TestOverlayConsumesKeysBeforeTheApp(t *testing.T) {
	s := newShell()
	s = s.Push(overlay.NewAlert(overlay.Info, "hi", "there", theme.Resolved{}))
	s, handled, cmd := s.Update(press("x"))
	require.True(t, handled)
	require.NotNil(t, cmd)
	s, handled, _ = s.Update(cmd()) // CloseMsg
	require.True(t, handled)
	require.False(t, s.HasOverlay())
}

// An overlay knows the body it floats in from the moment it is pushed, and
// again after a resize, before anything has drawn it.
func TestOverlaysAreSizedToTheBody(t *testing.T) {
	pgdown := tea.KeyPressMsg{Code: tea.KeyPgDown}
	pager := overlay.NewPager("Out", strings.Repeat("line\n", 200), theme.Resolved{})
	// page is how far one pgdown goes in a pager sized by hand.
	page := func(body layout.Rect) int {
		o, _ := pager.Resize(body).Update(pgdown)
		return o.(overlay.Pager).Offset()
	}

	s := sized(newShell()).Push(pager)
	s, _, _ = s.Update(pgdown)
	first := page(s.Body())
	require.Equal(t, first, s.TopOverlay().(overlay.Pager).Offset())

	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 120, Height: 16})
	s, _, _ = s.Update(pgdown)
	require.Equal(t, first+page(s.Body()), s.TopOverlay().(overlay.Pager).Offset())
}

// Without actions the shell handles no key: not even tab or "?".
func TestUnknownKeysAreLeftToTheApp(t *testing.T) {
	s := newShell(stubTab{name: "a"}, stubTab{name: "b"})
	for _, k := range []tea.KeyPressMsg{press("x"), press("?"), {Code: tea.KeyTab}} {
		_, handled, _ := s.Update(k)
		require.False(t, handled, "%s", k)
	}
}

func TestStatusExpiresOnlyForItsOwnTick(t *testing.T) {
	s := newShell()
	s, cmd1 := s.SetStatus("first", OK, time.Second)
	s, cmd2 := s.SetStatus("second", Error, time.Second)
	require.NotNil(t, cmd1)
	require.NotNil(t, cmd2)

	s, handled, _ := s.Update(statusExpiredMsg{seq: 1}) // stale
	require.True(t, handled)
	require.Equal(t, "second", s.status)

	s, _, _ = s.Update(statusExpiredMsg{seq: 2})
	require.Equal(t, "", s.status)
}

func TestStatusInputTakesTheKeys(t *testing.T) {
	in, _ := input.New(theme.Resolved{}).Focus()
	s := newShell().SetInput(in)
	s, handled, _ := s.Update(press("a"))
	require.True(t, handled)
	got, ok := s.Input()
	require.True(t, ok)
	require.Equal(t, "a", got.Value())
	s = s.ClearInput()
	_, ok = s.Input()
	require.False(t, ok)
	_, handled, _ = s.Update(press("a"))
	require.False(t, handled)
}

// The input is the shell's own copy: an edit shows only in the shell Update returned.
// A paste and the caret's blink are the input's too while it is up; after
// ClearInput they are the app's again.
func TestStatusInputTakesAPasteAndItsBlink(t *testing.T) {
	in, _ := input.New(theme.Resolved{CaretBlink: time.Nanosecond}).Focus()
	s, handled, blink := newShell().SetInput(in).Update(tea.PasteMsg{Content: "pasted"})
	require.True(t, handled)
	got, _ := s.Input()
	require.Equal(t, "pasted", got.Value())

	_, handled, next := s.Update(blink())
	require.True(t, handled)
	require.NotNil(t, next, "the blink runs on")

	_, handled, _ = s.ClearInput().Update(tea.PasteMsg{Content: "x"})
	require.False(t, handled)
}

// A dialog's caret blinks through the shell; a blink no field of the shell's
// armed belongs to an input the app drives itself.
func TestABlinkReachesTheFieldThatArmedIt(t *testing.T) {
	fast := theme.Resolved{CaretBlink: time.Nanosecond}
	s := newShell().Push(overlay.NewPrompt("name", nil, fast))
	s, _, cmd := s.Update(press("a"))
	_, handled, next := s.Update(cmd())
	require.True(t, handled)
	require.NotNil(t, next, "the prompt's blink re-arms")

	_, own := input.New(fast).Focus()
	_, handled, _ = s.Update(own())
	require.False(t, handled, "the app's own field")
}

func TestStatusInputIsAValue(t *testing.T) {
	in, _ := input.New(theme.Resolved{}).Focus()
	before := newShell().SetInput(in)
	after, _, _ := before.Update(press("a"))

	got, _ := after.Input()
	require.Equal(t, "a", got.Value())
	old, _ := before.Input()
	require.Empty(t, old.Value())
	require.Empty(t, in.Value())
}

type answer struct{}

// An async answer landing while a modal is up belongs to the app.
func TestForeignMessagesPassAnOpenOverlay(t *testing.T) {
	s := sized(New(actionConfig()))
	s = s.Push(overlay.NewText("t", []string{"x"}, theme.Resolved{}))
	_, handled, _ := s.Update(answer{})
	require.False(t, handled)
}

func TestKeysStillReachAnOpenOverlay(t *testing.T) {
	s := sized(New(actionConfig()))
	s = s.Push(overlay.NewText("t", []string{"x"}, theme.Resolved{}))
	_, handled, _ := s.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	require.True(t, handled)
}

// What an overlay's own command answers is the overlay's, even though other
// non-key messages pass it by for the app.
func TestTopMsgReachesAnOpenOverlay(t *testing.T) {
	s := sized(New(actionConfig()))
	s = s.Push(overlay.NewText("t", []string{"x"}, theme.Resolved{}))
	_, handled, _ := s.Update(overlay.TopMsg{Msg: answer{}})
	require.True(t, handled)
}

// A key the shell hands back to the app, and the frame after it: what every
// keystroke in a running app costs the shell.
func BenchmarkKeyAndFrame(b *testing.B) {
	s := newShell()
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	panes, key := demoPanes(), press("z")
	for b.Loop() {
		s, _, _ = s.Update(key)
		_ = s.View(panes)
	}
}
