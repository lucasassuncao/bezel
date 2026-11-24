package text

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
)

func TestATypistEndsWithTheTextRight(t *testing.T) {
	frames := animation.Record(NewTypist("hello, world. done", 20*ms).WithTypos(0.5), Typist.View, 10*time.Second)
	require.Equal(t, "", frames[0])
	require.Equal(t, "hello, world. done", frames[len(frames)-1])
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
	require.True(t, wrong, "some frame shows a key that is not the text's")
	clean := animation.Record(NewTypist(full, 20*ms).WithTypos(0), Typist.View, 10*time.Second)
	for _, f := range clean {
		require.True(t, strings.HasPrefix(full, f), "%q", f)
	}
}

func TestATypistIsTheSameForTheSameSeed(t *testing.T) {
	a := animation.Record(NewTypist("same text", 20*ms).WithSeed(7), Typist.View, 5*time.Second)
	b := animation.Record(NewTypist("same text", 20*ms).WithSeed(7), Typist.View, 5*time.Second)
	require.Equal(t, a, b)
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
	require.Greater(t, b-dot, 2*20*ms, "the letter after a sentence waits longer")
}

func TestStillTypistShowsTheTextWhole(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	ty := NewTypist("now", 20*ms)
	require.False(t, ty.Active())
	require.Equal(t, "now", ty.View())
}

func TestATypistTakesOnlyItsOwnFrames(t *testing.T) {
	ty := NewTypist("abc", 20*ms).WithDelay(ms)
	_, res := ty.Update(NewTypist("x", ms).FrameAt(time.Unix(1, 0)))
	require.Nil(t, res.Cmd)
	_, res = ty.Update(ty.FrameAt(time.Unix(1, 0)))
	require.NotNil(t, res.Cmd)
}

func TestATypistWithNoTimeShowsTheTextAtOnce(t *testing.T) {
	ty := NewTypist("now", 0)
	require.False(t, ty.Active())
	require.Equal(t, "now", ty.View())
}
