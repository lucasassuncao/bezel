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
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

func border(l look) any { return l.style().GetBorderTopForeground() }

func TestKindPicksTheBorderColour(t *testing.T) {
	th := theme.Resolve(theme.Mint(), true)
	require.Equal(t, th.Modal, look{}.withTheme(th).style(), "a dialog with no Kind is the theme's own modal")
	for _, k := range []draw.Kind{draw.Info, draw.Success, draw.Warning, draw.Danger} {
		require.Equal(t, k.Color(), border(look{}.withTheme(th).withKind(k)), "the Kind's fixed colour, not the theme's")
	}
}

func TestWithKindTintsEveryDialog(t *testing.T) {
	th := theme.Resolve(theme.Mint(), true)
	for name, l := range map[string]look{
		"alert":   themed(NewAlert(draw.Danger, "t", "m"), th).look,
		"confirm": themed(NewConfirm("t", "m", nil), th).WithKind(draw.Danger).look,
		"text":    themed(NewText("t", ""), th).WithKind(draw.Danger).look,
		"choice":  themed(NewChoice("t", "", nil), th).WithKind(draw.Danger).look,
		"pick":    themed(NewPick("t", nil, nil), th).WithKind(draw.Danger).look,
		"help":    themed(NewHelp(nil), th).WithKind(draw.Danger).look,
		"pager":   themed(NewPager("t", "x"), th).WithKind(draw.Danger).look,
		"prompt":  themed(NewPrompt("t", nil), th).WithKind(draw.Danger).look,
		"picker":  themed(NewFilePicker("t", t.TempDir(), nil), th).WithKind(draw.Danger).look,
	} {
		require.Equal(t, draw.Danger.Color(), border(l), name)
	}
}

// The picker draws its entries in the dialog's colour, so a new kind repaints them.
func TestFilePickerWithKindRepaintsEntries(t *testing.T) {
	th := theme.Resolve(theme.Mint(), true)
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "sub"), 0o750))
	p := themed(NewFilePicker("t", dir, nil), th).WithKind(draw.Danger)
	o, _ := p.Update(p.Init()())
	danger := lipgloss.NewStyle().Foreground(draw.Danger.Color())
	require.Contains(t, o.View(layout.Rect{W: 80, H: 30}), danger.Render(">"))
}

// The caret's blink comes back bare, at the theme's pace: a host that drives
// the prompt hands it on, and the shell routes it to the dialog that armed it.
func TestPromptCaretBlinksAtTheThemesPace(t *testing.T) {
	th := theme.Resolved{CaretBlink: time.Nanosecond}
	p, r := themed(NewPrompt("t", nil), th).Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	cmd := r.Cmd
	blink := cmd()
	require.IsType(t, input.BlinkMsg{}, blink)
	_, next := p.Update(blink)
	require.NotNil(t, next, "the prompt's own blink re-arms")
}

// Every dialog but Alert takes the edit, and a Kind still tints the border.
func TestWithStylesReachesEveryNeutralDialog(t *testing.T) {
	th := theme.Resolve(theme.Mint(), true)
	box := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	edit := func(s Styles) Styles { s.Box = box; return s }
	looks := map[string]look{
		"confirm": themed(NewConfirm("t", "m", nil), th).WithStyles(edit).look,
		"text":    themed(NewText("t", ""), th).WithStyles(edit).look,
		"choice":  themed(NewChoice("t", "", nil), th).WithStyles(edit).look,
		"pick":    themed(NewPick("t", nil, nil), th).WithStyles(edit).look,
		"help":    themed(NewHelp(nil), th).WithStyles(edit).look,
		"pager":   themed(NewPager("t", ""), th).WithStyles(edit).look,
		"prompt":  themed(NewPrompt("t", nil), th).WithStyles(edit).look,
		"files":   themed(NewFilePicker("t", ".", nil), th).WithStyles(edit).look,
	}
	for name, l := range looks {
		require.Equal(t, lipgloss.NormalBorder(), l.style().GetBorderStyle(), name)
		l = l.withKind(draw.Danger)
		require.Equal(t, draw.Danger.Color(), l.style().GetBorderTopForeground(), name+": the kind keeps its colour")
	}
}

// Every bezel dialog repaints under a new theme, so a stack can hand it on.
var _ = []Themer{Alert{}, Confirm{}, Text{}, Choice{}, Pick{}, Help{}, Pager{}, Prompt{}, FilePicker{}}

func TestStackSetThemeRepaintsTheDialogs(t *testing.T) {
	plain, mint := theme.Resolve(theme.Plain(), true), theme.Resolve(theme.Mint(), true)
	body := layout.Rect{W: 60, H: 20}
	s := Stack{}.Push(themed(NewText("t", "body"), plain)).SetTheme(mint)
	require.Equal(t, themed(NewText("t", "body"), mint).View(body), s.Top().View(body))
}

// Every dialog, Alert included, draws the theme's modal box once themed.
func TestEveryDialogDrawsTheModalBorder(t *testing.T) {
	th := theme.Resolve(theme.Mint(), true)
	body := layout.Rect{W: 60, H: 20}
	for name, o := range map[string]Overlay{
		"alert":   themed(NewAlert(draw.Success, "t", "m"), th),
		"confirm": themed(NewConfirm("t", "m", nil), th),
		"text":    themed(NewText("t", "m"), th),
		"choice":  themed(NewChoice("t", "m", nil), th),
		"pick":    themed(NewPick("t", []string{"a"}, nil), th),
		"help":    themed(NewHelp(nil), th),
		"pager":   themed(NewPager("t", "m"), th),
		"prompt":  themed(NewPrompt("t", nil), th),
	} {
		require.Contains(t, ansi.Strip(o.View(body)), "╭", name)
	}
}
