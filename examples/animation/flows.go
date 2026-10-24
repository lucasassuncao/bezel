package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/theme"
)

const (
	lanes    = 5 // bars the timeline staggers
	flowTime = 1800 * time.Millisecond
)

type flowMsg struct{} // r on the flow tab: play it all again

// flows is the sixth tab: a choreography of bars on one Timeline, and a
// value through Keyframes.
type flows struct {
	tl   animation.Timeline
	keys animation.Keyframes
}

func newFlows() flows { return flows{}.again() }

// again plays the three from the start.
func (f flows) again() flows {
	steps := make([]animation.Span, lanes)
	for i := range steps {
		steps[i] = animation.Step(fmt.Sprint(i), 0, track, 500*time.Millisecond).WithEase(animation.Out(animation.Cubic))
	}
	f.tl = animation.NewTimeline(animation.Seq(
		animation.Stagger(90*time.Millisecond, steps...),
		animation.Gap(300*time.Millisecond),
		animation.Step("fill", 0, 1, 400*time.Millisecond),
	))
	f.keys = animation.NewKeyframes(flowTime,
		animation.Key{At: 0, Value: 0},
		animation.Key{At: 0.35, Value: track - 1, Ease: animation.Out(animation.Back)},
		animation.Key{At: 0.6, Value: 8},
		animation.Key{At: 1, Value: 20, Ease: animation.Out(animation.Bounce)},
	)
	return f
}

func (f flows) advance(at time.Time) flows {
	f.tl, _ = f.tl.Advance(at)
	f.keys, _ = f.keys.Advance(at)
	return f
}

func (f flows) moving() bool { return f.tl.Active() || f.keys.Active() }

// rows is the tab's body: the staggered bars, then the dial.
func (f flows) rows(th theme.Resolved, label func(kind, variant string) string) string {
	var rows []string
	for i := range lanes {
		kind, variant := "", ""
		switch i {
		case 0:
			kind, variant = "timeline", "stagger 90ms"
		case 1:
			variant = "r replays"
		}
		cur := int(math.Round(f.tl.Value(fmt.Sprint(i))))
		style := th.Accent
		if f.tl.Value("fill") >= 1 {
			style = th.Success
		}
		cur = min(max(cur, 0), track)
		rows = append(rows, label(kind, variant)+style.Render(strings.Repeat("█", cur))+th.Muted.Render(strings.Repeat("░", track-cur)))
	}
	at := min(max(int(math.Round(f.keys.Cur)), 0), track-1)
	rows = append(rows, "",
		label("keyframes", "4 stops")+th.Muted.Render(strings.Repeat("·", at))+th.Accent.Render("●")+th.Muted.Render(strings.Repeat("·", track-1-at)))
	return strings.Join(rows, "\n")
}
