package overlay

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// fake records what it saw and renders a fixed box.
type fake struct {
	seen    []string
	got     []tea.Msg
	ownsEsc bool
	closeOn string
}

func (f *fake) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		f.got = append(f.got, msg)
	}
	if ok {
		f.seen = append(f.seen, k.String())
		if k.String() == f.closeOn {
			return f, Close()
		}
	}
	return f, nil
}
func (f *fake) View(layout.Rect) string { return "XX\nXX" }
func (f *fake) Legend() []legend.Entry  { return []legend.Entry{legend.New("x", "fake")} }
func (f *fake) OwnsEsc() bool           { return f.ownsEsc }

func press(k string) tea.KeyPressMsg { return tea.KeyPressMsg{Code: rune(k[0]), Text: k} }
func esc() tea.KeyPressMsg           { return tea.KeyPressMsg{Code: tea.KeyEscape} }

// drain runs a tea.BatchMsg (or a single msg) and returns what it produced.
func drain(m tea.Msg) []tea.Msg {
	batch, ok := m.(tea.BatchMsg)
	if !ok {
		return []tea.Msg{m}
	}
	var out []tea.Msg
	for _, c := range batch {
		if c != nil {
			out = append(out, c())
		}
	}
	return out
}

func TestOnlyTheTopOverlayGetsKeys(t *testing.T) {
	a, b := &fake{}, &fake{}
	s := Stack{}.Push(a).Push(b)
	_, handled, _ := s.Update(press("q"))
	require.True(t, handled)
	require.Equal(t, []string{"q"}, b.seen)
	require.Empty(t, a.seen)
}

func TestEscPopsUnlessTheOverlayOwnsIt(t *testing.T) {
	s := Stack{}.Push(&fake{})
	s, handled, cmd := s.Update(esc())
	require.True(t, handled)
	require.IsType(t, CloseMsg{}, cmd(), "esc closes through CloseMsg, so the app hears it")
	s, _, _ = s.Update(cmd())
	require.Equal(t, 0, s.Len())

	owner := &fake{ownsEsc: true}
	s = Stack{}.Push(owner)
	s, _, _ = s.Update(esc())
	require.Equal(t, 1, s.Len())
	require.Equal(t, []string{"esc"}, owner.seen)
}

func TestCloseAndPushMessagesMoveTheStack(t *testing.T) {
	s := Stack{}.Push(&fake{})
	s, handled, _ := s.Update(CloseMsg{})
	require.True(t, handled)
	require.Equal(t, 0, s.Len())

	s, _, _ = s.Update(PushMsg{Overlay: &fake{}})
	require.Equal(t, 1, s.Len())
}

func TestEmptyStackHandlesNothing(t *testing.T) {
	s, handled, cmd := Stack{}.Update(press("q"))
	require.False(t, handled)
	require.Nil(t, cmd)
	require.Equal(t, 0, s.Len())
	require.Nil(t, s.Top())
}

func TestViewCompositesEveryLayerCentred(t *testing.T) {
	bg := strings.Repeat("..........\n", 4) + ".........."
	s := Stack{}.Push(&fake{})
	got := s.View(layout.Rect{W: 10, H: 5}, bg)
	lines := strings.Split(got, "\n")
	require.Equal(t, "....XX....", lines[1])
	require.Equal(t, "....XX....", lines[2])
	require.Equal(t, "..........", lines[0])
	require.Equal(t, bg, Stack{}.View(layout.Rect{W: 10, H: 5}, bg))
}

type bottom struct{ fake }

func (bottom) AnchorBottom() bool { return true }

func TestBottomAnchoredOverlaysSitOnTheBodyFloor(t *testing.T) {
	bg := strings.Repeat("..........\n", 4) + ".........."
	got := Stack{}.Push(&bottom{}).View(layout.Rect{W: 10, H: 5}, bg)
	lines := strings.Split(got, "\n")
	require.Equal(t, "XX........", lines[3])
	require.Equal(t, "XX........", lines[4])
	require.Equal(t, "..........", lines[2])
}

type answerMsg struct{}

// An overlay's own command answers through TopMsg, which the stack hands to
// the top overlay instead of leaving it to the app.
func TestTopMsgReachesTheTopOverlay(t *testing.T) {
	under, top := &fake{}, &fake{}
	s := Stack{}.Push(under).Push(top)

	_, handled, _ := s.Update(ToTop(func() tea.Msg { return answerMsg{} })())
	require.True(t, handled)
	require.Equal(t, []tea.Msg{answerMsg{}}, top.got, "the payload arrives unwrapped")
	require.Empty(t, under.got)
}

// A late answer for a modal already closed has nobody to go to.
func TestTopMsgWithoutAnOverlayIsDropped(t *testing.T) {
	_, handled, cmd := Stack{}.Update(TopMsg{Msg: answerMsg{}})
	require.True(t, handled)
	require.Nil(t, cmd)
}

