package shell

import (
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/theme"
)

type ping struct{}

// ChangePane tells the app where focus went, so it can run enter and leave
// effects; nothing is sent when focus stays put.
func TestChangePaneSendsFocusMsg(t *testing.T) {
	s := New(Config{Layout: twoPane(), Theme: theme.Resolve(theme.ThemePlain, true), Actions: []Action{ChangePane()}})
	s, _, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	s, handled, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.True(t, handled)
	require.Equal(t, "detail", s.Focus())
	require.NotNil(t, cmd)
	require.Equal(t, FocusMsg{From: "list", To: "detail"}, cmd())

	solo := New(Config{Layout: layout.Fill("only"), Theme: theme.Resolve(theme.ThemePlain, true), Actions: []Action{ChangePane()}})
	solo, _, _ = solo.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	_, _, cmd = solo.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.Nil(t, cmd)
}

var keyP = tea.KeyPressMsg{Code: 'p', Text: "p"}

func TestHelpActionOpensTheHelp(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Help()))
	s, handled, _ := s.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
	require.True(t, handled)
	require.IsType(t, overlay.Help{}, s.TopOverlay())
}

func TestCustomActionRunsItsCommand(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Custom("p", "ping", Send(ping{}))))
	_, handled, cmd := s.Update(keyP)
	require.True(t, handled)
	require.Equal(t, ping{}, cmd())
}

func TestTheContextReachesTheAction(t *testing.T) {
	var got string
	s := sized(New(actionConfig()).SetActions(Custom("p", "ping", func(c ActionContext) tea.Cmd { got = c.Focus; return nil })))
	s.Update(keyP)
	require.Equal(t, s.Focus(), got)
}

func TestMoveNeverConsumesTheArrows(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Move()))
	_, handled, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.False(t, handled)
}

func TestACustomOnADisplayOnlyKeyStillRuns(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Move(), Custom("up", "page up", Send(ping{}))))
	_, handled, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	require.True(t, handled)
	require.Equal(t, ping{}, cmd())
}

func TestChangeTabWalksBothWays(t *testing.T) {
	cfg := actionConfig(stubTab{name: "a"}, stubTab{name: "b"}, stubTab{name: "c"})
	cfg.Actions = []Action{ChangeTab()}
	s := sized(New(cfg))
	s, _, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.Equal(t, 1, s.ActiveTab())
	s, _, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	require.Equal(t, 0, s.ActiveTab())
	s, _, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	require.Equal(t, 2, s.ActiveTab())
}

func TestChangePaneMovesFocus(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(ChangePane()))
	before := s.Focus()
	s, handled, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.True(t, handled)
	require.NotEqual(t, before, s.Focus())
}

func TestQuitActionQuits(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Quit()))
	_, _, cmd := s.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	require.IsType(t, tea.QuitMsg{}, cmd())
}

func TestRunWithReplacesAPrebuiltsWork(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(ChangePane(RunWith(Send(ping{}))), Quit(RunWith(Send(ping{})))))
	before := s.Focus()
	s, handled, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.True(t, handled)
	require.Equal(t, before, s.Focus(), "the app moves its own focus")
	require.Equal(t, ping{}, cmd())
	_, _, cmd = s.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	require.Equal(t, ping{}, cmd())
}

// RunWith swaps what a prebuilt does, not its text or its place in the legend.
func TestRunWithKeepsThePrebuiltText(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Custom("p", "ping", nil), Quit(RunWith(nil)), Help()))
	require.Contains(t, screen(s), "[p] ping  [q] quit  [?] help")
}

func TestADisplayOnlyCustomLeavesTheKeyToTheApp(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Custom("enter", "open", Send(ping{}), DisplayOnly())))
	_, handled, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.False(t, handled)
	require.Contains(t, screen(s), "[enter] open")
}

func TestARefusedActionDoesNotRun(t *testing.T) {
	cfg := actionConfig()
	cfg.Can = func(Capability) bool { return false }
	s := sized(New(cfg).SetActions(Custom("p", "ping", Send(ping{}), Needs("write"))))
	_, handled, _ := s.Update(keyP)
	require.False(t, handled)
}

func TestActionsStandAsideForAnOverlay(t *testing.T) {
	ran := false
	s := sized(New(actionConfig()).SetActions(Custom("p", "ping", func(ActionContext) tea.Cmd { ran = true; return nil })))
	s = s.Push(overlay.NewText("t", []string{"x"}, lipgloss.NewStyle(), legend.Style{}))
	s.Update(keyP)
	require.False(t, ran)
}

func TestActionsStandAsideForTheInput(t *testing.T) {
	ran := false
	s := sized(New(actionConfig()).SetActions(Custom("p", "ping", func(ActionContext) tea.Cmd { ran = true; return nil })))
	in := textinput.New()
	in.Focus()
	s = s.SetInput(&in)
	s.Update(keyP)
	require.False(t, ran)
	require.Equal(t, "p", in.Value())
}
