package animation

import (
	"cmp"
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/frames"
)

// Racer is one bar of a Race: its label, the value it has grown to, and its
// place in the order, 0 the top, fractional while it moves to a new one.
type Racer struct {
	Label string
	Value float64
	Place float64
}

// Race moves a set of bars to new values and new places together, a bar that
// overtakes another sliding past it: a leaderboard, the busiest services.
type Race struct {
	labels           []string
	fromV, toV       []float64
	fromPos, toPosns []float64
	dur              time.Duration
	ease             Ease
	p                float64
	run              frames.Run
}

// NewRace is labels at values, placed largest first, still until Set.
func NewRace(labels []string, values []float64, dur time.Duration) Race {
	r := Race{labels: slices.Clone(labels), dur: dur}
	r.toV = fit(values, len(labels))
	r.toPosns = places(r.toV)
	r.fromV, r.fromPos = slices.Clone(r.toV), slices.Clone(r.toPosns)
	r.p = 1
	return r
}

// fit is values cut or padded with zeros to n.
func fit(values []float64, n int) []float64 {
	out := make([]float64, n)
	copy(out, values)
	return out
}

// places is where each value stands, largest first, ties in label order.
func places(values []float64) []float64 {
	order := make([]int, len(values))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int { return cmp.Compare(values[b], values[a]) })
	out := make([]float64, len(values))
	for place, i := range order {
		out[i] = float64(place)
	}
	return out
}

// Set moves every bar from where it is now to values, and to its new place.
func (r Race) Set(values []float64) (Race, tea.Cmd) {
	now := r.Racers()
	r.fromV, r.fromPos = make([]float64, len(now)), make([]float64, len(now)) // the old value keeps its own
	for i, rc := range now {
		r.fromV[i], r.fromPos[i] = rc.Value, rc.Place
	}
	r.toV = fit(values, len(r.labels))
	r.toPosns = places(r.toV)
	r.run, r.p = frames.NewRun(r.dur), 0
	r.run.Ease = r.ease
	if !r.run.Active() {
		r.p = 1
	}
	return r, r.Tick()
}

// Racers is every bar as it stands now, in the order the labels were given;
// sort by Place to draw them top to bottom.
func (r Race) Racers() []Racer {
	q := frames.Or(r.run.Ease)(r.p)
	if r.p >= 1 {
		q = 1
	}
	out := make([]Racer, len(r.labels))
	for i, l := range r.labels {
		out[i] = Racer{
			Label: l,
			Value: r.fromV[i] + (r.toV[i]-r.fromV[i])*q,
			Place: r.fromPos[i] + (r.toPosns[i]-r.fromPos[i])*q,
		}
	}
	return out
}

// Active reports whether the bars are still moving.
func (r Race) Active() bool { return r.run.Active() }

// Tick is the command for r's next frame, nil once the bars settled.
func (r Race) Tick() tea.Cmd { return r.run.Tick() }

// FrameAt is r's frame for at, for a test or a host that delivers frames itself.
func (r Race) FrameAt(at time.Time) FrameMsg { return r.run.FrameAt(at) }

// Update moves on its own frames and asks for the next until the bars settle.
func (r Race) Update(msg tea.Msg) (Race, tea.Cmd) {
	run, p, ok, cmd := r.run.Update(msg)
	if ok {
		r.run, r.p = run, p
	}
	return r, cmd
}

// Advance is r at now, and whether the bars settled.
func (r Race) Advance(now time.Time) (Race, bool) {
	if !r.Active() {
		return r, true
	}
	r.run, r.p = r.run.Advance(now)
	return r, !r.Active()
}

// WithEase is r following e instead of EaseInOut, for every later Set.
func (r Race) WithEase(e Ease) Race {
	r.ease, r.run.Ease = e, e
	return r
}
