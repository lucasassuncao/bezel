package form

import (
	"math"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/theme"
)

// scale is a value held within lo and hi, moved step at a time and written
// with as many decimals as step has.
type scale struct {
	v, lo, hi, step float64
}

// clamp is v within lo and hi; NaN keeps the scale's own value.
func (s scale) clamp(v float64) float64 {
	if math.IsNaN(v) {
		return s.v
	}
	return min(max(v, s.lo), s.hi)
}

// moved is the scale n steps on, held at its ends.
func (s scale) moved(n int) scale {
	s.v = s.clamp(s.v + float64(n)*s.step)
	return s
}

// String is v with step's decimals.
func (s scale) String() string {
	d := 0
	if i := strings.IndexByte(strconv.FormatFloat(s.step, 'f', -1, 64), '.'); i >= 0 {
		d = len(strconv.FormatFloat(s.step, 'f', -1, 64)) - i - 1
	}
	return strconv.FormatFloat(s.v, 'f', d, 64)
}

// withRange is the scale over lo to hi, its value kept within them.
func (s scale) withRange(lo, hi float64) scale {
	s.lo, s.hi = lo, max(lo, hi)
	s.v = s.clamp(s.v)
	return s
}

// withStep is the scale moving step at a time; one not above zero is 1.
func (s scale) withStep(step float64) scale {
	if step <= 0 || math.IsNaN(step) {
		step = 1
	}
	s.step = step
	return s
}

// Number is a number typed or stepped: Up and Down move it a step within its
// range, digits, '.' and '-' type it, and leaving the field settles it.
type Number struct {
	label string
	sc    scale
	text  string // what is typed, the value written once it settles
	st    *Styles
	keys  Keys
}

// NewNumber offers value, stepping by 1 with no bounds.
func NewNumber(label string, value float64) Number {
	sc := scale{v: value, lo: math.Inf(-1), hi: math.Inf(1), step: 1}
	return Number{label: label, sc: sc, text: sc.String(), keys: DefaultKeys()}
}

// WithRange keeps the value from lo to hi, both included.
func (n Number) WithRange(lo, hi float64) Number {
	n.sc = n.sc.withRange(lo, hi)
	n.text = n.sc.String()
	return n
}

// WithStep moves the value step at a time and writes it with step's
// decimals; a step not above zero is 1.
func (n Number) WithStep(step float64) Number {
	n.sc = n.sc.withStep(step)
	n.text = n.sc.String()
	return n
}

// WithKeys rebinds the field: Up and Down step, Erase takes a character.
func (n Number) WithKeys(k Keys) Number {
	n.keys = k
	return n
}

func (n Number) rebind(k Keys) Field { return n.WithKeys(k) }

func (n Number) retheme(_ *theme.Resolved, st *Styles) Field {
	n.st = st
	return n
}

// Value is the number, the last one that parsed while text is typed.
func (n Number) Value() float64 { return n.sc.v }
func (n Number) Label() string  { return n.label }

func (n Number) Focus() (Field, tea.Cmd) { return n, nil }

// Blur settles what was typed: held within the range, or back to the last
// number when it does not parse.
func (n Number) Blur() Field {
	n.text = n.sc.String()
	return n
}

// Update steps on Up and Down and takes a typed digit, '.', '-' or Erase.
func (n Number) Update(msg tea.Msg) (Field, bezel.Result) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return n, bezel.Result{}
	}
	switch t := edit.Typed(k); {
	case is(msg, n.keys.Up):
		n.sc = n.sc.moved(1)
		n.text = n.sc.String()
	case is(msg, n.keys.Down):
		n.sc = n.sc.moved(-1)
		n.text = n.sc.String()
	case is(msg, n.keys.Erase):
		n = n.typed(edit.Erase(n.text))
	case t != "" && strings.ContainsAny(t, "0123456789.-"):
		n = n.typed(n.text + t)
	default:
		return n, bezel.Result{}
	}
	return n, bezel.Result{Handled: true}
}

// typed is text on the field, the value following it while it parses.
func (n Number) typed(text string) Number {
	n.text = text
	if v, err := strconv.ParseFloat(text, 64); err == nil {
		n.sc.v = n.sc.clamp(v)
	}
	return n
}

func (n Number) View(focused bool, labelW, width int) string {
	edge := styles(n.st).Edge
	if focused {
		edge = styles(n.st).FocusedEdge
	}
	return row(styles(n.st), focused, n.label, labelW, width, edge.Render("‹ ")+n.text+edge.Render(" ›"))
}

// Slider is a value along a bar: Left and Right move it a step within its
// range, from 0 to 100 until WithRange says otherwise.
type Slider struct {
	label string
	sc    scale
	width int
	st    *Styles
	set   icon.Set
	keys  Keys
}

// NewSlider offers value from 0 to 100, stepping by 1 on a bar 20 cells wide.
func NewSlider(label string, value float64) Slider {
	sc := scale{v: value, lo: 0, hi: 100, step: 1}
	sc.v = sc.clamp(value)
	return Slider{label: label, sc: sc, width: 20, set: icon.FromEnv(), keys: DefaultKeys()}
}

// WithRange keeps the value from lo to hi, both included.
func (s Slider) WithRange(lo, hi float64) Slider {
	s.sc = s.sc.withRange(lo, hi)
	return s
}

// WithStep moves the value step at a time and writes it with step's
// decimals; a step not above zero is 1.
func (s Slider) WithStep(step float64) Slider {
	s.sc = s.sc.withStep(step)
	return s
}

// WithWidth draws the bar n cells wide, at least 2.
func (s Slider) WithWidth(n int) Slider {
	s.width = max(n, 2)
	return s
}

// WithIcons draws the bar from set instead of icon.FromEnv's.
func (s Slider) WithIcons(set icon.Set) Slider {
	s.set = set
	return s
}

// WithKeys rebinds the slider: Left and Right move it.
func (s Slider) WithKeys(k Keys) Slider {
	s.keys = k
	return s
}

func (s Slider) rebind(k Keys) Field { return s.WithKeys(k) }

func (s Slider) retheme(_ *theme.Resolved, st *Styles) Field {
	s.st = st
	return s
}

// Value is where the slider stands.
func (s Slider) Value() float64          { return s.sc.v }
func (s Slider) Label() string           { return s.label }
func (s Slider) Focus() (Field, tea.Cmd) { return s, nil }
func (s Slider) Blur() Field             { return s }

// Update moves a step on Left and Right, which are its keys even at the ends.
func (s Slider) Update(msg tea.Msg) (Field, bezel.Result) {
	switch {
	case is(msg, s.keys.Left):
		s.sc = s.sc.moved(-1)
	case is(msg, s.keys.Right):
		s.sc = s.sc.moved(1)
	default:
		return s, bezel.Result{}
	}
	return s, bezel.Result{Handled: true}
}

// View is the bar, filled up to a knob where the value stands, then the value.
func (s Slider) View(focused bool, labelW, width int) string {
	st := styles(s.st)
	at := 0
	if span := s.sc.hi - s.sc.lo; span > 0 {
		at = int(math.Round((s.sc.v - s.sc.lo) / span * float64(s.width-1)))
	}
	bar := st.Chosen.Render(strings.Repeat(s.set.Line, at)+s.set.Running) +
		st.Edge.Render(strings.Repeat(s.set.Line, s.width-1-at))
	return row(st, focused, s.label, labelW, width, bar+" "+s.sc.String())
}
