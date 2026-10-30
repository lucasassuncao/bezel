package animation

import (
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// track is one named value moving over a stretch of a timeline.
type track struct {
	name     string
	from, to float64
	start    time.Duration
	dur      time.Duration
	ease     Ease
}

// Span is a piece of a timeline: steps, and how long the piece lasts. Seq,
// Par and Stagger build bigger spans out of smaller ones.
type Span struct {
	tracks []track
	length time.Duration
}

// Step moves the value called name from from to to over dur.
func Step(name string, from, to float64, dur time.Duration) Span {
	dur = max(dur, 0)
	return Span{tracks: []track{{name: name, from: from, to: to, dur: dur}}, length: dur}
}

// Gap is a pause of d, for a Seq that waits between two steps.
func Gap(d time.Duration) Span { return Span{length: max(d, 0)} }

// WithEase is s with every step in it following e.
func (s Span) WithEase(e Ease) Span {
	s.tracks = slices.Clone(s.tracks)
	for i := range s.tracks {
		s.tracks[i].ease = e
	}
	return s
}

// shift is s starting d later.
func (s Span) shift(d time.Duration) []track {
	out := slices.Clone(s.tracks)
	for i := range out {
		out[i].start += d
	}
	return out
}

// Seq plays spans one after another.
func Seq(spans ...Span) Span {
	var out Span
	for _, s := range spans {
		out.tracks = append(out.tracks, s.shift(out.length)...)
		out.length += s.length
	}
	return out
}

// Par plays spans together, lasting as long as the longest.
func Par(spans ...Span) Span { return Stagger(0, spans...) }

// Stagger plays spans together, each starting gap after the one before: a
// list whose rows come in one by one.
func Stagger(gap time.Duration, spans ...Span) Span {
	var out Span
	for i, s := range spans {
		at := time.Duration(i) * max(gap, 0)
		out.tracks = append(out.tracks, s.shift(at)...)
		out.length = max(out.length, at+s.length)
	}
	return out
}

// Timeline plays a Span on one clock and reports where each named value is:
// a choreography of several values without a delay worked out by hand.
type Timeline struct {
	span    Span
	elapsed time.Duration // into the span, after the run's loop
	run     frames.Run
}

var _ bezel.Model[Timeline] = Timeline{}

// NewTimeline plays s once, its clock started by its first frame. Empty, or
// with motion Still, it is inactive with every value at its end.
func NewTimeline(s Span) Timeline {
	t := Timeline{span: s, elapsed: s.length}
	if s.length > 0 {
		t.run = frames.NewRun(s.length)
	}
	if t.run.Active() {
		t.elapsed = 0
	}
	return t
}

// WithRepeat plays t n more times after the first, or without end for Forever.
func (t Timeline) WithRepeat(n int) Timeline {
	t.run.Repeat = max(n, Forever)
	return t
}

// WithYoyo plays every other pass of t backwards.
func (t Timeline) WithYoyo(on bool) Timeline {
	t.run.Yoyo = on
	return t
}

// Value is where the value called name is now: the last step of that name
// to have started, or the first step's start before any has.
func (t Timeline) Value(name string) float64 {
	v, _ := t.at(name)
	return v
}

// Progress is how far the current step of name is, in [0,1] before its ease.
func (t Timeline) Progress(name string) float64 {
	_, p := t.at(name)
	return p
}

// at is name's value and its step's raw progress at the elapsed time.
func (t Timeline) at(name string) (float64, float64) {
	var cur *track
	for i := range t.span.tracks {
		tr := &t.span.tracks[i]
		if tr.name != name {
			continue
		}
		if cur == nil || tr.start <= t.elapsed && tr.start >= cur.start {
			cur = tr
		}
	}
	if cur == nil {
		return 0, 0
	}
	if t.elapsed < cur.start {
		return cur.from, 0
	}
	p := 1.0
	if cur.dur > 0 {
		p = min(float64(t.elapsed-cur.start)/float64(cur.dur), 1)
	}
	return cur.from + (cur.to-cur.from)*frames.Or(cur.ease)(p), p
}

// Elapsed is how far into the span t is.
func (t Timeline) Elapsed() time.Duration { return t.elapsed }

// Active reports whether t is still playing.
func (t Timeline) Active() bool { return t.run.Active() }

// Tick is the command for t's next frame, nil once it ended.
func (t Timeline) Tick() tea.Cmd { return t.run.Tick() }

// FrameAt is t's frame for at, for a test or a host that delivers frames itself.
func (t Timeline) FrameAt(at time.Time) FrameMsg { return t.run.FrameAt(at) }

// Update plays on its own frames and asks for the next until it ends.
func (t Timeline) Update(msg tea.Msg) (Timeline, bezel.Result) {
	f, ok := t.run.Own(msg)
	if !ok {
		return t, bezel.Result{Handled: ok}
	}
	if t, _ = t.Advance(f.At); !t.Active() {
		return t, bezel.Result{Handled: true}
	}
	var cmd tea.Cmd
	t.run, cmd = t.run.Next()
	return t, bezel.Result{Handled: true, Cmd: cmd}
}

// Advance is t at now, and whether it ended.
func (t Timeline) Advance(now time.Time) (Timeline, bool) {
	if !t.Active() {
		return t, true
	}
	var p float64
	t.run, p = t.run.Advance(now)
	t.elapsed = time.Duration(p * float64(t.span.length))
	return t, !t.Active()
}

// delayer is any animation that can wait before it starts.
type delayer[T any] interface {
	WithDelay(d time.Duration) T
}

// Cascade delays each of items gap more than the one before, so animations
// built apart start one after another: the rows of a list sliding in.
func Cascade[T delayer[T]](gap time.Duration, items ...T) []T {
	out := make([]T, len(items))
	for i, it := range items {
		out[i] = it.WithDelay(time.Duration(i) * max(gap, 0))
	}
	return out
}
