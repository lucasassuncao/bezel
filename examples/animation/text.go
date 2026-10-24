package main

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/animation/text"
	"github.com/lucasassuncao/bezel/theme"
)

const (
	listening = "listening for changes"
	boarding  = "GATE 12 · BOARDING"
	searched  = "the line a search found"
	chore     = "buy milk"
)

type (
	decryptMsg struct{} // d: the board lands again
	tickerMsg  struct{} // t: the price moves
	markMsg    struct{} // l: the marker runs again
	strikeMsg  struct{} // k: the chore is crossed out again
	copyMsg    struct{} // c: copied!
)

// prices are what the ticker moves between, up and down in turn.
var prices = []int{1284, 1302, 1297, 1311, 1288}

// texts is the fifth tab. The wave, the rainbow, the caret and the dots keep
// their own ticks; the rest run on the gallery's clock.
type texts struct {
	wave     text.Wave
	rainbow  text.Rainbow
	caret    text.Caret
	ellipsis text.Ellipsis
	decrypt  text.Decrypt
	ticker   text.Ticker
	price    int
	mark     text.Highlight
	strike   text.Strike
	rise     text.Rise
	th       theme.Resolved
}

func newTexts(th theme.Resolved) texts {
	return texts{
		wave:     text.NewWave(listening, 1, 1500*time.Millisecond),
		rainbow:  text.NewRainbow(title, 3*time.Second).WithTone(th.Accent.GetForeground()),
		caret:    text.NewCaret("▌", time.Second),
		ellipsis: text.NewEllipsis(3, 1200*time.Millisecond),
		decrypt:  text.NewDecrypt(boarding, 0),
		ticker: text.NewTicker(prices[0], th.Bold.GetForeground(), th.Success.GetForeground(),
			th.Danger.GetForeground(), 900*time.Millisecond),
		mark:   text.NewHighlight(searched, th.Accent.GetForeground(), 0),
		strike: text.NewStrike(chore, 0).WithInk(th.Muted.GetForeground()),
		rise:   text.NewRise("copied!", 2, th.Success.GetForeground(), black, 0),
		th:     th,
	}
}

// start sets the four that loop going on their own ticks, and lands the board.
func (x texts) start() (texts, tea.Cmd) {
	x = x.land()
	return x, tea.Batch(x.wave.Tick(), x.rainbow.Tick(), x.caret.Tick(), x.ellipsis.Tick())
}

func (x texts) land() texts {
	x.decrypt = text.NewDecrypt(boarding, 1400*time.Millisecond)
	return x
}

// move sends the ticker to the next price.
func (x texts) move() texts {
	x.price++
	x.ticker, _ = x.ticker.To(prices[x.price%len(prices)])
	return x
}

func (x texts) remark() texts {
	x.mark = text.NewHighlight(searched, x.th.Accent.GetForeground(), 900*time.Millisecond).WithEase(animation.EaseOut)
	return x
}

func (x texts) cross() texts {
	x.strike = text.NewStrike(chore, 600*time.Millisecond).WithInk(x.th.Muted.GetForeground())
	return x
}

func (x texts) copied() texts {
	x.rise = text.NewRise("copied!", 2, x.th.Success.GetForeground(), black, 900*time.Millisecond)
	return x
}

// update hands a frame to the four on their own ticks.
func (x texts) update(msg tea.Msg) (texts, tea.Cmd) {
	var a, b, c, d tea.Cmd
	x.wave, a = x.wave.Update(msg)
	x.rainbow, b = x.rainbow.Update(msg)
	x.caret, c = x.caret.Update(msg)
	x.ellipsis, d = x.ellipsis.Update(msg)
	return x, tea.Batch(a, b, c, d)
}

// advance moves the ones on the shared clock to at.
func (x texts) advance(at time.Time) texts {
	x.decrypt, _ = x.decrypt.Advance(at)
	x.ticker, _ = x.ticker.Advance(at)
	x.mark, _ = x.mark.Advance(at)
	x.strike, _ = x.strike.Advance(at)
	x.rise, _ = x.rise.Advance(at)
	return x
}

func (x texts) moving() bool {
	return x.decrypt.Active() || x.ticker.Active() || x.mark.Active() || x.strike.Active() || x.rise.Active()
}

// rows is the tab's body: a row for each, three for the wave and the rise.
func (x texts) rows(th theme.Resolved, label func(kind, variant string) string) string {
	var rows []string
	for i, line := range strings.Split(x.wave.View(), "\n") {
		kind, variant := "", ""
		if i == 1 {
			kind, variant = "wave", "forever"
		}
		rows = append(rows, label(kind, variant)+th.Info.Render(line))
	}
	rows = append(rows,
		label("rainbow", "turns forever")+x.rainbow.View(),
		label("decrypt", "d lands")+th.Bold.Render(x.decrypt.View()),
		label("caret", "blinks")+"Name: "+th.Bold.Render("lucas")+th.Accent.Render(x.caret.View()),
		label("ellipsis", "grows")+"Saving"+x.ellipsis.View(),
		label("ticker", "t moves")+th.Muted.Render("$")+x.ticker.View(),
		label("highlight", "l marks")+x.mark.View(),
		label("strike", "k crosses")+x.strike.View(),
	)
	for _, line := range strings.Split(x.rise.View(), "\n") {
		rows = append(rows, label("", "")+line)
	}
	return strings.Join(append(rows, label("rise", "c copies")+th.Accent.Render("[ copy link ]")), "\n")
}
