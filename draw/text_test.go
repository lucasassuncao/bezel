package draw

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
)

// Tail must not pin the whole log's backing array, nor write through to it.
func TestTailIsACopy(t *testing.T) {
	log := []string{"a", "b", "c", "d"}
	tail := Tail(log, 2)
	require.Equal(t, []string{"c", "d"}, tail)
	tail[0] = "x"
	require.Equal(t, "c", log[2])
}

func TestTruncateMeasuresCellsNotBytes(t *testing.T) {
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Render("hello world")
	got := Truncate(styled, 5)
	require.Equal(t, 5, ansi.StringWidth(got))
	require.True(t, strings.HasSuffix(ansi.Strip(got), "…"))
	require.Equal(t, "abc", Truncate("abc", 10))
	require.Equal(t, "abc", Truncate("abc", 0))
}

func TestFitIsExactlyWidth(t *testing.T) {
	require.Equal(t, "ab   ", Fit("ab", 5))
	require.Equal(t, 5, ansi.StringWidth(Fit("abcdefgh", 5)))
	require.Equal(t, "", Fit("abc", 0))
}

func TestBlockSizeAndPadHeight(t *testing.T) {
	w, h := BlockSize("ab\ncdef\ng")
	require.Equal(t, 4, w)
	require.Equal(t, 3, h)
	require.Equal(t, "a\n\n", PadHeight("a", 3))
	require.Equal(t, "a\nb", PadHeight("a\nb", 1))
}

func TestFitBlockCutsRowsAndColumns(t *testing.T) {
	got := FitBlock("aaaa\nbbbb\ncccc", 2, 2)
	lines := strings.Split(got, "\n")
	require.Len(t, lines, 2)
	for _, l := range lines {
		require.Equal(t, 2, ansi.StringWidth(l))
	}
	require.Equal(t, "", FitBlock("x", 3, 0))
	require.Equal(t, "", FitBlock("x", 3, -1))
}

func TestInnerRectRemovesBorderAndPadding(t *testing.T) {
	require.Equal(t, layout.Rect{X: 12, Y: 6, W: 36, H: 8}, InnerRect(layout.Rect{X: 10, Y: 5, W: 40, H: 10}))
	require.Equal(t, layout.Rect{X: 2, Y: 1, W: 1, H: 1}, InnerRect(layout.Rect{X: 0, Y: 0, W: 2, H: 1}))
}

func TestBreadcrumbScrollToAndNumberLines(t *testing.T) {
	require.Equal(t, "a › b › c", Breadcrumb([]string{"a", "b", "c"}))

	ten := "1\n2\n3\n4\n5\n6\n7\n8\n9\n10"
	require.Equal(t, "1\n2\n3", ScrollTo(ten, 3, 0))
	require.Equal(t, "5\n6\n7", ScrollTo(ten, 3, 6))   // centred on the target
	require.Equal(t, "8\n9\n10", ScrollTo(ten, 3, 10)) // clamped at the end
	require.Equal(t, ten, ScrollTo(ten, 20, 5))
	require.Equal(t, "", ScrollTo(ten, 0, 5))

	got := NumberLines("a\nb", lipgloss.NewStyle())
	require.Equal(t, "   1 │ a\n   2 │ b", got)
	require.Equal(t, GutterWidth, ansi.StringWidth(Gutter(1, lipgloss.NewStyle())))
}

func TestWrapBreaksOnWordsAndHyphens(t *testing.T) {
	require.Equal(t, []string{"Instalação", "completa", "do pacote"}, Wrap("Instalação completa do pacote", 10))
	require.Equal(t, []string{"foo-bar-", "baz qux"}, Wrap("foo-bar-baz qux", 8))
}

// Escape codes take no cells, so styled text breaks where the plain text would.
func TestWrapMeasuresStyledTextByWhatIsDrawn(t *testing.T) {
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Render("red text here")
	rows := Wrap(styled, 5)
	for i, r := range rows {
		rows[i] = ansi.Strip(r)
	}
	require.Equal(t, []string{"red", "text", "here"}, rows)
}

func TestWrapCutsWordsWiderThanTheRow(t *testing.T) {
	require.Equal(t, []string{"superlon", "gwordwit", "houtspac", "es here"}, Wrap("superlongwordwithoutspaces here", 8))
	for _, tc := range []struct {
		s string
		w int
	}{
		{"superlongwordwithoutspaces here", 8},
		{"日本語テキスト", 5},
		{"https://example.com/a/very/long/path/to/a/package", 12},
	} {
		for _, row := range Wrap(tc.s, tc.w) {
			require.LessOrEqual(t, ansi.StringWidth(row), tc.w, "%q at %d", row, tc.w)
		}
	}
}

func TestWrapKeepsNewlines(t *testing.T) {
	require.Equal(t, []string{"line one", "line two", "is", "longer"}, Wrap("line one\nline two is longer", 8))
	require.Equal(t, []string{"a", ""}, Wrap("a\n", 5))
}

func TestWrapEdgeWidthsAndInput(t *testing.T) {
	require.Nil(t, Wrap("", 10))
	require.Equal(t, []string{"a long line", "b"}, Wrap("a long line\nb", 0))
	require.Equal(t, []string{"a long line"}, Wrap("a long line", -3))
	require.Equal(t, []string{"   "}, Wrap("   ", 5))
	require.NotPanics(t, func() { Wrap("日本", 1) })
}
