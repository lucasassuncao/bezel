package overlay

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

func border(l look) any { return l.style().GetBorderTopForeground() }

func TestKindPicksTheBorderColour(t *testing.T) {
	th := theme.Resolve(theme.Mint, true)
	require.Equal(t, th.Modal, look{th: th}.style(), "Info is the theme's own modal")
	require.Equal(t, lipgloss.Color(th.Colors.Danger), border(look{th: th, kind: Danger}))
	require.Equal(t, lipgloss.Color(th.Colors.Success), border(look{th: th, kind: Success}))
	require.Equal(t, lipgloss.Color(th.Colors.Warning), border(look{th: th, kind: Warning}))
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
	p := NewFilePicker("t", t.TempDir(), nil, th).WithKind(Danger)
	require.Equal(t, lipgloss.Color(th.Colors.Danger), p.picker.Styles.Cursor.GetForeground())
}

func TestPromptCaretFollowsTheTheme(t *testing.T) {
	th := theme.Resolved{CaretBlink: 7}
	require.EqualValues(t, 7, NewPrompt("t", nil, th).input.Styles().Cursor.BlinkSpeed)
}
