package animation

import (
	"image/color"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
)

func TestBlendMixesEachChannel(t *testing.T) {
	require.Equal(t, color.RGBA{128, 128, 128, 255}, Blend(black, white, 0.5))
	require.Equal(t, black, Blend(black, white, 0))
	require.Equal(t, white, Blend(black, white, 1))
}

// A nil colour is the terminal's own, which has no channels to mix.
func TestBlendWithANilEndSwitchesHalfWay(t *testing.T) {
	require.Nil(t, Blend(nil, white, 0.2))
	require.Equal(t, white, Blend(nil, white, 0.7))
}

func TestFadeEndsOnTo(t *testing.T) {
	start := time.Unix(1000, 0)
	f := NewFade(black, white, 100*time.Millisecond).WithEase(Linear)
	require.Equal(t, black, f.Value(), "before its first frame it is From")

	f, res := f.Update(f.FrameAt(start))
	require.NotNil(t, res.Cmd)
	f, _ = f.Update(f.FrameAt(start.Add(50 * time.Millisecond)))
	require.Equal(t, color.RGBA{128, 128, 128, 255}, f.Value())
	f, res = f.Update(f.FrameAt(start.Add(time.Second)))
	require.Nil(t, res.Cmd)
	require.Equal(t, white, f.Value())
	require.False(t, f.Active())
}

// A fade passes the options on too: it pulses.
func TestAFadeRepeatsToo(t *testing.T) {
	moving(t)
	start := time.Unix(1000, 0)
	fd := NewFade(black, white, 100*ms).WithEase(Linear).WithYoyo(true).WithRepeat(Forever)
	fd, _ = fd.Update(fd.FrameAt(start))
	fd, _ = fd.Update(fd.FrameAt(start.Add(150 * ms)))
	require.True(t, fd.Active())
}