// The stack sizes what it holds, on a push and on every new body, so Update
// knows what fits before anything has been drawn.
func TestStackSizesItsOverlays(t *testing.T) {
	pgdown := tea.KeyPressMsg{Code: tea.KeyPgDown}
	pager := NewPager("Out", numbered(60), theme.Resolved{})

	s := Stack{}.Resize(layout.Rect{W: 60, H: 18}).Push(pager)
	s, _, _ = s.Update(pgdown)
	require.Equal(t, 10, s.Top().(Pager).Offset(), "pushed onto a sized stack")

	s = s.Push(NewText("t", nil, theme.Resolved{})).Resize(layout.Rect{W: 60, H: 13}).Pop()
	s, _, _ = s.Update(pgdown)
	require.Equal(t, 15, s.Top().(Pager).Offset(), "a layer under the top was resized too")

	s = Stack{}.Push(pager).Resize(layout.Rect{W: 60, H: 18})
	s, _, _ = s.Update(pgdown)
	require.Equal(t, 10, s.Top().(Pager).Offset(), "pushed before the stack had a body")
}

func TestToTopKeepsNothingAsNothing(t *testing.T) {
	require.Nil(t, ToTop(nil))
	require.Nil(t, ToTop(func() tea.Msg { return nil })())
}

func hint(o Overlay) string { return ansi.Strip(legend.HintLine(o.Legend(), legend.Style{})) }

// What each dialog lists for its keys, as the help overlay prints it.
func TestDialogLegendsNameTheirKeys(t *testing.T) {
	th := theme.Resolved{}
	for want, o := range map[string]Overlay{
		"[←/→] choose · [enter] pick · [y/n] answer · [esc] cancel": NewConfirm("t", "m", nil, th),
		"[↑/↓] choose · [enter] pick · [esc] cancel":                NewPick("t", []string{"a"}, nil, th),
		"[esc] close": NewHelp(nil, th),
		"[↑/↓] scroll · [←/→] pan · [pgup/pgdn] page · [esc] close": NewPager("t", "x", th),
		"[enter] confirm · [esc] cancel":                            NewPrompt("t", nil, th),
	} {
		require.Equal(t, want, hint(o))
	}
}

func bind(keys, help string) key.Binding {
	return key.NewBinding(key.WithKeys(strings.Split(keys, "/")...), key.WithHelp(keys, help))
}

func TestPushingNilLeavesTheStackAlone(t *testing.T) {
	s := Stack{}.Push(nil)
	require.Equal(t, 0, s.Len())
	require.NotPanics(t, func() { _ = s.View(layout.Rect{W: 40, H: 10}, "base") })
}

var clickBody = layout.Rect{X: 0, Y: 2, W: 60, H: 20}

// cellOf is the screen cell where text first shows in the stack's view.
func cellOf(t *testing.T, s Stack, text string) (x, y int) {
	t.Helper()
	blank := strings.Repeat(strings.Repeat(" ", clickBody.W)+"\n", clickBody.Y+clickBody.H)
	for y, row := range strings.Split(ansi.Strip(s.View(clickBody, blank)), "\n") {
		if at := strings.Index(row, text); at >= 0 {
			return ansi.StringWidth(row[:at]), y
		}
	}
	t.Fatalf("%q is not on screen", text)
	return 0, 0
}

// isClose reports whether cmd, or one in its batch, closes the dialog.
func isClose(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	switch msg := cmd().(type) {
	case CloseMsg:
		return true
	case tea.BatchMsg:
		for _, c := range msg {
			if isClose(c) {
				return true
			}
		}
	}
	return false
}

func TestAClickReachesTheTopDialogInItsOwnCells(t *testing.T) {
	s := Stack{}.Push(NewAlert(Info, "Saved", "all done", theme.Resolved{}))
	x, y := cellOf(t, s, "OK")
	_, cmd := s.Click(x, y, clickBody)
	require.True(t, isClose(cmd), "a click on OK")
	_, cmd = s.Click(0, 0, clickBody)
	require.Nil(t, cmd, "a click off the dialog does nothing")
	_, cmd = Stack{}.Push(&fake{}).Click(x, y, clickBody)
	require.Nil(t, cmd, "a dialog that is no Clicker ignores it")
}

func TestButtonAtFindsEachButtonOnItsRow(t *testing.T) {
	frame := lipgloss.NewStyle().PaddingLeft(1)
	view := " title\n\n  [ A ]  [ A ] "
	buttons := []string{"[ A ]", "[ A ]"}
	require.Equal(t, 0, buttonAt(view, " ", frame, buttons, 2, 2))
	require.Equal(t, 1, buttonAt(view, " ", frame, buttons, 9, 2), "found by place, not by label")
	require.Equal(t, -1, buttonAt(view, " ", frame, buttons, 7, 2), "the gap")
	require.Equal(t, -1, buttonAt(view, " ", frame, buttons, 2, 0), "another row")
	require.Equal(t, -1, buttonAt(view, " ", frame, buttons, 1, 2), "the lead")
}
