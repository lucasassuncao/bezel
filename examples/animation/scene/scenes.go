package main

import (
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel"
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
		return scene.Rain(th.Success.GetForeground(), animation.Blend(black, th.Success.GetForeground(), 0.6))
	}},
	{"snow", func(th theme.Resolved) scene.Scene { return scene.Snow(th.Bold.GetForeground()) }},
	{"confetti", func(th theme.Resolved) scene.Scene { return scene.Confetti(seriesColors(th)...) }},
	{"fireworks", func(th theme.Resolved) scene.Scene { return scene.Fireworks(seriesColors(th)...) }},
	{"starfield", func(th theme.Resolved) scene.Scene { return scene.Starfield(th.Bold.GetForeground()) }},
	{"static", func(th theme.Resolved) scene.Scene { return scene.Static(th.Bold.GetForeground()) }},
	{"plasma", func(th theme.Resolved) scene.Scene {
		return scene.Plasma(black, animation.Blend(black, th.Accent.GetForeground(), 0.6))
	}},
	{"fire", func(th theme.Resolved) scene.Scene {
		return scene.Fire(th.Warning.GetForeground(), th.Danger.GetForeground())
	}},
	{"aurora", func(th theme.Resolved) scene.Scene { return scene.Aurora(seriesColors(th)[:3]...) }},
	{"bubbles", func(th theme.Resolved) scene.Scene { return scene.Bubbles(th.Info.GetForeground()) }},
	{"sea", func(th theme.Resolved) scene.Scene {
		return scene.Sea(th.Bold.GetForeground(), animation.Blend(black, th.Info.GetForeground(), 0.5))
	}},
	{"life", func(th theme.Resolved) scene.Scene { return scene.Life(th.Success.GetForeground()) }},
}

// seriesColors are the theme's series colours, for the scenes that want many.
func seriesColors(th theme.Resolved) []color.Color {
	out := make([]color.Color, len(th.Series))
	for i, s := range th.Series {
		out[i] = s.GetForeground()
	}
	return out
}

// scenes is the gallery's one screen: a scene filling the pane, a title over it.
type scenes struct {
	kind  int
	scene scene.Scene
}

func newScenes(th theme.Resolved) scenes { return scenes{scene: sceneKinds[0].make(th)} }

// next swaps in the next scene; the old one's frames stop being its.
func (s scenes) next(th theme.Resolved) (scenes, tea.Cmd) {
	s.kind = (s.kind + 1) % len(sceneKinds)
	s.scene = sceneKinds[s.kind].make(th)
	return s, s.scene.Tick()
}

func (s scenes) update(msg tea.Msg) (scenes, tea.Cmd) {
	var r bezel.Result
	s.scene, r = s.scene.Update(msg)
	return s, r.Cmd
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
