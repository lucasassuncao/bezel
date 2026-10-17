package themebrowser

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/theme"
)

// Without a terminal the names come out one category per line, sorted, so a
// script can read them.
func TestListGroupsTheNamesByCategory(t *testing.T) {
	var b strings.Builder
	if err := List(&b); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if strings.Count(out, "\n") != len(theme.Categories()) {
		t.Fatalf("want one line per category:\n%s", out)
	}
	for _, c := range theme.Categories() {
		if !strings.Contains(out, c.Name) {
			t.Errorf("category %q is missing:\n%s", c.Name, out)
		}
	}
	if !strings.Contains(out, "grape") {
		t.Errorf("a theme name is missing:\n%s", out)
	}
}

func TestBrowserModel_RendersAndNavigates(t *testing.T) {
	m := newBrowserModel(theme.Theme{}, true)

	view := ansi.Strip(m.View().Content)
	for _, want := range []string{"Theme", "Category", "plain", "Miscellaneous"} {
		if !strings.Contains(view, want) {
			t.Errorf("view should contain %q", want)
		}
	}

	// down moves the cursor.
	before := m.tbl.Cursor()
	updated, _ := m.Update(tea.KeyPressMsg{Text: "down", Code: tea.KeyDown})
	m = updated.(*browserModel)
	if m.tbl.Cursor() != before+1 {
		t.Errorf("cursor = %d, want %d after pressing down", m.tbl.Cursor(), before+1)
	}

	// up moves it back.
	updated, _ = m.Update(tea.KeyPressMsg{Text: "up", Code: tea.KeyUp})
	m = updated.(*browserModel)
	if m.tbl.Cursor() != before {
		t.Errorf("cursor = %d, want %d after pressing up", m.tbl.Cursor(), before)
	}
}

// TestBrowserModel_NotFullScreen guards the actual bug report: the view must
// not set tea.View.AltScreen (which clears/takes over the whole terminal
// just to show a small table).
func TestBrowserModel_NotFullScreen(t *testing.T) {
	m := newBrowserModel(theme.Theme{}, true)
	if v := m.View(); v.AltScreen {
		t.Error("View().AltScreen = true, want false - this should render inline, not take over the screen")
	}
}

// TestBrowserModel_NoOverflowTruncation guards that the border sizes to the
// table's content, never leaving a stray "…" from a line truncated against a
// manually computed width.
func TestBrowserModel_NoOverflowTruncation(t *testing.T) {
	m := newBrowserModel(theme.Theme{}, true)
	view := ansi.Strip(m.View().Content)
	if strings.Contains(view, "…") {
		t.Error("view contains a stray ellipsis - a line was truncated against a mismatched box width")
	}
}

// TestBrowserModel_SelectedRowHighlightSpansFullRow guards that the cursor
// row is one unbroken highlight. It used to stop after the first column
// because Cell's own Foreground reset cut off Selected's Background.
func TestBrowserModel_SelectedRowHighlightSpansFullRow(t *testing.T) {
	m := newBrowserModel(theme.Theme{}, true)
	lines := strings.Split(m.View().Content, "\n")

	var selectedLine string
	for _, l := range lines {
		if strings.Contains(ansi.Strip(l), "plain") {
			selectedLine = l
			break
		}
	}
	if selectedLine == "" {
		t.Fatal("could not find the row containing the initially-selected theme (plain)")
	}

	// No reset may sit between the first and last column's text: a mid-row
	// "[m" means the highlight broke at a column boundary.
	start := strings.Index(selectedLine, "plain")
	end := strings.Index(selectedLine, "Miscellaneous")
	if start < 0 || end < 0 || end < start {
		t.Fatalf("could not locate both column values in the selected row: %q", selectedLine)
	}
	if between := selectedLine[start:end]; strings.Contains(between, "\x1b[m") {
		t.Errorf("selected row resets its style between columns - the highlight does not span the full row:\n%q", between)
	}
}

func TestBrowserModel_QuitKey(t *testing.T) {
	m := newBrowserModel(theme.Theme{}, true)
	_, cmd := m.Update(tea.KeyPressMsg{Text: "q", Code: 'q'})
	if cmd == nil {
		t.Fatal("expected a Cmd for the quit key")
	}
	if msg := cmd(); msg != tea.Quit() {
		t.Errorf("quit key should produce tea.Quit, got %#v", msg)
	}
}

// esc only goes back, and the browser has nowhere to go back to.
func TestBrowserModel_EscDoesNotQuit(t *testing.T) {
	m := newBrowserModel(theme.Theme{}, true)
	if _, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape}); cmd != nil {
		t.Errorf("esc should not quit, got %#v", cmd())
	}
}

// A key release must not move the cursor a second time.
func TestBrowserModel_IgnoresKeyRelease(t *testing.T) {
	m := newBrowserModel(theme.Theme{}, true)
	before := m.tbl.Cursor()
	updated, _ := m.Update(tea.KeyReleaseMsg{Text: "down", Code: tea.KeyDown})
	if got := updated.(*browserModel).tbl.Cursor(); got != before {
		t.Errorf("cursor = %d, want %d after a key release", got, before)
	}
}

// rowOf is the rendered line of the table row whose name column is name.
func rowOf(t *testing.T, m *browserModel, name string) string {
	t.Helper()
	for _, l := range strings.Split(m.View().Content, "\n") {
		if slices.Contains(strings.Fields(ansi.Strip(l)), name) {
			return l
		}
	}
	t.Fatalf("no row for %q", name)
	return ""
}

// sgr is the truecolor foreground code lipgloss writes for a "#RRGGBB" colour.
func sgr(hex string) string {
	var r, g, b int
	if _, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("38;2;%d;%d;%dm", r, g, b)
}

// Each row shows six blocks, in the colours that tell one theme from another,
// in the order accent, selection, border, dim, success, danger.
func TestBrowserModel_ShowsEachThemesPalette(t *testing.T) {
	m := newBrowserModel(theme.Theme{}, true)
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "Palette") {
		t.Fatalf("the Palette column is missing:\n%s", view)
	}
	row := rowOf(t, m, "mint")
	if !strings.Contains(ansi.Strip(row), "██ ██ ██ ██ ██ ██") {
		t.Fatalf("want six blocks in the mint row: %q", ansi.Strip(row))
	}
	at := 0
	for _, hex := range []string{"#3EB489", "#98DFAF", "#4A7B6F", "#2E4F46", "#2ECC71", "#E74C3C"} {
		i := strings.Index(row[at:], sgr(hex))
		if i < 0 {
			t.Fatalf("mint's %s is missing or out of order: %q", hex, row)
		}
		at += i
	}
}

// The adaptive default's blocks follow the background the terminal reported.
func TestBrowserModel_PaletteFollowsTheBackground(t *testing.T) {
	dark := theme.ResolveColors(theme.Theme{}, true).Accent
	light := theme.ResolveColors(theme.Theme{}, false).Accent
	row := rowOf(t, newBrowserModel(theme.Theme{}, false), "default")
	if !strings.Contains(row, sgr(light)) || strings.Contains(row, sgr(dark)) {
		t.Errorf("on a light terminal, default should show the light accent %s: %q", light, row)
	}
}
