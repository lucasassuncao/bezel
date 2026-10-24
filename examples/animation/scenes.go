package main

import (
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/animation/scene"
	"github.com/lucasassuncao/bezel/theme"
)

type sceneMsg struct{} // n on the scene tab: the next scene

// sceneKinds are the scenes the tab plays, one at a time.
var sceneKinds = []struct {
	name string
	make func(th theme.Resolved) scene.Scene
}{
	{"rain", func(th theme.Resolved) scene.Scene {
		return scene.NewRain(th.Success.GetForeground(), animation.Blend(black, th.Success.GetForeground(), 0.6))
	}},
	{"snow", func(th theme.Resolved) scene.Scene { return scene.NewSnow(th.Bold.GetForeground()) }},
	{"confetti", func(th theme.Resolved) scene.Scene { return scene.NewConfetti(seriesColors(th)...) }},
	{"fireworks", func(th theme.Resolved) scene.Scene { return scene.NewFireworks(seriesColors(th)...) }},
	{"starfield", func(th theme.Resolved) scene.Scene { return scene.NewStarfield(th.Bold.GetForeground()) }},
	{"static", func(th theme.Resolved) scene.Scene { return scene.NewStatic(th.Bold.GetForeground()) }},
	{"plasma", func(th theme.Resolved) scene.Scene {
		return scene.NewPlasma(black, animation.Blend(black, th.Accent.GetForeground(), 0.6))
	}},
}

// seriesColors are the theme's series colours, for the scenes that want many.
func seriesColors(th theme.Resolved) []color.Color {
	out := make([]color.Color, len(th.Series))
	for i, s := range th.Series {
		out[i] = s.GetForeground()
	}
	return out
}

// scenes is the seventh tab: a scene filling the pane, a title over it.
type scenes struct {
	kind  int
	scene scene.Scene
}

func newScenes(th theme.Resolved) scenes { return scenes{scene: sceneKinds[0].make(th)} }

// start sets the scene going on its own ticks.
func (s scenes) start() (scenes, tea.Cmd) { return s, s.scene.Tick() }

// next swaps in the next scene; the old one's frames stop being its.
func (s scenes) next(th theme.Resolved) (scenes, tea.Cmd) {
	s.kind = (s.kind + 1) % len(sceneKinds)
	s.scene = sceneKinds[s.kind].make(th)
	return s, s.scene.Tick()
}

func (s scenes) update(msg tea.Msg) (scenes, tea.Cmd) {
	var cmd tea.Cmd
	s.scene, cmd = s.scene.Update(msg)
	return s, cmd
}

// rows is a line naming the scene, then the scene w×h with a title over it.
func (s scenes) rows(th theme.Resolved, label func(kind, variant string) string, w, h int) string {
	head := label("scene", sceneKinds[s.kind].name) + th.Muted.Render("n next")
	if w <= 0 || h <= 1 {
		return head
	}
	badge := lipgloss.NewStyle().Background(black).Foreground(th.Bold.GetForeground()).Padding(0, 2).
		Render("bezel ✦ " + sceneKinds[s.kind].name)
	over := lipgloss.Place(w, h-1, lipgloss.Center, lipgloss.Center, badge)
	return head + "\n" + s.scene.Under(over)
}
