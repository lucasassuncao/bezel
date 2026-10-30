package overlay

import (
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

func TestAlertClosesOnAnyKey(t *testing.T) {
	a := NewAlert(draw.Info, "Saved", "written to disk", theme.Resolved{})
	_, r := a.Update(press("x"))
	cmd := r.Cmd
	require.NotNil(t, cmd)
	require.IsType(t, CloseMsg{}, cmd())
	require.Contains(t, ansi.Strip(a.View(layout.Rect{W: 60, H: 20})), "written to disk")
}

// Every ready-made dialog, whatever its text, draws within the body it is
// shown over, once the body has room for a dialog: 30 by 14 cells.
func FuzzDialogsStayInTheBody(f *testing.F) {
	f.Add("Delete", "This removes the file for good.", 40, 12)
	f.Add("全角の題名", "本文\n二行目", 20, 8)
	f.Add("", "", 12, 6)
	f.Fuzz(func(t *testing.T, title, text string, w, h int) {
		if !utf8.ValidString(title+text) || strings.ContainsFunc(title+text, func(r rune) bool { return unicode.IsControl(r) && r != '\n' }) {
			t.Skip()
		}
		w, h = 30+(w%100+100)%100, 14+(h%40+40)%40
		th := theme.Resolve(theme.Plain(), true)
		body := layout.Rect{W: w, H: h}
		for name, o := range map[string]Overlay{
			"alert":   NewAlert(draw.Info, title, text, th),
			"confirm": NewConfirm(title, text, nil, th),
			"text":    NewText(title, text, th),
			"choice":  NewChoice(title, text, []Option{Opt("y", "yes", nil), Opt("n", "no", nil)}, th),
			"pick":    NewPick(title, []string{text, title, "third"}, nil, th),
			"pager":   NewPager(title, text, th),
			"prompt":  NewPrompt(title, nil, th),
		} {
			if r, ok := o.(Resizer); ok {
				o = r.Resize(body)
			}
			bw, bh := draw.BlockSize(o.View(body))
			require.LessOrEqual(t, bw, w, "%s is wider than the body", name)
			require.LessOrEqual(t, bh, h, "%s is taller than the body", name)
		}
	})
}
