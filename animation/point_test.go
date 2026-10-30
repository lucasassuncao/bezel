package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
)

func TestMoveEasesAPointAlongBothAxes(t *testing.T) {
	start := time.Unix(1000, 0)
	m := NewMove(Point{0, 0}, Point{10, 4}, 100*ms).WithEase(Linear)
	m, _ = m.Update(m.FrameAt(start))
	m, _ = m.Update(m.FrameAt(start.Add(50 * ms)))
	assert.Equal(t, Point{5, 2}, m.Cur)
	m, res := m.Update(m.FrameAt(start.Add(time.Second)))
	assert.Nil(t, res.Cmd)
	assert.Equal(t, Point{10, 4}, m.Cur)
	x, y := m.Cur.Round()
	assert.Equal(t, [2]int{10, 4}, [2]int{x, y})
}

func TestFollowSpringsAPointOnOneClock(t *testing.T) {
	at := time.Unix(1000, 0)
	f, cmd := NewFollow(Point{0, 0}).To(Point{10, 5})
	require.NotNil(t, cmd)
	res := bezel.Result{Cmd: cmd}
	for i := 0; res.Cmd != nil; i++ {
		require.Less(t, i, 1000, "it never came to rest")
		f, res = f.Update(f.FrameAt(at))
		at = at.Add(Frame)
	}
	assert.Equal(t, Point{10, 5}, f.Pos())
	assert.False(t, f.Active())

	_, cmd = f.To(Point{10, 5})
	assert.Nil(t, cmd, "already there")
}

func TestStillPointsJump(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	assert.Equal(t, Point{3, 4}, NewMove(Point{}, Point{3, 4}, time.Second).Cur)
	f, cmd := NewFollow(Point{}).To(Point{3, 4})
	assert.Nil(t, cmd)
	assert.Equal(t, Point{3, 4}, f.Pos())
}
