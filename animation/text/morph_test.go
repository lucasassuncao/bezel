package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
)

func TestMorphKeepsWhatTheTwoTextsShare(t *testing.T) {
	m := NewMorph("saving", "saved", 100*ms)
	frames := animation.Record(m, Morph.View, time.Second)
	require.Equal(t, "saving", frames[0])
	require.Equal(t, "saved", frames[len(frames)-1])
	for _, f := range frames {
		require.Contains(t, f, "sav", "the shared start never leaves")
	}
}

func TestMorphTakesOutBeforeItPutsIn(t *testing.T) {
	m := NewMorph("abc", "axy", 100*ms)
	start := time.Unix(10, 0)
	m, _ = m.Advance(start)
	mid, _ := m.Advance(start.Add(50 * ms))
	require.Equal(t, "a", mid.View(), "halfway the old letters are gone and the new not yet in")
}

func TestMorphOfEqualTextsDoesNothing(t *testing.T) {
	m := NewMorph("same", "same", 100*ms)
	require.False(t, m.Active())
	require.Equal(t, "same", m.View())
}

func TestMorphToCarriesOnFromWhereItWasHeading(t *testing.T) {
	m, cmd := NewMorph("one", "two", 100*ms).To("three")
	require.NotNil(t, cmd)
	require.Equal(t, "two", m.View())
	end := animation.Record(m.WithDelay(0), Morph.View, time.Second)
	require.Equal(t, "three", end[len(end)-1])
}

func TestStillMorphIsBornAtItsEnd(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	require.Equal(t, "done", NewMorph("doing", "done", time.Second).View())
}

func TestMorphTakesOnlyItsOwnFrames(t *testing.T) {
	m := NewMorph("ab", "cd", 100*ms)
	_, res := m.Update(NewMorph("x", "y", ms).FrameAt(time.Unix(1, 0)))
	require.Nil(t, res.Cmd)
	_, res = m.Update(m.FrameAt(time.Unix(1, 0)))
	require.NotNil(t, res.Cmd)
}
