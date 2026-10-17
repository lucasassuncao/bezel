package overlay

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

func hint(o Overlay) string { return ansi.Strip(legend.HintLine(o.Legend(), legend.Style{})) }

// What each dialog lists for its keys, as the help overlay prints it.
func TestDialogLegendsNameTheirKeys(t *testing.T) {
	th := theme.Resolved{}
	for want, o := range map[string]Overlay{
		"[←/→] choose · [enter] pick · [y/n] answer · [esc] cancel": NewConfirm("t", "m", nil, th),
		"[↑/↓] choose · [enter] pick · [esc] cancel":                NewPick("t", []string{"a"}, nil, th),
		"[esc] close": NewHelp(nil, th),
		"[↑/↓] scroll · [pgup/pgdn] page · [esc] close": NewPager("t", "x", th),
		"[enter] confirm · [esc] cancel":                NewPrompt("t", nil, th),
	} {
		require.Equal(t, want, hint(o))
	}
}

func bind(keys, help string) key.Binding {
	return key.NewBinding(key.WithKeys(strings.Split(keys, "/")...), key.WithHelp(keys, help))
}

func TestConfirmAnswersToTheKeysItWasGiven(t *testing.T) {
	keys := DefaultConfirmKeys()
	keys.Yes = bind("s", "sim")
	c := NewConfirm("Apagar?", "", func() tea.Msg { return nil }, theme.Resolved{}).WithKeys(keys)

	_, cmd := c.Update(press("y"))
	require.Nil(t, cmd, "y is no longer an answer")
	_, cmd = c.Update(press("s"))
	require.Contains(t, drain(cmd()), tea.Msg(CloseMsg{}))
	require.Contains(t, hint(c), "[s] sim")
}

func TestPickAnswersToTheKeysItWasGiven(t *testing.T) {
	keys := DefaultPickKeys()
	keys.Up, keys.Down = bind("k", "move"), key.NewBinding(key.WithKeys("j"))
	p := NewPick("t", []string{"a", "b"}, nil, theme.Resolved{}).WithKeys(keys)

	o, _ := p.Update(press("j"))
	require.Equal(t, 1, o.(Pick).Cursor())
	o, _ = o.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	require.Equal(t, 1, o.(Pick).Cursor(), "the arrow is no longer bound")
	require.Contains(t, ansi.Strip(p.View(layout.Rect{W: 60, H: 20})), "[k] move")
}

func TestHelpClosesOnTheKeysItWasGiven(t *testing.T) {
	keys := DefaultHelpKeys()
	keys.Close = bind("x", "close")
	h := NewHelp(nil, theme.Resolved{}).WithKeys(keys)

	_, cmd := h.Update(press("q"))
	require.Nil(t, cmd)
	_, cmd = h.Update(press("x"))
	require.Equal(t, tea.Msg(CloseMsg{}), cmd())
	require.Contains(t, ansi.Strip(h.View(layout.Rect{W: 60, H: 20})), "x closes")
}

func TestPagerScrollsOnTheKeysItWasGiven(t *testing.T) {
	keys := DefaultPagerKeys()
	keys.Down = key.NewBinding(key.WithKeys("n"))
	p := NewPager("t", "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12", theme.Resolved{}).WithKeys(keys)
	o := p.Resize(layout.Rect{W: 60, H: 12})

	o, _ = o.Update(press("j"))
	require.Equal(t, 0, o.(Pager).Offset(), "j is no longer bound")
	o, _ = o.Update(press("n"))
	require.Equal(t, 1, o.(Pager).Offset())
}

func TestPromptAcceptsOnTheKeyItWasGiven(t *testing.T) {
	keys := DefaultPromptKeys()
	keys.Accept = bind("ctrl+s", "save")
	p := newPrompt().WithKeys(keys)

	_, cmd := p.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Nil(t, cmd)
	_, cmd = p.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	require.Contains(t, drain(cmd()), tea.Msg(CloseMsg{}))
	require.Contains(t, hint(p), "[ctrl+s] confirm", "the key from the binding, what it does from the hint")
}
