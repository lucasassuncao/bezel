package overlay

import (
	"github.com/charmbracelet/x/ansi"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/input"
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
		"text":    NewText("t", "", th).WithKind(Danger).look,
		"choice":  NewChoice("t", "", nil, th).WithKind(Danger).look,
		"pick":    NewPick("t", nil, nil, th).WithKind(Danger).look,
		"help":    NewHelp(nil, th).WithKind(Danger).look,
		"pager":   NewPager("t", "x", th).WithKind(Danger).look,
		"prompt":  NewPrompt("t", nil, th).WithKind(Danger).look,
		"picker":  NewFilePicker("t", t.TempDir(), nil, th).WithKind(Danger).look,
	} {
		require.Equal(t, lipgloss.Color(th.Colors.Danger), border(l), name)
	}
}

// The picker draws its entries in the dialog's colour, so a new kind repaints them.
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
	p, r := NewPrompt("t", nil, th).Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	cmd := r.Cmd
	blink := cmd()
	require.IsType(t, input.BlinkMsg{}, blink)
	_, next := p.Update(blink)
	require.NotNil(t, next, "the prompt's own blink re-arms")
}

// Every dialog but Alert takes the edit, and a Kind still tints the border.
func TestWithStylesReachesEveryNeutralDialog(t *testing.T) {
	th := theme.Resolve(theme.Mint, true)
	box := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	edit := func(s Styles) Styles { s.Box = box; return s }
	looks := map[string]look{
		"confirm": NewConfirm("t", "m", nil, th).WithStyles(edit).look,
		"text":    NewText("t", "", th).WithStyles(edit).look,
		"choice":  NewChoice("t", "", nil, th).WithStyles(edit).look,
		"pick":    NewPick("t", nil, nil, th).WithStyles(edit).look,
		"help":    NewHelp(nil, th).WithStyles(edit).look,
		"pager":   NewPager("t", "", th).WithStyles(edit).look,
		"prompt":  NewPrompt("t", nil, th).WithStyles(edit).look,
		"files":   NewFilePicker("t", ".", nil, th).WithStyles(edit).look,
	}
	for name, l := range looks {
		require.Equal(t, lipgloss.NormalBorder(), l.style().GetBorderStyle(), name)
		l.kind = Danger
		require.Equal(t, lipgloss.Color(th.Colors.Danger), l.style().GetBorderTopForeground(), name+": the kind keeps its colour")
	}
}

// Every bezel dialog repaints under a new theme, so a stack can hand it on.
var _ = []Themer{Alert{}, Confirm{}, Text{}, Choice{}, Pick{}, Help{}, Pager{}, Prompt{}, FilePicker{}}

func TestStackSetThemeRepaintsTheDialogs(t *testing.T) {
	plain, mint := theme.Resolve(theme.Plain, true), theme.Resolve(theme.Mint, true)
	body := layout.Rect{W: 60, H: 20}
	s := Stack{}.Push(NewText("t", "body", plain)).SetTheme(mint)
	require.Equal(t, NewText("t", "body", mint).View(body), s.Top().View(body))
}

// Every dialog draws the theme's modal box, Alert included: one built
// without lookOf had no styles and so no border.
func TestEveryDialogDrawsTheModalBorder(t *testing.T) {
	th := theme.Resolve(theme.Mint, true)
	body := layout.Rect{W: 60, H: 20}
	for name, o := range map[string]Overlay{
		"alert":   NewAlert(Success, "t", "m", th),
		"confirm": NewConfirm("t", "m", nil, th),
		"text":    NewText("t", "m", th),
		"choice":  NewChoice("t", "m", nil, th),
		"pick":    NewPick("t", []string{"a"}, nil, th),
		"help":    NewHelp(nil, th),
		"pager":   NewPager("t", "m", th),
		"prompt":  NewPrompt("t", nil, th),
	} {
		require.Contains(t, ansi.Strip(o.View(body)), "╭", name)
	}
}
