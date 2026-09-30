package textbox

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/theme"
)

func th() theme.Resolved { return theme.Resolve(theme.ThemePlain, true) }

// The widget stops at 99 lines and 400 characters by default, silently.
func TestALongValueIsKeptWhole(t *testing.T) {
	ta := New(th())
	long := strings.Repeat("0123456789\n", 150) + strings.Repeat("x", 1000)
	SetText(&ta, long)
	require.Equal(t, long, ta.Value())
}

// The widget breaks a line on \r and on \n alike, so CRLF would double every line.
func TestSetTextFoldsCRLF(t *testing.T) {
	ta := New(th())
	SetText(&ta, "a\r\nb\r\nc")
	require.Equal(t, "a\nb\nc", ta.Value())
	require.Equal(t, 3, ta.LineCount())
}

func TestRowsCarryTheSharedGutter(t *testing.T) {
	ta := New(th())
	ta.SetWidth(40)
	ta.SetHeight(3)
	SetText(&ta, "first")
	out := ansi.Strip(ta.View())
	require.Contains(t, out, ansi.Strip(draw.Gutter(1, th().Muted))+"first")
	require.Contains(t, out, "~", "rows past the end are marked")
}

// A paste from Windows arrives with CRLF; through Update it keeps one line per line.
func TestUpdateFoldsCRLFInAPaste(t *testing.T) {
	ta := New(th())
	ta.Focus()
	ta, _ = Update(ta, tea.PasteMsg{Content: "a\r\nb\r\nc"})
	require.Equal(t, "a\nb\nc", ta.Value())
}

// Height is the caller's: a panel sizes it from its layout, a form from its content.
func TestHeightIsLeftToTheCaller(t *testing.T) {
	ta := New(th())
	ta.SetHeight(7)
	SetText(&ta, strings.Repeat("line\n", 20))
	require.Equal(t, 7, ta.Height())
}
