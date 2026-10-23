// Animation: every kind the animation package has, side by side. The timed
// ones share one Clock; the marquee and the sliding hints pane keep their own.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/progress"
	"github.com/lucasassuncao/bezel/shell"
	"github.com/lucasassuncao/bezel/theme"
)

// hintsHeight is the open pane, borders included.
const hintsHeight = 8

const (
	slide  = 300 * time.Millisecond
	race   = 1500 * time.Millisecond
	track  = 32 // cells in a bar or a spring's track
	nudge  = 8  // cells an arrow moves the springs' target
	counts = 12345
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
)

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
	typed          animation.Typewriter
	shake          animation.Shake
	marquee        animation.Marquee

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
		Layout: shape(0),
		Theme:  th,
		Title:  "animation",
		Actions: []shell.Action{
			shell.Custom("r", "replay", shell.Send(replayMsg{})),
			shell.Custom("←/→", "spring", nil, shell.DisplayOnly()),
			shell.Custom("f", "fade", shell.Send(fadeMsg{})),
			shell.Custom("t", "type", shell.Send(typeMsg{})),
			shell.Custom("s", "shake", shell.Send(shakeMsg{})),
			shell.Custom("h", "hints", shell.Send(toggleMsg{})),
			shell.Quit(),
		},
	})
	muted := th.Muted.GetForeground()
	return model{
		sh: sh, th: th, clock: animation.NewClock(),
		steady: animation.NewSpring(0), bouncy: animation.NewSpring(0).WithDamping(0.3),
		fade:    animation.NewFade(muted, muted, 0),
		typed:   animation.NewTypewriter(headline, 0),
		shake:   animation.NewShake(3, 0),
		marquee: animation.NewMarquee(longPath, 30),
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
	switch msg := msg.(type) {
	case startMsg:
		m.marquee, cmd = m.marquee.Start()
		m = m.replay()
		return m.run(cmd)
	case replayMsg:
		return m.replay().run(nil)
	case fadeMsg:
		m.fade = animation.NewFade(m.th.Accent.GetForeground(), m.th.Muted.GetForeground(), 1500*time.Millisecond)
		return m.run(nil)
	case typeMsg:
		m.typed = animation.NewTypewriter(headline, 35*time.Millisecond)
		return m.run(nil)
	case shakeMsg:
		m.shake = animation.NewShake(3, 450*time.Millisecond)
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
	var slideCmd, marqueeCmd, clockCmd tea.Cmd
	if m.tween, slideCmd = m.tween.Update(msg); m.tween.Cur != m.height {
		m.height = m.tween.Cur
		m.sh = m.sh.SetLayout(shape(m.height))
	}
	m.marquee, marqueeCmd = m.marquee.Update(msg)
	if at, ok := m.clock.Frame(msg); ok {
		m = m.advance(at)
		if m.moving() {
			m.clock, clockCmd = m.clock.Next()
		} else {
			m.clock = m.clock.Stop()
		}
	}
	return m, tea.Batch(slideCmd, marqueeCmd, clockCmd)
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
	return m
}

// moving reports whether anything on the shared clock still needs frames.
func (m model) moving() bool {
	for _, b := range m.bars {
		if b.Active() {
			return true
		}
	}
	return m.steady.Active() || m.bouncy.Active() || m.smooth.Active() || m.count.Active() ||
		m.fade.Active() || m.typed.Active() || m.shake.Active()
}

func (m model) View() tea.View {
	v := tea.NewView(m.sh.View(map[string]shell.Pane{
		"main": {Title: "every animation", Body: func(layout.Rect) string { return m.gallery() }},
		"hints": {Title: "hints", Body: func(layout.Rect) string {
			return "This pane slides on a Tween of its own.\nThe marquee steps on its own ticks too;\neverything else shares one Clock."
		}},
	}))
	v.AltScreen = true
	return v
}

// gallery is one row per animation, its kind and variant on the left.
func (m model) gallery() string {
	label := func(kind, variant string) string {
		return m.th.Accent.Render(fmt.Sprintf("%-11s", kind)) + m.th.Muted.Render(fmt.Sprintf("%-13s", variant))
	}
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
