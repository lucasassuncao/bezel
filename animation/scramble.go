package animation

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/edit"
)

// letters is what an unsettled letter shows by default.
const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// Scramble shows text as churning noise that settles from the left, letter by
// letter: a token, a hash or a code that seems decoded rather than printed.
type Scramble struct {
	text    []string // the graphemes
	charset []rune
	p       float64
	seed    uint64 // the last frame's time: the noise is a function of it, not of a source
	run     run
}

// NewScramble settles text over dur. Letters and digits churn; spaces and
// punctuation stand still, so the shape of the text reads from the start.
func NewScramble(text string, dur time.Duration) Scramble {
	s := Scramble{text: edit.Graphemes(text), charset: []rune(letters), run: newRun(dur)}
	s.run.ease = Linear
	return s
}

// WithCharset draws the noise from chars instead of letters and digits.
func (s Scramble) WithCharset(chars string) Scramble {
	if r := []rune(chars); len(r) > 0 {
		s.charset = r
	}
	return s
}

// WithDelay holds every letter in noise for d before the first settles.
func (s Scramble) WithDelay(d time.Duration) Scramble {
	s.run.delay = max(d, 0)
	return s
}

// View is the text with its unsettled letters drawn from the charset.
func (s Scramble) View() string {
	settled := len(s.text)
	if s.Active() {
		settled = int(s.p * float64(len(s.text)))
	}
	var b strings.Builder
	var k uint64 // i as the hash takes it
	for i, g := range s.text {
		k++
		if i < settled || !churns(g) {
			b.WriteString(g)
			continue
		}
		b.WriteRune(s.charset[noise(s.seed, k)%uint64(len(s.charset))])
	}
	return b.String()
}

// churns reports whether g is a letter or digit, the graphemes the noise replaces.
func churns(g string) bool {
	r, size := utf8.DecodeRuneInString(g)
	return size == len(g) && (unicode.IsLetter(r) || unicode.IsDigit(r))
}

// noise is a hash of seed and k (splitmix64): the same frame draws the same
// letters, the next frame others.
func noise(seed, k uint64) uint64 {
	z := seed + k*0x9E3779B97F4A7C15
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

// Active reports whether any letter is still churning.
func (s Scramble) Active() bool { return s.run.active() }

// Tick is the command for s's next frame, nil once it settled.
func (s Scramble) Tick() tea.Cmd { return s.run.tick() }

// FrameAt is s's frame for at, for a test or a host that delivers frames itself.
func (s Scramble) FrameAt(at time.Time) FrameMsg { return s.run.frameAt(at) }

// Update churns on its own frames and asks for the next until it settles.
func (s Scramble) Update(msg tea.Msg) (Scramble, tea.Cmd) {
	f, ok := s.run.own(msg)
	if !ok {
		return s, nil
	}
	if s, _ = s.Advance(f.At); !s.Active() {
		return s, nil
	}
	var cmd tea.Cmd
	s.run, cmd = s.run.next()
	return s, cmd
}

// Advance is s at now, and whether it settled.
func (s Scramble) Advance(now time.Time) (Scramble, bool) {
	if !s.Active() {
		return s, true
	}
	s.seed = uint64(now.UnixNano()) // #nosec G115 -- any bits will do for noise
	s.run, s.p = s.run.advance(now)
	return s, !s.Active()
}
