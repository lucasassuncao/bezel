package animation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
)

func TestMoveEasesAPointAlongBothAxes(t *testing.T) {
	start := time.Unix(1000, 0)
	m := NewMove(Point{0, 0}, Point{10, 4}, 100*ms).WithEase(Linear)
	m, _ = m.Update(m.FrameAt(start))
	m, _ = m.Update(m.FrameAt(start.Add(50 * ms)))
	require.Equal(t, Point{5, 2}, m.cur)
	m, res := m.Update(m.FrameAt(start.Add(time.Second)))
	require.Nil(t, res.Cmd)
	require.Equal(t, Point{10, 4}, m.cur)
	x, y := m.cur.Round()
	require.Equal(t, [2]int{10, 4}, [2]int{x, y})
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
	require.Equal(t, Point{10, 5}, f.Value())
	require.False(t, f.Active())

	_, cmd = f.To(Point{10, 5})
	require.Nil(t, cmd, "already there")
}

func TestStillPointsJump(t *testing.T) {
	t.Setenv("BEZEL_NO_MOTION", "1")
	require.Equal(t, Point{3, 4}, NewMove(Point{}, Point{3, 4}, time.Second).cur)
	f, cmd := NewFollow(Point{}).To(Point{3, 4})
	require.Nil(t, cmd)
	require.Equal(t, Point{3, 4}, f.Value())
}
