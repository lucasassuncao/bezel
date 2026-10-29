package list

import (
	"image/color"
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/internal/frames"
)

// motion is what a list with motion keeps of its last change: where the
// cursor glides from, which rows arrived and which left.
type motion struct {
	dur   time.Duration
	from  float64         // the row the cursor glides from
	fresh map[string]bool // labels that arrived with the last SetRows
	gone  []ghost         // rows the last SetRows took away
	p     float64
	run   frames.Run
}

// ghost is a row that left, drawn fading where it was until the motion ends.
type ghost struct {
	label string
	at    int
}

// WithMotion makes the list move instead of jump: the cursor glides to its
// new row, a row that arrives lights up and a row that leaves fades where
// it was, over dur. Update still returns no command: call Tick after it.
func (m Model[T]) WithMotion(dur time.Duration) Model[T] {
	m.mo.dur = max(dur, 0)
	return m
}

// start begins a motion from the cursor row from, when the list has motion.
func (m Model[T]) start(from float64) Model[T] {
	if m.mo.dur <= 0 {
		return m
	}
	if m.Active() { // glide on from where the highlight is drawn now
		from = m.mo.from + (from-m.mo.from)*frames.Or(m.mo.run.Ease)(m.mo.p)
	}
	m.mo.from = from
	m.mo.run, m.mo.p = frames.NewRun(m.mo.dur), 0
	if !m.mo.run.Active() {
		m.mo = motion{dur: m.mo.dur, p: 1}
	}
	return m
}

// changed is m after its rows went from before to the current ones: the
// labels that arrived are fresh and the ones that left are ghosts.
func (m Model[T]) changed(before []Row[T], cursor int) Model[T] {
	if m.mo.dur <= 0 {
		return m
	}
	had, has := map[string]bool{}, map[string]bool{}
	for _, r := range before {
		if !r.Section {
			had[r.Label] = true
		}
	}
	fresh := map[string]bool{}
	for _, r := range m.rows {
		if !r.Section {
			has[r.Label] = true
			if !had[r.Label] {
				fresh[r.Label] = true
			}
		}
	}
	var gone []ghost
	for i, r := range before {
		if !r.Section && !has[r.Label] {
			gone = append(gone, ghost{label: r.Label, at: i})
		}
	}
	if len(fresh) == 0 && len(gone) == 0 && cursor == m.win.Cursor {
		return m
	}
	m = m.start(float64(cursor))
	m.mo.fresh, m.mo.gone = fresh, gone
	return m
}

// glide is the row the cursor is drawn on now, fractional while it moves.
func (m Model[T]) glide() float64 {
	if !m.Active() {
		return float64(m.win.Cursor)
	}
	q := frames.Or(m.mo.run.Ease)(m.mo.p)
	return m.mo.from + (float64(m.win.Cursor)-m.mo.from)*q
}

// Active reports whether the list is still moving.
func (m Model[T]) Active() bool { return m.mo.run.Active() }

// Tick is the command for the list's next frame, nil while it is still: call
// it after Update or SetRows to start what they set moving.
func (m Model[T]) Tick() tea.Cmd { return m.mo.run.Tick() }

// FrameAt is the list's frame for at, for a test that delivers frames itself.
func (m Model[T]) FrameAt(at time.Time) tea.Msg { return m.mo.run.FrameAt(at) }

// Animate moves on the list's own frames and asks for the next until it
// settles; every other message leaves it alone.
func (m Model[T]) Animate(msg tea.Msg) (Model[T], tea.Cmd) {
	run, p, ok, cmd := m.mo.run.Update(msg)
	if !ok {
		return m, nil
	}
	m.mo.run, m.mo.p = run, p
	if !m.Active() {
		m.mo = motion{dur: m.mo.dur, p: 1}
	}
	return m, cmd
}

// cursorRow is the row the highlight is drawn on now.
func (m Model[T]) cursorRow() int {
	return int(math.Round(m.glide()))
}

// arriving is the style of a row that just arrived, fading from the accent
// to the row's own colour; nil when the row is settled.
func (m Model[T]) arriving(r Row[T], st *Styles) *lipgloss.Style {
	if !m.Active() || !m.mo.fresh[r.Label] {
		return nil
	}
	own := st.Row
	if r.Style != nil {
		own = *r.Style
	}
	s := own.Foreground(mix(st.Arrive.GetForeground(), own.GetForeground(), m.mo.p))
	return &s
}

// ghostLine is the ghost at row i, faded by how far the motion is, and false
// when no ghost sits there.
func (m Model[T]) ghostLine(i int, st *Styles) (string, bool) {
	if !m.Active() {
		return "", false
	}
	for _, g := range m.mo.gone {
		if g.at == i {
			fg := mix(st.Row.GetForeground(), st.Gone.GetForeground(), m.mo.p)
			return lipgloss.NewStyle().Foreground(fg).Strikethrough(true).Render("  " + g.label), true
		}
	}
	return "", false
}

// mix is a blended p of the way to b; a missing colour gives way to the other.
func mix(a, b color.Color, p float64) color.Color {
	if a == nil || b == nil {
		if p < 0.5 && a != nil {
			return a
		}
		return b
	}
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	at := func(x, y uint32) uint8 {
		return uint8((float64(x>>8) + (float64(y>>8)-float64(x>>8))*min(max(p, 0), 1)) + 0.5) // #nosec G115 -- within a byte
	}
	return color.RGBA{at(ar, br), at(ag, bg), at(ab, bb), 255}
}
