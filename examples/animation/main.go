// Animation: every kind the animation package has, a tab for each family. The timed
// ones share one Clock; the looping ones and the hints pane keep their own.
package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/animation/text"
	"github.com/lucasassuncao/bezel/component/progress"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// hintsHeight is the open pane, borders included.
const hintsHeight = 8

const (
	slide   = 300 * time.Millisecond
	race    = 1500 * time.Millisecond
	track   = 32 // cells in a bar or a spring's track
	nudge   = 8  // cells an arrow moves the springs' target
	counts  = 12345
	blocks  = 6                      // in the cascade
	block   = 5                      // cells each one grows to
	stagger = 120 * time.Millisecond // between one block starting and the next
)

const (
	saved    = "✓ saved"
	headline = "Copying 1,204 files to the backup drive…"
	longPath = "~/Downloads/Media/Images/wallhaven-2026/landscapes/aurora-over-the-fjord.png"
	password = "[ password: •••••• ]"
)

type (
	startMsg  struct{}
	toggleMsg struct{}
	replayMsg struct{}
	fadeMsg   struct{}
	typeMsg   struct{}
	shakeMsg  struct{}
	pulseMsg  struct{}
)

// page is a tab of the gallery: its name and the keys that act on it.
type page struct {
	name    string
	actions []shell.Action
}

func (p page) Name() string                         { return p.name }
func (p page) Status(shell.Context) string          { return "" }
func (p page) Actions(shell.Context) []shell.Action { return p.actions }

// eases are the curves the bars race on, in the order they are drawn.
var eases = []struct {
	name string
	ease animation.Ease
}{
	{"linear", animation.Linear},
	{"ease in", animation.EaseIn},
	{"ease out", animation.EaseOut},
	{"ease in-out", animation.EaseInOut},
}

type model struct {
	sh    shell.Shell
	th    theme.Resolved
	clock animation.Clock // drives everything below it but the marquee and the hints

	bars           []animation.Tween
	steady, bouncy animation.Spring
	smooth         animation.Float
	count          animation.Tween
	fade           animation.Fade
	typed          text.Typewriter
	shake          text.Shake
	pulse          animation.Fade
	cascade        []animation.Tween
	marquee        text.Marquee
	fx             effects // the second tab's
	cv             curves  // the third tab's
	dr             drawing // the fourth tab's
	tx             texts
	lt             lights
	sc             scenes
	sw             swaps
	ph             physics
	dt             datas

	open   bool
	tween  animation.Tween // the hints pane's height, on its own frames
	height int
}

// shape is the layout with the hints pane at h rows; at 0 it is left out.
func shape(h int) layout.Node {
	if h == 0 {
		return layout.Rows(layout.Fill("main"))
	}
	return layout.Rows(layout.Fill("main"), layout.Fixed("hints", layout.Lines(h)))
}

