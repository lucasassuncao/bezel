package animation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func showFloat(f Float) string { return fmt.Sprintf("%.1f", f.cur) }

func TestRecordKeepsEveryFrameToTheEnd(t *testing.T) {
	frames := Record(NewFloat(0, 1, 50*time.Millisecond).WithEase(Linear), showFloat, time.Second)
	require.Equal(t, "0.0", frames[0])
	require.Equal(t, "1.0", frames[len(frames)-1])
	require.Less(t, len(frames), 10, "it stops when the animation ends")
}

func TestRecordWithANilViewRecordsNothing(t *testing.T) {
	require.Nil(t, Record(NewFloat(0, 1, 50*time.Millisecond), nil, time.Second))
}

func TestRecordStopsAtItsLimit(t *testing.T) {
	frames := Record(NewFloat(0, 1, 50*time.Millisecond).WithRepeat(Forever), showFloat, 5*Frame)
	require.Len(t, frames, 6)
}
