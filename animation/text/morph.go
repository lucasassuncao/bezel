package text

import (
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/edit"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// op is one grapheme of a morph: kept, taken out or put in.
type op struct {
	g    string
	kind byte // '=' kept, '-' removed, '+' added
}

// Morph turns one text into another, taking out the letters the new one
// lacks and then putting in the ones it adds, keeping what the two share:
// "saving…" becomes "saved" without a flash of nothing between.
type Morph struct {
	ops      []op
	dels     int // ops of kind '-'
	adds     int // ops of kind '+'
	dur      time.Duration
	p        float64
	run      frames.Run
	from, to string
}

// NewMorph turns from into to over dur.
func NewMorph(from, to string, dur time.Duration) Morph {
	m := Morph{ops: align(edit.Graphemes(from), edit.Graphemes(to)), dur: dur, from: from, to: to}
	for _, o := range m.ops {
		switch o.kind {
		case '-':
			m.dels++
		case '+':
			m.adds++
		}
	}
	if m.dels+m.adds > 0 {
		m.run = frames.NewRun(dur)
	}
	if !m.Active() {
		m.p = 1
	}
	return m
}

// To morphs on from the text m is heading for to next, over the same duration.
func (m Morph) To(next string) (Morph, tea.Cmd) {
	n := NewMorph(m.to, next, m.dur)
	return n, n.Tick()
}

// align is the shortest edit from a to b, by longest common subsequence.
func align(a, b []string) []op {
	n, k := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, k+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := k - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var ops []op
	i, j := 0, 0
	for i < n || j < k {
		switch {
		case i < n && j < k && a[i] == b[j]:
			ops = append(ops, op{a[i], '='})
			i, j = i+1, j+1
		case j < k && (i == n || lcs[i][j+1] >= lcs[i+1][j]):
			ops = append(ops, op{b[j], '+'})
			j++
		default:
			ops = append(ops, op{a[i], '-'})
			i++
		}
	}
	return ops
}

// View is the text between the two: the removals going first, right to left,
// then the additions, left to right.
func (m Morph) View() string {
	share := float64(m.dels) / float64(max(m.dels+m.adds, 1))
	gone, put := 0, 0
	switch {
	case m.p >= 1:
		gone, put = m.dels, m.adds
	case m.p < share:
		gone = int(math.Floor(m.p / share * float64(m.dels)))
	default:
		gone = m.dels
		put = int(math.Floor((m.p - share) / (1 - share) * float64(m.adds)))
	}
	var b strings.Builder
	del, add := 0, 0
	for _, o := range m.ops {
		switch o.kind {
		case '=':
			b.WriteString(o.g)
		case '-':
			if del < m.dels-gone {
				b.WriteString(o.g)
			}
			del++
		case '+':
			if add < put {
				b.WriteString(o.g)
			}
			add++
		}
	}
	return b.String()
}

// Active reports whether m is still changing.
func (m Morph) Active() bool { return m.run.Active() }

// Tick is the command for m's next frame, nil once it landed.
func (m Morph) Tick() tea.Cmd { return m.run.Tick() }

// FrameAt is m's frame for at, for a test or a host that delivers frames itself.
func (m Morph) FrameAt(at time.Time) animation.FrameMsg { return m.run.FrameAt(at) }

// Update changes on its own frames and asks for the next until it lands.
func (m Morph) Update(msg tea.Msg) (Morph, tea.Cmd) {
	r, p, ok, cmd := m.run.Update(msg)
	if ok {
		m.run, m.p = r, p
	}
	return m, cmd
}

// Advance is m at now, and whether it landed.
func (m Morph) Advance(now time.Time) (Morph, bool) {
	if !m.Active() {
		return m, true
	}
	m.run, m.p = m.run.Advance(now)
	return m, !m.Active()
}

// WithDelay holds the old text for d after the first frame.
func (m Morph) WithDelay(d time.Duration) Morph {
	m.run.Delay = max(d, 0)
	return m
}