func newModel() model {
	th := theme.Resolve(theme.Mint, true)
	sh := shell.New(shell.Config{
		Layout:     shape(0),
		Theme:      th,
		Title:      "animation",
		Transition: shell.SlideTabs(300 * time.Millisecond),
		Tabs: []shell.Tab{
			page{"basics", []shell.Action{
				shell.Custom("r", "replay", shell.Send(replayMsg{})),
				shell.Custom("←/→", "spring", nil, shell.DisplayOnly()),
				shell.Custom("f", "fade", shell.Send(fadeMsg{})),
				shell.Custom("t", "type", shell.Send(typeMsg{})),
				shell.Custom("s", "shake", shell.Send(shakeMsg{})),
				shell.Custom("p", "pulse", shell.Send(pulseMsg{})),
			}},
			page{"effects", []shell.Action{
				shell.Custom("x", "scramble", shell.Send(scrambleMsg{})),
				shell.Custom("o", "odometer", shell.Send(odometerMsg{})),
				shell.Custom("m", "move", shell.Send(moveMsg{})),
			}},
			page{"curves", []shell.Action{
				shell.Custom("r", "race", shell.Send(curvesMsg{})),
				shell.Custom("i", "in/out", shell.Send(turnMsg{})),
			}},
			page{"draw", []shell.Action{
				shell.Custom("e", "effect", shell.Send(effectMsg{})),
				shell.Custom("p", "pattern", shell.Send(patternMsg{})),
				shell.Custom("r", "again", shell.Send(againMsg{})),
			}},
			page{"text", []shell.Action{
				shell.Custom("d", "decrypt", shell.Send(decryptMsg{})),
				shell.Custom("t", "ticker", shell.Send(tickerMsg{})),
				shell.Custom("l", "highlight", shell.Send(markMsg{})),
				shell.Custom("k", "strike", shell.Send(strikeMsg{})),
				shell.Custom("c", "copy", shell.Send(copyMsg{})),
			}},
			page{"light", []shell.Action{
				shell.Custom("e", "light", shell.Send(lightMsg{})),
				shell.Custom("r", "again", shell.Send(shineMsg{})),
				shell.Custom("f", "diff", shell.Send(diffMsg{})),
				shell.Custom("u", "undo", shell.Send(undoMsg{})),
			}},
			page{"scene", []shell.Action{
				shell.Custom("n", "next scene", shell.Send(sceneMsg{})),
			}},
			page{"swap", []shell.Action{
				shell.Custom("n", "next", shell.Send(swapMsg{})),
				shell.Custom("r", "back", shell.Send(againSwap{})),
			}},
			page{"motion", []shell.Action{
				shell.Custom("d", "drop", shell.Send(dropMsg{})),
				shell.Custom("f", "flock", shell.Send(flockMsg{})),
				shell.Custom("b", "rubber", shell.Send(rubberMsg{})),
			}},
			page{"data", []shell.Action{
				shell.Custom("g", "grow", shell.Send(growMsg{})),
				shell.Custom("p", "push", shell.Send(pushMsg{})),
				shell.Custom("n", "gauge", shell.Send(gaugeMsg{})),
				shell.Custom("w", "warm", shell.Send(warmMsg{})),
			}},
		},
		Actions: []shell.Action{
			shell.ChangeTab(),
			shell.Custom("h", "hints", shell.Send(toggleMsg{})),
			shell.Quit(),
		},
	})
	muted := th.Muted.GetForeground()
	return model{
		sh: sh, th: th, clock: animation.NewClock(),
		steady: animation.NewSpring(0), bouncy: animation.NewSpring(0).WithDamping(0.3),
		fade:    animation.NewFade(muted, muted, 0),
		typed:   text.NewTypewriter(headline, 0),
		shake:   text.NewShake(3, 0),
		pulse:   animation.NewFade(muted, muted, 0),
		marquee: text.NewMarquee(longPath, 30),
		fx:      newEffects(th),
		tx:      newTexts(th),
		lt:      newLights(th),
		sc:      newScenes(th),
		sw:      newSwaps(),
		ph:      newPhysics(),
		dt:      newDatas(),
	}
}

