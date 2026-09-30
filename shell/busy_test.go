package shell

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/spinner"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func view(s Shell) string {
	return ansi.Strip(s.View(map[string]Pane{"list": {}, "detail": {}}))
}

func TestBusyShowsTheSpinnerAndTextOnTheStatusRow(t *testing.T) {
	s := sized(New(actionConfig(stubTab{name: "a"})))
	require.Contains(t, view(s), "status of a")

	s, cmd := s.Busy("refreshing…")
	require.NotNil(t, cmd, "going busy starts the spinner")
	require.True(t, s.IsBusy())
	v := view(s)
	require.Contains(t, v, s.Spinner()+" refreshing…")
	require.NotContains(t, v, "status of a", "busy outranks the tab's own status")

	s, cmd = s.Busy("installing…")
	require.Nil(t, cmd, "already spinning: a second loop would double the speed")
	require.Contains(t, view(s), "installing…")

	s = s.Idle()
	require.False(t, s.IsBusy())
	require.Contains(t, view(s), "status of a")
}

// A banner set while busy must not be lost to the spinner, and the spinner
// comes back once the banner expires.
func TestABannerShowsOverBusyAndBusyReturns(t *testing.T) {
	s := sized(New(actionConfig()))
	s, _ = s.Busy("refreshing…")
	s, _ = s.SetStatus("could not reach scoop", Error, time.Second)
	require.Contains(t, view(s), "could not reach scoop")
	require.NotContains(t, view(s), "refreshing…")

	s, _, _ = s.Update(s.StatusExpiry())
	require.Contains(t, view(s), "refreshing…")
}

// The tick loop runs only while busy: after Idle the next tick ends it.
func TestTheSpinnerTicksOnlyWhileBusy(t *testing.T) {
	s := sized(New(actionConfig()))
	s, cmd := s.Busy("refreshing…")
	tick := cmd()
	require.IsType(t, spinner.TickMsg{}, tick)

	before := s.Spinner()
	s, handled, next := s.Update(tick)
	require.True(t, handled)
	require.NotNil(t, next, "still busy: the loop re-arms")
	require.NotEqual(t, before, s.Spinner(), "a tick advances the frame")

	s = s.Idle()
	_, handled, next = s.Update(next())
	require.True(t, handled, "the shell's own tick is never the app's")
	require.Nil(t, next, "idle: the loop stops")
}

func TestAnotherSpinnersTickIsTheApps(t *testing.T) {
	s := sized(New(actionConfig()))
	s, _ = s.Busy("refreshing…")
	_, handled, _ := s.Update(spinner.New().Tick())
	require.False(t, handled)
}

func TestIdleLeavesABannerAlone(t *testing.T) {
	s := sized(New(actionConfig()))
	s, _ = s.SetStatus("saved", OK, 0)
	require.Contains(t, view(s.Idle()), "saved")
}
