package overlay

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/input"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

func border(l look) any { return l.style().GetBorderTopForeground() }

func TestKindPicksTheBorderColour(t *testing.T) {
	th := theme.Resolve(theme.Mint, true)
	require.Equal(t, th.Modal, lookOf(th).style(), "Info is the theme's own modal")
	require.Equal(t, lipgloss.Color(th.Colors.Danger), border(look{th: &th, kind: Danger}))
	require.Equal(t, lipgloss.Color(th.Colors.Success), border(look{th: &th, kind: Success}))
	require.Equal(t, lipgloss.Color(th.Colors.Warning), border(look{th: &th, kind: Warning}))
}

func TestWithKindTintsEveryDialog(t *testing.T) {
	th := theme.Resolve(theme.Mint, true)
	for name, l := range map[string]look{
		"alert":   NewAlert(Danger, "t", "m", th).look,
		"confirm": NewConfirm("t", "m", nil, th).WithKind(Danger).look,
		"text":    NewText("t", nil, th).WithKind(Danger).look,
		"choice":  NewChoice("t", nil, nil, th).WithKind(Danger).look,
		"pick":    NewPick("t", nil, nil, th).WithKind(Danger).look,
		"help":    NewHelp(nil, th).WithKind(Danger).look,
		"pager":   NewPager("t", "x", th).WithKind(Danger).look,
		"prompt":  NewPrompt("t", nil, th).WithKind(Danger).look,
		"picker":  NewFilePicker("t", t.TempDir(), nil, th).WithKind(Danger).look,
	} {
		require.Equal(t, lipgloss.Color(th.Colors.Danger), border(l), name)
	}
}

// The picker's entries are drawn by bubbles, so a new kind must reach them too.
func TestFilePickerWithKindRepaintsEntries(t *testing.T) {
	th := theme.Resolve(theme.Mint, true)
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "sub"), 0o750))
	p := NewFilePicker("t", dir, nil, th).WithKind(Danger)
	o, _ := p.Update(p.Init()())
	danger := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Colors.Danger))
	require.Contains(t, o.View(layout.Rect{W: 80, H: 30}), danger.Render(">"))
}

// The caret's blink comes back bare, at the theme's pace: a host that drives
// the prompt hands it on, and the shell routes it to the dialog that armed it.
func TestPromptCaretBlinksAtTheThemesPace(t *testing.T) {
	th := theme.Resolved{CaretBlink: time.Nanosecond}
	p, cmd := NewPrompt("t", nil, th).Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	blink := cmd()
	require.IsType(t, input.BlinkMsg{}, blink)
	_, next := p.Update(blink)
	require.NotNil(t, next, "the prompt's own blink re-arms")
}
