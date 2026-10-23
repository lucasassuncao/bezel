package animation

import (
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
