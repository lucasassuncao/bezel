package animation

import (
	"image/color"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
)

func TestBlendMixesEachChannel(t *testing.T) {
	assert.Equal(t, color.RGBA{128, 128, 128, 255}, Blend(black, white, 0.5))
	assert.Equal(t, black, Blend(black, white, 0))
	assert.Equal(t, white, Blend(black, white, 1))
}

// A nil colour is the terminal's own, which has no channels to mix.
func TestBlendWithANilEndSwitchesHalfWay(t *testing.T) {
	assert.Nil(t, Blend(nil, white, 0.2))
	assert.Equal(t, white, Blend(nil, white, 0.7))
}

func TestFadeEndsOnTo(t *testing.T) {
	start := time.Unix(1000, 0)
	f := NewFade(black, white, 100*time.Millisecond).WithEase(Linear)
	assert.Equal(t, black, f.Value(), "before its first frame it is From")

	f, res := f.Update(f.FrameAt(start))
	assert.NotNil(t, res.Cmd)
	f, _ = f.Update(f.FrameAt(start.Add(50 * time.Millisecond)))
	assert.Equal(t, color.RGBA{128, 128, 128, 255}, f.Value())
	f, res = f.Update(f.FrameAt(start.Add(time.Second)))
	assert.Nil(t, res.Cmd)
	assert.Equal(t, white, f.Value())
	assert.False(t, f.Active())
}
