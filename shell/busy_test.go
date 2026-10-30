package shell

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/component/spinner"
)

// busyShell is a sized shell whose spinner turns every nanosecond, so a
// test that runs a tick never waits on one.
func busyShell(tabs ...Tab) Shell {
	c := actionConfig(tabs...)
	c.Spinner = spinner.Spinner{Frames: spinner.Dot().Frames, Every: time.Nanosecond}
	return sized(New(c))
}

func view(s Shell) string {
	return ansi.Strip(s.View(map[string]Pane{"list": {}, "detail": {}}))
}

func TestBusyShowsTheSpinnerAndTextOnTheStatusRow(t *testing.T) {
	s := busyShell(stubTab{name: "a"})
	require.Contains(t, view(s), "status of a")

	s, cmd := s.Busy("refreshing…")
	require.NotNil(t, cmd, "going busy starts the spinner")
	require.True(t, s.IsBusy())
	v := view(s)
	require.Contains(t, v, s.SpinnerFrame()+" refreshing…")
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
	s := busyShell()
	s, _ = s.Busy("refreshing…")
	s, _ = s.SetStatus("could not reach scoop", Error, time.Second)
	require.Contains(t, view(s), "could not reach scoop")
	require.NotContains(t, view(s), "refreshing…")

	s, _ = s.Update(s.StatusExpiry())
	require.Contains(t, view(s), "refreshing…")
}

// The tick loop runs only while busy: after Idle the next tick ends it.
func TestTheSpinnerTicksOnlyWhileBusy(t *testing.T) {
	s := busyShell()
	s, cmd := s.Busy("refreshing…")
	tick := cmd()
	require.IsType(t, animation.FrameMsg{}, tick)

	before := s.SpinnerFrame()
	s, r := s.Update(tick)
	require.True(t, r.Handled)
	require.NotNil(t, r.Cmd, "still busy: the loop re-arms")
	require.NotEqual(t, before, s.SpinnerFrame(), "a tick advances the frame")

	next := r.Cmd
	s = s.Idle()
	_, r = s.Update(next())
	require.True(t, r.Handled, "the shell's own tick is never the app's")
	require.Nil(t, r.Cmd, "idle: the loop stops")
}

func TestAnotherShellsTickIsTheApps(t *testing.T) {
	s := busyShell()
	s, _ = s.Busy("refreshing…")
	_, cmd := busyShell().Busy("elsewhere")
	_, r := s.Update(cmd())
	require.False(t, r.Handled)
}

// Idle then Busy again before the old loop's tick lands: that tick ends,
// or the spinner would turn twice as fast.
func TestAStaleLoopsTickDoesNotDoubleTheSpinner(t *testing.T) {
	s := busyShell()
	s, first := s.Busy("refreshing…")
	s, second := s.Idle().Busy("again")
	before := s.SpinnerFrame()

	s, r := s.Update(first())
	require.True(t, r.Handled)
	require.Nil(t, r.Cmd, "the old loop stops")
	require.Equal(t, before, s.SpinnerFrame())

	_, r = s.Update(second())
	require.NotNil(t, r.Cmd, "the new loop runs on")
}

func TestIdleLeavesABannerAlone(t *testing.T) {
	s := busyShell()
	s, _ = s.SetStatus("saved", OK, 0)
	require.Contains(t, view(s.Idle()), "saved")
}

func TestAZeroSpinnerIsTheDot(t *testing.T) {
	s, _ := sized(New(actionConfig())).Busy("x")
	require.Equal(t, "⣾ ", s.SpinnerFrame())
}

func TestConfigSpinnerPicksTheBusyMark(t *testing.T) {
	c := actionConfig()
	c.Spinner = spinner.Line()
	s, _ := sized(New(c)).Busy("refreshing…")
	require.Contains(t, view(s), "| refreshing…")
}

func TestAnotherSpinnersTickIsTheApps(t *testing.T) {
	_, cmd := spinner.New(spinner.Spinner{Frames: []string{"a"}, Every: time.Nanosecond}).Start()
	s, _ := busyShell(stubTab{name: "a"}).Busy("x")
	_, r := s.Update(cmd())
	require.False(t, r.Handled)
}

func TestBusyWithNoTextGoesIdleAndEndsTheLoop(t *testing.T) {
	s, cmd := busyShell().Busy("refreshing…")
	s, _ = s.Busy("")
	require.False(t, s.IsBusy())
	_, r := s.Update(cmd())
	require.True(t, r.Handled)
	require.Nil(t, r.Cmd, "idle: the loop stops")
}

func TestAConfigSpinnerWithOnlyAPaceKeepsIt(t *testing.T) {
	c := actionConfig()
	c.Spinner = spinner.Spinner{Every: time.Nanosecond}
	sp := New(c).spin.Spinner()
	require.Equal(t, time.Nanosecond, sp.Every)
	require.Equal(t, spinner.Dot().Frames, sp.Frames)
}
