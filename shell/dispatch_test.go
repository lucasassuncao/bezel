package shell

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/input"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/theme"
)

type ping struct{}

// ChangePane tells the app where focus went, so it can run enter and leave
// effects; nothing is sent when focus stays put.
func TestChangePaneSendsFocusMsg(t *testing.T) {
	s := New(Config{Layout: twoPane(), Theme: theme.Resolve(theme.Plain(), true), Actions: []Action{ChangePane()}})
	s, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	s, r := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.True(t, r.Handled)
	require.Equal(t, "detail", s.Focused())
	require.NotNil(t, r.Cmd)
	require.Equal(t, FocusMsg{From: "list", To: "detail"}, r.Cmd())

	solo := New(Config{Layout: layout.Fill("only"), Theme: theme.Resolve(theme.Plain(), true), Actions: []Action{ChangePane()}})
	solo, _ = solo.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	_, r = solo.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.Nil(t, r.Cmd)
}

var keyP = tea.KeyPressMsg{Code: 'p', Text: "p"}

func TestHelpActionOpensTheHelp(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Help()))
	s, r := s.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
	require.True(t, r.Handled)
	require.IsType(t, overlay.Help{}, s.TopOverlay())
}

func TestCustomActionRunsItsCommand(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Custom("p", "ping", Send(ping{}))))
	_, r := s.Update(keyP)
	require.True(t, r.Handled)
	require.Equal(t, ping{}, r.Cmd())
}

func TestTheContextReachesTheAction(t *testing.T) {
	var got string
	s := sized(New(actionConfig()).SetActions(Custom("p", "ping", func(c ActionContext) tea.Cmd { got = c.Focus; return nil })))
	s.Update(keyP)
	require.Equal(t, s.Focused(), got)
}

func TestMoveNeverConsumesTheArrows(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Move()))
	_, r := s.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.False(t, r.Handled)
}

func TestACustomOnADisplayOnlyKeyStillRuns(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Move(), Custom("up", "page up", Send(ping{}))))
	_, r := s.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	require.True(t, r.Handled)
	require.Equal(t, ping{}, r.Cmd())
}

func TestChangeTabWalksBothWays(t *testing.T) {
	cfg := actionConfig(stubTab{name: "a"}, stubTab{name: "b"}, stubTab{name: "c"})
	cfg.Actions = []Action{ChangeTab()}
	s := sized(New(cfg))
	s, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.Equal(t, 1, s.ActiveTab())
	s, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	require.Equal(t, 0, s.ActiveTab())
	s, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	require.Equal(t, 2, s.ActiveTab())
}

func TestChangePaneMovesFocus(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(ChangePane()))
	before := s.Focused()
	s, r := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.True(t, r.Handled)
	require.NotEqual(t, before, s.Focused())
}

func TestQuitActionQuits(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Quit()))
	_, r := s.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	require.IsType(t, tea.QuitMsg{}, r.Cmd())
}

func TestRunWithReplacesAPrebuiltsWork(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(ChangePane(RunWith(Send(ping{}))), Quit(RunWith(Send(ping{})))))
	before := s.Focused()
	s, r := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.True(t, r.Handled)
	require.Equal(t, before, s.Focused(), "the app moves its own focus")
	require.Equal(t, ping{}, r.Cmd())
	_, r = s.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	require.Equal(t, ping{}, r.Cmd())
}

// RunWith swaps what a prebuilt does, not its text or its place in the legend.
func TestRunWithKeepsThePrebuiltText(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Custom("p", "ping", nil), Quit(RunWith(nil)), Help()))
	require.Contains(t, screen(s), "[p] ping  [q] quit  [?] help")
}

func TestADisplayOnlyCustomLeavesTheKeyToTheApp(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Custom("enter", "open", Send(ping{}), DisplayOnly())))
	_, r := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.False(t, r.Handled)
	require.Contains(t, screen(s), "[enter] open")
}

func TestARefusedActionDoesNotRun(t *testing.T) {
	cfg := actionConfig()
	cfg.Can = func(Capability) bool { return false }
	s := sized(New(cfg).SetActions(Custom("p", "ping", Send(ping{}), Needs("write"))))
	_, r := s.Update(keyP)
	require.False(t, r.Handled)
}

func TestActionsStandAsideForAnOverlay(t *testing.T) {
	ran := false
	s := sized(New(actionConfig()).SetActions(Custom("p", "ping", func(ActionContext) tea.Cmd { ran = true; return nil })))
	s = s.Push(overlay.NewText("t", "x", theme.Resolved{}))
	s.Update(keyP)
	require.False(t, ran)
}

func TestActionsStandAsideForTheInput(t *testing.T) {
	ran := false
	s := sized(New(actionConfig()).SetActions(Custom("p", "ping", func(ActionContext) tea.Cmd { ran = true; return nil })))
	in, _ := input.New(theme.Resolved{}).Focus()
	s, _ = s.SetInput(in).Update(keyP)
	require.False(t, ran)
	got, _ := s.Input()
	require.Equal(t, "p", got.Value())
}
