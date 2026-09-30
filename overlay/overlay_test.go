package overlay

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
)

// fake records what it saw and renders a fixed box.
type fake struct {
	seen    []string
	ownsEsc bool
	closeOn string
}

func (f *fake) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
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
