package animation

import "time"

// advancer is any animation that moves itself to a time.
type advancer[T any] interface {
	Advance(now time.Time) (T, bool)
}

// Record plays a from its first frame to its end, or for limit at most, one
// frame every Frame, and keeps what view draws of each: an animation as text,
// for a test that checks it frame by frame or a recording built from it.
func Record[T advancer[T]](a T, view func(T) string, limit time.Duration) []string {
	start := time.Unix(1_000_000, 0)
	var frames []string
	for at := time.Duration(0); at <= limit; at += Frame {
		var done bool
		a, done = a.Advance(start.Add(at))
		frames = append(frames, view(a))
		if done {
			break
		}
	}
	return frames
}