// Init starts the race and the marquee as soon as the program runs.
func (m model) Init() tea.Cmd { return func() tea.Msg { return startMsg{} } }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	if next, cmd, ok := m.familyKey(msg); ok {
		return next.run(cmd)
	}
	switch msg := msg.(type) {
	case startMsg:
		var fxCmd, txCmd, scCmd, phCmd tea.Cmd
		m.marquee, cmd = m.marquee.Start()
		m.fx, fxCmd = m.fx.start()
		m.tx, txCmd = m.tx.start()
		m.sc, scCmd = m.sc.start()
		m.ph, phCmd = m.ph.start()
		m.cv = m.cv.race()
		m.dr = m.dr.again(m.th)
		m.lt = m.lt.diff(m.th)
		m.dt = m.dt.replay()
		return m.replay().run(tea.Batch(cmd, fxCmd, txCmd, scCmd, phCmd))
	case effectMsg:
		m.dr = m.dr.nextEffect(m.th)
		return m.run(nil)
	case patternMsg:
		m.dr = m.dr.nextPattern(m.th)
		return m.run(nil)
	case againMsg:
		m.dr = m.dr.again(m.th)
		return m.run(nil)
	case curvesMsg:
		m.cv = m.cv.race()
		return m.run(nil)
	case turnMsg:
		m.cv = m.cv.next()
		return m.run(nil)
	case scrambleMsg:
		m.fx = m.fx.decode()
		return m.run(nil)
	case odometerMsg:
		m.fx = m.fx.roll()
		return m.run(nil)
	case moveMsg:
		m.fx = m.fx.travel()
		return m.run(nil)
	case replayMsg:
		return m.replay().run(nil)
	case fadeMsg:
		m.fade = animation.NewFade(m.th.Accent.GetForeground(), m.th.Muted.GetForeground(), 1500*time.Millisecond)
		return m.run(nil)
	case typeMsg:
		m.typed = text.NewTypewriter(headline, 35*time.Millisecond)
		return m.run(nil)
	case shakeMsg:
		m.shake = text.NewShake(3, 450*time.Millisecond)
		return m.run(nil)
	case pulseMsg:
		if m.pulse.Active() {
			muted := m.th.Muted.GetForeground()
			m.pulse = animation.NewFade(muted, muted, 0)
			return m, nil
		}
		m.pulse = animation.NewFade(m.th.Accent.GetForeground(), m.th.Muted.GetForeground(), 700*time.Millisecond).
			WithYoyo().WithRepeat(animation.Forever)
		return m.run(nil)
	case toggleMsg:
		m.open = !m.open
		target := 0
		if m.open {
			target = hintsHeight
		}
		// A toggle mid-slide starts from where the pane is; the old tween's
		// frame still on its way is not this one's.
		m.tween = animation.New(m.height, target, slide)
		return m, m.tween.Tick()
	case animation.FrameMsg:
		return m.frame(msg)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "left", "right":
			by := nudge
			if msg.String() == "left" {
				by = -nudge
			}
			target := min(max(m.steady.Target+float64(by), 0), track-1)
			m.steady, _ = m.steady.To(target)
			m.bouncy, _ = m.bouncy.To(target)
			return m.run(nil)
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// replay sets the bars, the smooth bar and the counter off again.
func (m model) replay() model {
	m.bars = make([]animation.Tween, len(eases))
	for i, e := range eases {
		m.bars[i] = animation.New(0, track, race).WithEase(e.ease)
	}
	m.smooth = animation.NewFloat(0, 1, race).WithEase(animation.Linear)
	m.count = animation.New(0, counts, race).WithEase(animation.EaseOut)
	m.cascade = make([]animation.Tween, blocks)
	for i := range m.cascade {
		m.cascade[i] = animation.New(0, block, 300*time.Millisecond).WithEase(animation.EaseOut).WithDelay(time.Duration(i) * stagger)
	}
	return m
}

// run starts the shared clock, if it is not running, beside cmd.
func (m model) run(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	var tick tea.Cmd
	m.clock, tick = m.clock.Start()
	return m, tea.Batch(cmd, tick)
}

// frame hands a frame to whoever owns it: the hints tween and the marquee
// take their own, and the clock's moves everything else to its time.
func (m model) frame(msg animation.FrameMsg) (tea.Model, tea.Cmd) {
	var slideCmd, marqueeCmd, fxCmd, txCmd, scCmd, phCmd, clockCmd tea.Cmd
	if m.tween, slideCmd = m.tween.Update(msg); m.tween.Cur != m.height {
		m.height = m.tween.Cur
		m.sh = m.sh.SetLayout(shape(m.height))
	}
	m.marquee, marqueeCmd = m.marquee.Update(msg)
	m.fx, fxCmd = m.fx.update(msg)
	m.tx, txCmd = m.tx.update(msg)
	m.sc, scCmd = m.sc.update(msg)
	m.ph, phCmd = m.ph.update(msg)
	var at time.Time
	var ok bool
	if m.clock, at, ok = m.clock.Take(msg); ok {
		m = m.advance(at)
		if m.moving() {
			m.clock, clockCmd = m.clock.Next()
		} else {
			m.clock = m.clock.Stop()
		}
	}
	return m, tea.Batch(slideCmd, marqueeCmd, fxCmd, txCmd, scCmd, phCmd, clockCmd)
}

// advance moves every animation on the shared clock to at.
func (m model) advance(at time.Time) model {
	for i := range m.bars {
		m.bars[i], _ = m.bars[i].Advance(at)
	}
	m.steady, _ = m.steady.Advance(at)
	m.bouncy, _ = m.bouncy.Advance(at)
	m.smooth, _ = m.smooth.Advance(at)
	m.count, _ = m.count.Advance(at)
	m.fade, _ = m.fade.Advance(at)
	m.typed, _ = m.typed.Advance(at)
	m.shake, _ = m.shake.Advance(at)
	m.pulse, _ = m.pulse.Advance(at)
	for i := range m.cascade {
		m.cascade[i], _ = m.cascade[i].Advance(at)
	}
	m.fx = m.fx.advance(at)
	m.cv = m.cv.advance(at)
	m.dr = m.dr.advance(at)
	m.tx = m.tx.advance(at)
	m.lt = m.lt.advance(at)
	m.sw = m.sw.advance(at)
	m.ph = m.ph.advance(at)
	m.dt = m.dt.advance(at)
	return m
}

// moving reports whether anything on the shared clock still needs frames.
func (m model) moving() bool {
	for _, b := range append(slices.Clone(m.bars), m.cascade...) {
		if b.Active() {
			return true
		}
	}
	return m.steady.Active() || m.bouncy.Active() || m.smooth.Active() || m.count.Active() ||
		m.fade.Active() || m.typed.Active() || m.shake.Active() || m.pulse.Active() || m.fx.moving() || m.cv.moving() || m.dr.moving() ||
		m.tx.moving() || m.lt.moving() || m.sw.moving() || m.ph.moving() || m.dt.moving()
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"main": {Title: "every animation", Right: still(), Body: func(r layout.Rect) string {
			switch m.sh.ActiveTab() {
			case 1:
				return m.fx.rows(m.th, m.label)
			case 2:
				return m.cv.rows(m.th, m.label)
			case 3:
				return m.dr.rows(m.th, m.label)
			case 4:
				return m.tx.rows(m.th, m.label)
			case 5:
				return m.lt.rows(m.th, m.label)
			case 6:
				return m.sc.rows(m.th, m.label, r.W, r.H)
			case 7:
				return m.sw.rows(m.th, m.label)
			case 8:
				return m.ph.rows(m.th, m.label)
			case 9:
				return m.dt.rows(m.th, m.label)
			}
			return m.gallery()
		}},
		"hints": {Title: "hints", Body: func(layout.Rect) string {
			return "This pane slides on a Tween of its own.\nThe marquee steps on its own ticks too;\neverything else shares one Clock."
		}},
	}))
	v.AltScreen = true
	return v
}

