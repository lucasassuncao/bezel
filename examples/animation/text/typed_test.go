package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTypeAndShakeLandAndStop(t *testing.T) {
	m, _ := onTab(t, 0)
	m = press(t, m, "t")
	m = press(t, m, "s")
	ty := m.(model).ty
	require.True(t, ty.typewriter.Active() && ty.shake.Active())
	require.Equal(t, "", ty.typewriter.View(), "the headline starts empty")
	m, out := clockFor(t, m, time.Unix(1000, 0))
	require.Contains(t, out, headline)
	require.Contains(t, out, password)
	require.Equal(t, 0, m.(model).ty.shake.Offset())
}

// The marquee keeps its own ticks: it moves with the shared clock stopped.
func TestTheMarqueeStepsOnItsOwnTicks(t *testing.T) {
	m, _ := onTab(t, 0)
	before := m.(model).ty.marquee.View()
	at := time.Unix(1000, 0)
	m, _ = frame(t, m, m.(model).ty.marquee.FrameAt(at))
	m, _ = frame(t, m, m.(model).ty.marquee.FrameAt(at.Add(time.Second)))
	require.NotEqual(t, before, m.(model).ty.marquee.View())
}
