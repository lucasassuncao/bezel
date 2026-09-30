package shell

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/overlay"
)

func TestOverlayConsumesKeysBeforeTheApp(t *testing.T) {
	s := newShell()
	s = s.Push(overlay.NewAlert(overlay.Info, "hi", "there", lipgloss.NewStyle(), legend.Style{}))
	s, handled, cmd := s.Update(press("x"))
	require.True(t, handled)
	require.NotNil(t, cmd)
	s, handled, _ = s.Update(cmd()) // CloseMsg
	require.True(t, handled)
	require.False(t, s.HasOverlay())
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
	s := newShell().SetInput(&in)
	s, handled, _ := s.Update(press("a"))
	require.True(t, handled)
	require.Equal(t, "a", s.Input().Value())
	s = s.SetInput(nil)
	_, handled, _ = s.Update(press("a"))
	require.False(t, handled)
}

type answer struct{}

// An async answer landing while a modal is up belongs to the app.
func TestForeignMessagesPassAnOpenOverlay(t *testing.T) {
	s := sized(New(actionConfig()))
	s = s.Push(overlay.NewText("t", []string{"x"}, lipgloss.NewStyle(), legend.Style{}))
	_, handled, _ := s.Update(answer{})
	require.False(t, handled)
}

func TestKeysStillReachAnOpenOverlay(t *testing.T) {
	s := sized(New(actionConfig()))
	s = s.Push(overlay.NewText("t", []string{"x"}, lipgloss.NewStyle(), legend.Style{}))
	_, handled, _ := s.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	require.True(t, handled)
}
