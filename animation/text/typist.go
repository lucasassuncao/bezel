package text

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/internal/edit"
	"github.com/lucasassuncao/bezel/internal/noise"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// stroke is the text a Typist shows from at onwards.
type stroke struct {
	at    time.Duration
	shown []string
}

// neighbours is, for each letter, the keys beside it on a QWERTY keyboard: a
// typo lands on one of them.
var neighbours = map[rune]string{
	'a': "qsz", 'b': "vgn", 'c': "xdv", 'd': "sfe", 'e': "wrd", 'f': "dgr", 'g': "fht",
	'h': "gjy", 'i': "uok", 'j': "hku", 'k': "jli", 'l': "ko", 'm': "nj", 'n': "bmh",
	'o': "ipl", 'p': "o", 'q': "wa", 'r': "etf", 's': "adw", 't': "ryg", 'u': "yij",
	'v': "cfb", 'w': "qes", 'x': "zcs", 'y': "tuh", 'z': "xa",
}

// Typist types text the way a person does: unevenly, slower after a word or
// a full stop, now and then hitting the key beside the right one and taking
// it back. The same text and seed type the same way every time.
type Typist struct {
	text    []string
	perChar time.Duration
	typos   float64
	seed    uint64
	script  []stroke
	shown   []string
	run     frames.Run
}

var _ bezel.Model[Typist] = Typist{}

// NewTypist types text at about one grapheme per perChar, with a typo in
// roughly one letter of every twenty; a perChar of 0 shows it at once.
func NewTypist(text string, perChar time.Duration) Typist {
	t := Typist{text: edit.Graphemes(text), perChar: perChar, typos: 0.05, seed: 1}
	return t.plan()
}

// WithTypos sets the share of letters typed wrong first, 0 for none.
func (t Typist) WithTypos(rate float64) Typist {
	t.typos = min(max(rate, 0), 1)
	return t.plan()
}

// WithSeed types the same text another way.
func (t Typist) WithSeed(seed uint64) Typist {
	t.seed = seed
	return t.plan()
}

// plan writes the keystrokes out ahead, and the run that plays them.
func (t Typist) plan() Typist {
	t.script = []stroke{{}}
	var at time.Duration
	var shown []string
	gap := func(k uint64, scale float64) time.Duration {
		jitter := 0.5 + noise.Unit(noise.Hash(t.seed, k))
		return time.Duration(float64(t.perChar) * jitter * scale)
	}
	push := func() {
		t.script = append(t.script, stroke{at: at, shown: append([]string(nil), shown...)})
	}
	for i, g := range t.text {
		k := uint64(i) * 4
		r, size := utf8.DecodeRuneInString(g)
		if wrong, ok := neighbours[unicode.ToLower(r)]; ok && size == len(g) && noise.Unit(noise.Hash(t.seed, k+1)) < t.typos {
			at += gap(k+2, 1)
			shown = append(shown, string(wrong[noise.Hash(t.seed, k+3)%uint64(len(wrong))]))
			push()
			at += gap(k+2, 4) // a moment to notice
			shown = shown[:len(shown)-1]
			push()
		}
		scale := 1.0
		if i > 0 && strings.ContainsAny(t.text[i-1], " ,") {
			scale = 2
		}
		if i > 0 && strings.ContainsAny(t.text[i-1], ".!?") {
			scale = 5
		}
		at += gap(k, scale)
		shown = append(shown, g)
		push()
	}
	delay := t.run.Delay
	t.run = frames.Run{}
	if len(t.text) > 0 && t.perChar > 0 {
		t.run = frames.NewRun(at)
		t.run.Delay = delay
	}
	t.shown = nil
	if !t.run.Active() {
		t.shown = t.text
	}
	return t
}

// View is the text as typed so far, typos and all.
func (t Typist) View() string { return strings.Join(t.shown, "") }

// Active reports whether t is still typing.
func (t Typist) Active() bool { return t.run.Active() }

// Tick is the command for t's next frame, nil once it is done.
func (t Typist) Tick() tea.Cmd { return t.run.Tick() }

// FrameAt is t's frame for at, for a test or a host that delivers frames itself.
func (t Typist) FrameAt(at time.Time) animation.FrameMsg { return t.run.FrameAt(at) }

// Update types on its own frames and asks for the next until it is done.
func (t Typist) Update(msg tea.Msg) (Typist, bezel.Result) {
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

// Advance is t at now, and whether it is done.
func (t Typist) Advance(now time.Time) (Typist, bool) {
	if !t.Active() {
		return t, true
	}
	total := t.run.Dur
	var p float64
	if t.run, p = t.run.Advance(now); !t.run.Active() {
		t.shown = t.text
		return t, true
	}
	at := time.Duration(p * float64(total))
	for _, s := range t.script {
		if s.at <= at {
			t.shown = s.shown
		}
	}
	return t, false
}

// WithDelay holds t empty for d after its first frame.
func (t Typist) WithDelay(d time.Duration) Typist {
	t.run.Delay = max(d, 0)
	return t
}
