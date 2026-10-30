package text

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/animation"
)

func TestATypistEndsWithTheTextRight(t *testing.T) {
	frames := animation.Record(NewTypist("hello, world. done", 20*ms).WithTypos(0.5), Typist.View, 10*time.Second)
	assert.Equal(t, "", frames[0])
	assert.Equal(t, "hello, world. done", frames[len(frames)-1])
}

func TestATypistMakesAndFixesTypos(t *testing.T) {
	const full = "the quick brown fox"
	frames := animation.Record(NewTypist(full, 20*ms).WithTypos(1), Typist.View, 10*time.Second)
	wrong := false
	for _, f := range frames {
		if !strings.HasPrefix(full, f) {
			wrong = true
		}
	}
	assert.True(t, wrong, "some frame shows a key that is not the text's")
	clean := animation.Record(NewTypist(full, 20*ms).WithTypos(0), Typist.View, 10*time.Second)
	for _, f := range clean {
		assert.True(t, strings.HasPrefix(full, f), "%q", f)
	}
}

func TestATypistIsTheSameForTheSameSeed(t *testing.T) {
	a := animation.Record(NewTypist("same text", 20*ms).WithSeed(7), Typist.View, 5*time.Second)
	b := animation.Record(NewTypist("same text", 20*ms).WithSeed(7), Typist.View, 5*time.Second)
	assert.Equal(t, a, b)
}

func TestATypistPausesAfterAFullStop(t *testing.T) {
	ty := NewTypist("a. b", 20*ms).WithTypos(0)
	var dot, b time.Duration
	for _, s := range ty.script {
		switch strings.Join(s.shown, "") {
		case "a. ":
			dot = s.at
		case "a. b":
			b = s.at
		}
	}
	assert.Greater(t, b-dot, 2*20*ms, "the letter after a sentence waits longer")
}

func TestStillTypistShowsTheTextWhole(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	ty := NewTypist("now", 20*ms)
	assert.False(t, ty.Active())
	assert.Equal(t, "now", ty.View())
}

func TestATypistTakesOnlyItsOwnFrames(t *testing.T) {
	ty := NewTypist("abc", 20*ms).WithDelay(ms)
	_, res := ty.Update(NewTypist("x", ms).FrameAt(time.Unix(1, 0)))
	assert.Nil(t, res.Cmd)
	_, res = ty.Update(ty.FrameAt(time.Unix(1, 0)))
	assert.NotNil(t, res.Cmd)
}

func TestATypistWithNoTimeShowsTheTextAtOnce(t *testing.T) {
	ty := NewTypist("now", 0)
	assert.False(t, ty.Active())
	assert.Equal(t, "now", ty.View())
}