// gallery is one row per animation, its kind and variant on the left.
func (m model) gallery() string {
	label := m.label
	var rows []string
	for i, e := range eases {
		kind := ""
		if i == 0 {
			kind = "easing"
		}
		cur := 0
		if i < len(m.bars) {
			cur = m.bars[i].Cur
		}
		rows = append(rows, label(kind, e.name)+m.bar(cur))
	}
	rows = append(rows, "",
		label("spring", "steady")+m.spring(m.steady),
		label("", "bouncy")+m.spring(m.bouncy),
		"",
		label("float", "sub-cell")+progress.Smooth(int(m.smooth.Cur*1000), 1000, track, m.th),
		label("count-up", "ease out")+m.th.Bold.Render(thousands(m.count.Cur)+" files"),
		label("fade", "accent→muted")+lipgloss.NewStyle().Foreground(m.fade.Color()).Render(saved),
		label("typewriter", "35ms a char")+m.th.Bold.Render(m.typed.View()),
		label("marquee", "30 cells")+m.th.Muted.Render("[")+m.marquee.View()+m.th.Muted.Render("]"),
		label("shake", "3 cells")+m.th.Danger.Render(m.shake.View(password)),
		"",
		label("pulse", "yoyo forever")+lipgloss.NewStyle().Foreground(m.pulse.Color()).Render("● waiting for the server"),
		label("cascade", "120ms apart")+m.blocks(),
	)
	return strings.Join(rows, "\n")
}

