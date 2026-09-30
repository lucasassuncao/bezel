package shell

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/overlay"
	"github.com/lucasassuncao/bezel/palette"
)

var colon = tea.KeyPressMsg{Code: ':', Text: ":"}

// deliver runs cmd and feeds what it returns back through Update, the way
// the bubbletea loop would, batches in order.
func deliver(t *testing.T, s Shell, cmd tea.Cmd) Shell {
	t.Helper()
	if cmd == nil {
		return s
	}
	switch msg := cmd().(type) {
	case nil:
	case tea.BatchMsg:
		for _, c := range msg {
			s = deliver(t, s, c)
		}
	default:
		var next tea.Cmd
		s, _, next = s.Update(msg)
		s = deliver(t, s, next)
	}
	return s
}

func typeLine(s Shell, line string) Shell {
	for _, r := range line {
		s, _, _ = s.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return s
}

func enter(t *testing.T, s Shell) Shell {
	s, _, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	return deliver(t, s, cmd)
}

func TestCommandsOpensThePaletteOnTheStatusRow(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Commands(), Custom("r", "reveal", nil, Named("reveal"))))
	s, handled, _ := s.Update(colon)
	require.True(t, handled)
	require.IsType(t, palette.Model{}, s.TopOverlay())
	screen := ansi.Strip(s.View(map[string]Pane{"list": {}, "detail": {}}))
	require.Contains(t, screen, ":reveal")
	rows := strings.Split(screen, "\n")
	require.True(t, strings.HasPrefix(rows[len(rows)-2], ":"), "the typed line takes the status row")
}

func TestThePaletteRunsTheNamedAction(t *testing.T) {
	got := false
	s := sized(New(actionConfig()).SetActions(Commands(), Custom("x", "reveal", func(ActionContext) tea.Cmd { got = true; return nil }, Named("reveal"))))
	s, _, _ = s.Update(colon)
	s = enter(t, typeLine(s, "rev"))
	require.True(t, got)
	require.False(t, s.HasOverlay())
}

func TestThePalettePassesTheArgument(t *testing.T) {
	var arg string
	s := sized(New(actionConfig()).SetActions(Commands(), Command("goto", "jump", func(c ActionContext) tea.Cmd { arg = c.Arg; return nil }, Arg("path"))))
	s, _, _ = s.Update(colon)
	enter(t, typeLine(s, "goto kv/app"))
	require.Equal(t, "kv/app", arg)
}

func TestThePaletteSaysNotAvailableForARefusedAction(t *testing.T) {
	never := func(Context) bool { return false }
	s := sized(New(actionConfig()).SetActions(Commands(), Custom("e", "edit", nil, Named("edit"), When(never))))
	s, _, _ = s.Update(colon)
	// Only the DoneMsg: delivering everything would also fire the status TTL.
	s, _, cmd := typeLine(s, "edit").Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	s, _, _ = s.Update(cmd())
	require.Equal(t, "edit is not available here", s.Status())
	require.Equal(t, Error, s.StatusLevel())
	require.False(t, s.HasOverlay())
}

// Closing the palette and opening the help must not race: the help stays.
func TestHelpRunFromThePaletteStaysOpen(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Help(), Commands()))
	s, _, _ = s.Update(colon)
	s = enter(t, typeLine(s, "help"))
	require.IsType(t, overlay.Help{}, s.TopOverlay())
	require.Equal(t, 1, strings.Count(ansi.Strip(s.View(map[string]Pane{"list": {}, "detail": {}})), "esc closes"))
}

func TestPaletteOnlyCommandsAreListed(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Commands(), Command("goto", "jump to a path", nil, Arg("path"))))
	s, _, _ = s.Update(colon)
	require.Contains(t, ansi.Strip(s.View(map[string]Pane{"list": {}, "detail": {}})), ":goto <path>")
}

// A second enter queued before the first DoneMsg lands must not run twice.
func TestADoubleEnterRunsTheActionOnce(t *testing.T) {
	runs := 0
	s := sized(New(actionConfig()).SetActions(Commands(), Command("go", "go", func(ActionContext) tea.Cmd { runs++; return nil })))
	s, _, _ = s.Update(colon)
	s = typeLine(s, "go")
	s, _, first := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	s, _, second := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	s = deliver(t, s, first)
	deliver(t, s, second)
	require.Equal(t, 1, runs)
}

// The shell hands a paste to the open palette.
func TestPasteReachesTheOpenPalette(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Commands(), Command("goto", "jump", nil, Arg("path"))))
	s, _, _ = s.Update(colon)
	s = typeLine(s, "goto ")
	s, _, _ = s.Update(tea.PasteMsg{Content: "kv/app"})
	require.True(t, strings.HasPrefix(ansi.Strip(s.TopOverlay().(palette.Model).StatusLine(80)), ":goto kv/app"))
}

func screen(s Shell) string {
	return ansi.Strip(s.View(map[string]Pane{"list": {}, "detail": {}}))
}

func TestOpenCommandsListsTheDeclaredCommands(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Command("goto", "jump to a path", nil, Arg("path"))))
	s = s.OpenCommands()
	require.IsType(t, palette.Model{}, s.TopOverlay())
	require.Contains(t, screen(s), ":goto <path>")
}

// A key the app handles itself: printed in the palette, never bound or listed.
func TestKeyHintShowsInThePaletteOnly(t *testing.T) {
	s := sized(New(actionConfig()).SetActions(Command("reveal", "reveal the secret", Send(ping{}), KeyHint("r"))))
	require.NotContains(t, legendRows(s), "reveal")
	_, handled, _ := s.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	require.False(t, handled)
	require.Regexp(t, `:reveal\s+r\s+reveal the secret`, screen(s.OpenCommands()))
}

func TestSetActionsJoinsATabApp(t *testing.T) {
	s := sized(New(actionConfig(stubTab{name: "a"}, stubTab{name: "b"})).SetActions(Command("goto", "jump", nil)))
	require.Contains(t, screen(s.OpenCommands()), ":goto")
}

// Scopes are judged live: an answer landing under the open palette can bring
// a command into reach, and the list follows on the next keystroke.
func TestThePaletteJudgesScopesLive(t *testing.T) {
	ready := false
	s := sized(New(actionConfig()).SetActions(Command("goto", "jump", nil, When(func(Context) bool { return ready }))))
	s = s.OpenCommands()
	require.NotContains(t, screen(s), ":goto")
	ready = true
	s = typeLine(s, "g")
	require.Contains(t, screen(s), ":goto")
}

// The palette's errors expire on the app's clock, not a fixed one.
func TestPaletteErrorsUseTheConfiguredTTL(t *testing.T) {
	cfg := actionConfig()
	cfg.StatusTTL = time.Millisecond
	s := sized(New(cfg).SetActions(Commands()))
	s, _, _ = s.Update(colon)
	s = typeLine(s, "nope")
	s, _, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	_, _, tick := s.Update(cmd())
	require.NotNil(t, tick)

	got := make(chan tea.Msg, 1)
	go func() { got <- tick() }()
	select {
	case msg := <-got:
		if b, ok := msg.(tea.BatchMsg); ok {
			msg = b[0]()
		}
		require.True(t, IsStatusExpiry(msg))
	case <-time.After(time.Second):
		t.Fatal("the status timer ignored Config.StatusTTL")
	}
}
