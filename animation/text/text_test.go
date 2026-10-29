package text

import (
	"image/color"
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestTypewriterRevealsAGraphemeAtATime(t *testing.T) {
	start := time.Unix(1000, 0)
	w := NewTypewriter("héllo", 10*time.Millisecond)
	assert.Equal(t, "", w.View())

	w, cmd := w.Update(w.FrameAt(start))
	assert.NotNil(t, cmd)
	w, _ = w.Update(w.FrameAt(start.Add(25 * time.Millisecond)))
	assert.Equal(t, "hé", w.View())
	w, cmd = w.Update(w.FrameAt(start.Add(time.Second)))
	assert.Nil(t, cmd)
	assert.Equal(t, "héllo", w.View())
	assert.False(t, w.Active())

	assert.Equal(t, "héllo", NewTypewriter("héllo", time.Second).Skip().View())
	assert.False(t, NewTypewriter("", time.Second).Active())
}

func TestMarqueeScrollsTextWiderThanItsWidth(t *testing.T) {
	start := time.Unix(1000, 0)
	m := NewMarquee("abcdefghij", 4).WithEvery(100 * time.Millisecond)
	assert.Equal(t, "abcd", m.View())

	m, cmd := m.Start()
	assert.NotNil(t, cmd)
	m, _ = m.Advance(start)
	m, _ = m.Advance(start.Add(250 * time.Millisecond))
	assert.Equal(t, "cdef", m.View())
	m, _ = m.Advance(start.Add(900 * time.Millisecond))
	assert.Equal(t, "j   ", m.View(), "the gap comes round before the text again")
	for ms := 0; ms < 3000; ms += 70 {
		m, _ = m.Advance(start.Add(time.Duration(ms) * time.Millisecond))
		assert.Equal(t, 4, ansi.StringWidth(m.View()))
	}
	assert.Equal(t, "abcd", m.Stop().View(), "stopping goes back to the start")
}

func TestMarqueeLeavesTextThatFitsAlone(t *testing.T) {
	m, cmd := NewMarquee("ab", 4).Start()
	assert.Nil(t, cmd)
	assert.False(t, m.Active())
	assert.Equal(t, "ab  ", m.View())
}

func TestMarqueeTakesItsOwnTicks(t *testing.T) {
	m, _ := NewMarquee("abcdefghij", 4).WithEvery(100 * time.Millisecond).Start()
	start := time.Unix(1000, 0)
	m, cmd := m.Update(m.FrameAt(start))
	assert.NotNil(t, cmd)
	m, _ = m.Update(m.FrameAt(start.Add(100 * time.Millisecond)))
	assert.Equal(t, "bcde", m.View())
	_, cmd = m.Stop().Update(m.FrameAt(start))
	assert.Nil(t, cmd, "a stopped marquee drops its ticks")
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

func TestStillTextIsBornWhole(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.False(t, NewShake(3, time.Second).Active())
	assert.Equal(t, "hello", NewTypewriter("hello", time.Second).View())
	m, cmd := NewMarquee("abcdefghij", 4).Start()
	assert.Nil(t, cmd)
	assert.False(t, m.Active())
	assert.Equal(t, "abcd", m.View())
}

// A typewriter that yoyos types and then erases itself.
func TestTextRepeatsToo(t *testing.T) {
	start := time.Unix(1000, 0)
	w := NewTypewriter("abcd", 25*ms).WithYoyo(true).WithRepeat(1).WithDelay(0)
	w, _ = w.Update(w.FrameAt(start))
	w, _ = w.Update(w.FrameAt(start.Add(150 * ms)))
	assert.Equal(t, "ab", w.View(), "erasing")
	w, _ = w.Update(w.FrameAt(start.Add(time.Second)))
	assert.Equal(t, "", w.View())

	s := NewShake(2, 100*ms).WithDelay(50 * ms).WithRepeat(1)
	s, _ = s.Update(s.FrameAt(start))
	s, _ = s.Update(s.FrameAt(start.Add(150 * ms)))
	assert.True(t, s.Active())
}

const ms = time.Millisecond

var (
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
)

// TestMain clears the motion switches, so a CI with TERM=dumb still tests motion.
func TestMain(m *testing.M) {
	_ = os.Setenv("BEZEL_NO_MOTION", "")
	_ = os.Setenv("BEZEL_MOTION", "")
	_ = os.Setenv("TERM", "xterm-256color")
	os.Exit(m.Run())
}