// bar is cur of track cells filled.
func (m model) bar(cur int) string {
	cur = min(max(cur, 0), track)
	return m.th.Accent.Render(strings.Repeat("█", cur)) + m.th.Muted.Render(strings.Repeat("░", track-cur))
}

// spring is a track with a dot where s is now, kept on the track when it
// overshoots past either end.
func (m model) spring(s animation.Spring) string {
	at := min(max(s.Int(), 0), track-1)
	return m.th.Muted.Render(strings.Repeat("·", at)) + m.th.Accent.Render("●") + m.th.Muted.Render(strings.Repeat("·", track-1-at))
}

// thousands is n with a comma every three digits.
func thousands(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// blocks is the cascade: each block grown as far as its own tween has come.
func (m model) blocks() string {
	parts := make([]string, blocks)
	for i := range parts {
		cur := 0
		if i < len(m.cascade) {
			cur = m.cascade[i].Cur
		}
		parts[i] = m.th.Accent.Render(strings.Repeat("▮", cur)) + strings.Repeat(" ", block-cur)
	}
	return strings.Join(parts, " ")
}

// still names the switch when motion is off, so a frozen gallery says why.
func still() string {
	if animation.Still() {
		return "motion off"
	}
	return ""
}

// label is a row's kind and variant, in two fixed columns.
func (m model) label(kind, variant string) string {
	return m.th.Accent.Render(fmt.Sprintf("%-11s", kind)) + m.th.Muted.Render(fmt.Sprintf("%-14s", variant))
}

// familyKey acts on the keys of the text, light, scene, swap, motion and data
// tabs, and reports whether msg was one of them.
func (m model) familyKey(msg tea.Msg) (model, tea.Cmd, bool) {
	var cmd tea.Cmd
	switch msg.(type) {
	case decryptMsg:
		m.tx = m.tx.land()
	case tickerMsg:
		m.tx = m.tx.move()
	case markMsg:
		m.tx = m.tx.remark()
	case strikeMsg:
		m.tx = m.tx.cross()
	case copyMsg:
		m.tx = m.tx.copied()
	case lightMsg:
		m.lt = m.lt.next(m.th)
	case shineMsg:
		m.lt = m.lt.again(m.th)
	case diffMsg:
		m.lt = m.lt.diff(m.th)
	case undoMsg:
		m.lt = m.lt.undo(m.th)
	case sceneMsg:
		m.sc, cmd = m.sc.next(m.th)
	case swapMsg:
		m.sw = m.sw.next()
	case againSwap:
		m.sw = m.sw.again()
	case dropMsg:
		m.ph = m.ph.dropAgain()
	case flockMsg:
		m.ph = m.ph.fly()
	case rubberMsg:
		m.ph = m.ph.stretch()
	case growMsg:
		m.dt = m.dt.replay()
	case pushMsg:
		m.dt = m.dt.feed()
	case gaugeMsg:
		m.dt = m.dt.read()
	case warmMsg:
		m.dt = m.dt.warmUp()
	default:
		return m, nil, false
	}
	return m, cmd, true
}
