package shell

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/theme"
)

// requireLaidOut fails unless s holds what its state resolves to now: the
// rects for the current body, and focus on a leaf that was placed.
func requireLaidOut(t *testing.T, s Shell, op string) {
	t.Helper()
	present := func(name string) bool { return s.present == nil || s.present[name] }
	want, collapsed := layout.Resolve(s.cfg.Layout, s.Body(), present)
	require.Equal(t, want, s.rects, op)
	require.Equal(t, collapsed, s.collapsed, op)
	_, placed := s.rects[s.Focus()]
	require.True(t, placed, "%s: focus on %q, which is not placed", op, s.Focus())
}

func send(msg tea.Msg) func(Shell) Shell {
	return func(s Shell) Shell {
		s, _, _ = s.Update(msg)
		return s
	}
}

// The placement is cached on the shell. Every way to change the shell must
// leave it current, or the app reads rects for a body that is gone.
func TestEveryChangeLeavesTheLayoutResolved(t *testing.T) {
	cfg := actionConfig(stubTab{name: "a"}, stubTab{name: "b"})
	cfg.Actions = []Action{Help(), Commands(), ChangeTab(), Custom("ctrl+p", "pane", nil)}
	// The header takes a second row off the list, so focus alone moves the body.
	cfg.Header = func(ctx Context, _ int) []string {
		if ctx.Focus == "list" {
			return []string{"one"}
		}
		return []string{"one", "two"}
	}
	in := textinput.New()
	ops := []struct {
		name string
		do   func(Shell) Shell
	}{
		{"SetFocus", func(s Shell) Shell { return s.SetFocus("detail") }},
		{"SetTab", func(s Shell) Shell { return s.SetTab(1) }},
		{"SetTabs", func(s Shell) Shell { return s.SetTabs(stubTab{name: "a"}) }},
		{"SetLayout", func(s Shell) Shell { return s.SetLayout(layout.Rows(layout.Fill("detail"), layout.Fill("list"))) }},
		{"SetTheme", func(s Shell) Shell { return s.SetTheme(theme.Resolve(theme.Mint, false)) }},
		{"SetSubtitle", func(s Shell) Shell { return s.SetSubtitle("sub") }},
		{"SetActions", func(s Shell) Shell { return s.SetActions(Custom("x", "x", nil)) }},
		{"SetLegendLines", func(s Shell) Shell { return s.SetLegendLines(3) }},
		{"ShowHelp", func(s Shell) Shell { return s.ShowHelp() }},
		{"OpenCommands", func(s Shell) Shell { return s.OpenCommands() }},
		{"Push", func(s Shell) Shell { return s.Push(overlay.NewText("t", []string{"x"}, theme.Resolved{})) }},
		{"Pop", func(s Shell) Shell { return s.ShowHelp().Pop() }},
		{"SetInput", func(s Shell) Shell { return s.SetInput(in) }},
		{"ClearInput", func(s Shell) Shell { return s.SetInput(in).ClearInput() }},
		{"SetStatus", func(s Shell) Shell { s, _ = s.SetStatus("hi", OK, time.Second); return s }},
		{"ClearStatus", func(s Shell) Shell { return s.ClearStatus() }},
		{"Busy", func(s Shell) Shell { s, _ = s.Busy("working"); return s }},
		{"Idle", func(s Shell) Shell { s, _ = s.Busy("working"); return s.Idle() }},
		{"resize", send(tea.WindowSizeMsg{Width: 60, Height: 20})},
		{"tab key", send(bezeltest.Key("tab"))},
		{"help key", send(bezeltest.Key("?"))},
		{"esc on help", func(s Shell) Shell { return send(overlay.CloseMsg{})(s.ShowHelp()) }},
		{"focus message", send(FocusMsg{From: "list", To: "detail"})},
	}
	base := sized(New(cfg))
	for _, from := range []Shell{base, base.SetFocus("detail")} {
		requireLaidOut(t, from, "start")
		for _, op := range ops {
			requireLaidOut(t, op.do(from), op.name)
		}
	}
}
