package shell

import (
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

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
	in := textinput.New()
	in.Focus()
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
func TestStatusInputIsAValue(t *testing.T) {
	in := textinput.New()
	in.Focus()
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
