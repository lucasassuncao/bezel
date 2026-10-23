package animation

import (
	"image/color"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// moving clears both switches, so a test reads motion the way a normal terminal has it.
func moving(t *testing.T) {
	t.Helper()
	t.Setenv("BEZEL_NO_MOTION", "")
	t.Setenv("TERM", "xterm-256color")
}

func TestStillIsReadFromTheEnvironment(t *testing.T) {
	moving(t)
	assert.False(t, Still())
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.True(t, Still())
	moving(t)
	t.Setenv("TERM", "dumb")
	assert.True(t, Still(), "a dumb terminal redraws too poorly to animate")
}

// Still, every animation is born where it would have ended, asking for no frame.
func TestStillAnimationsAreBornAtTheirEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")

	tw := New(0, 10, time.Second)
	assert.False(t, tw.Active())
	assert.Equal(t, 10, tw.Cur)
	assert.Nil(t, tw.Tick())

	f := NewFloat(0, 1, time.Second)
	assert.False(t, f.Active())
	assert.Equal(t, 1.0, f.Cur)

	assert.Equal(t, white, NewFade(black, white, time.Second).Color())
	assert.False(t, NewShake(3, time.Second).Active())
	assert.Equal(t, "hello", NewTypewriter("hello", time.Second).View())

	s, cmd := NewSpring(0).To(10)
	assert.Nil(t, cmd)
	assert.Equal(t, 10.0, s.Pos, "a still spring jumps")

	m, cmd := NewMarquee("abcdefghij", 4).Start()
	assert.Nil(t, cmd)
	assert.False(t, m.Active())
	assert.Equal(t, "abcd", m.View())
}

// TestMain clears both switches, so a CI with TERM=dumb still tests motion.
func TestMain(m *testing.M) {
	_ = os.Setenv("BEZEL_NO_MOTION", "")
	_ = os.Setenv("TERM", "xterm-256color")
	os.Exit(m.Run())
}

func TestStillEffectsAreBornAtTheirEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.False(t, NewShimmer("abc", black, white, time.Second).Active())
	assert.Equal(t, "ab-cd", NewScramble("ab-cd", time.Second).View())
	assert.Equal(t, "123", NewOdometer(0, 123, time.Second).View())
	g := NewGradient("abcd", []color.Color{black, white}, time.Second)
	assert.False(t, g.Active())
	assert.Equal(t, white, g.colorAt(2), "a still gradient is drawn, only not moved")
}
