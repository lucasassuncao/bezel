package main

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/animation/text"
	"github.com/lucasassuncao/bezel/theme"
)

const (
	answer   = "Streaming arrives in bursts, but it reads at an even pace. "
	sentence = "Typed by hand, mistakes and all."
	secret   = "this message will self-destruct"
	dropped  = "BEZEL"
)

type (
	streamMsg struct{} // s: another piece of the answer arrives
	morphMsg  struct{} // m: the status changes word
	typistMsg struct{} // y: the sentence is typed again
	flipMsg   struct{} // f: the clock turns a minute
	burnMsg   struct{} // b: the message burns
	fallMsg   struct{} // a: the letters fall again
)

// statuses are the words the morph turns between.
var statuses = []string{"saving…", "saved", "syncing…", "synced"}

// words is the fourth tab, arrive: text that streams, morphs, is typed, flips,
// glows, burns, and falls. The stream and the neon keep their own frames.
type words struct {
	stream  text.Stream
	morph   text.Morph
	status  int
	typist  text.Typist
	flip    text.Flip
	minute  int
	neon    text.Neon
	burn    text.Burn
	burning bool // the message has been set alight
	fall    text.Fall
	pieces  int
	neonOn  color.Color
	neonOff color.Color
}

func newWords(th theme.Resolved) words {
	w := words{
		stream:  text.NewStream(25 * time.Millisecond),
		morph:   text.NewMorph(statuses[0], statuses[0], 0),
		typist:  text.NewTypist(sentence, 0),
		flip:    text.NewFlip(clockAt(0), clockAt(0), 0),
		fall:    text.NewFall(dropped, 2, 0),
		neonOn:  th.Danger.GetForeground(),
		neonOff: black,
	}
	w.neon = text.NewNeon("OPEN 24/7", w.neonOn, w.neonOff)
	return w
}

// clockAt is a clock face minutes after 12:45.
func clockAt(minutes int) string {
	m := 45 + minutes
	return fmt.Sprintf("%02d:%02d", 12+m/60, m%60)
}

func (w words) start() (words, tea.Cmd) {
	w, cmd := w.more()
	return w, tea.Batch(cmd, w.neon.Tick())
}

// more pushes the next piece of the answer onto the stream.
func (w words) more() (words, tea.Cmd) {
	w.pieces++
	var cmd tea.Cmd
	w.stream, cmd = w.stream.Push(answer)
	return w, cmd
}

func (w words) changeWord() words {
	w.status = (w.status + 1) % len(statuses)
	w.morph = text.NewMorph(statuses[(w.status+len(statuses)-1)%len(statuses)], statuses[w.status], 600*time.Millisecond)
	return w
}

func (w words) retype() words {
	w.typist = text.NewTypist(sentence, 45*time.Millisecond).WithSeed(uint64(time.Now().UnixNano())) // #nosec G115 -- any seed
	return w
}

func (w words) turnClock() words {
	w.minute++
	w.flip = text.NewFlip(clockAt(w.minute-1), clockAt(w.minute), 500*time.Millisecond)
	return w
}

func (w words) burnIt() words {
	w.burn, w.burning = text.NewBurn(secret, 1500*time.Millisecond), true
	return w
}

func (w words) drop() words {
	w.fall = text.NewFall(dropped, 2, 900*time.Millisecond)
	return w
}

// update takes the stream's and the neon's own frames.
func (w words) update(msg tea.Msg) (words, tea.Cmd) {
	var a, b tea.Cmd
	w.stream, a = w.stream.Update(msg)
	w.neon, b = w.neon.Update(msg)
	return w, tea.Batch(a, b)
}

func (w words) advance(at time.Time) words {
	w.morph, _ = w.morph.Advance(at)
	w.typist, _ = w.typist.Advance(at)
	w.flip, _ = w.flip.Advance(at)
	w.burn, _ = w.burn.Advance(at)
	w.fall, _ = w.fall.Advance(at)
	return w
}

func (w words) moving() bool {
	return w.morph.Active() || w.typist.Active() || w.flip.Active() || w.burn.Active() || w.fall.Active()
}

// rows is the tab's body: a row for each, three for the big clock and the fall.
func (w words) rows(th theme.Resolved, label func(kind, variant string) string) string {
	tail := w.stream.View()
	if over := ansi.StringWidth(tail) - 50; over > 0 {
		tail = "…" + ansi.Cut(tail, over+1, ansi.StringWidth(tail))
	}
	rows := []string{
		label("stream", "s more") + th.Bold.Render(tail),
		label("morph", "m changes") + th.Success.Render(w.morph.View()),
		label("typist", "y retypes") + w.typist.View(),
		label("flip", "f turns") + th.Accent.Render(w.flip.View()),
		label("neon", "flickers") + w.neon.View(),
		label("burn", "b burns") + w.burnt(),
		"",
	}
	for i, line := range strings.Split(text.Big(w.flip.View()), "\n") {
		kind, variant := "", ""
		if i == 1 {
			kind, variant = "big", "f turns"
		}
		rows = append(rows, label(kind, variant)+th.Accent.Render(line))
	}
	rows = append(rows, "")
	for i, line := range strings.Split(w.fall.View(), "\n") {
		kind, variant := "", ""
		if i == 2 {
			kind, variant = "fall", "a drops"
		}
		rows = append(rows, label(kind, variant)+th.Bold.Render(line))
	}
	return strings.Join(rows, "\n")
}

// burnt is the message, whole until it is set alight.
func (w words) burnt() string {
	if !w.burning {
		return secret
	}
	return w.burn.View()
}
