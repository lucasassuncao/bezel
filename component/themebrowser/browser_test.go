package themebrowser

import (
	"errors"
	"fmt"
	"github.com/lucasassuncao/bezel/component/table"
	"slices"
	"strings"
	"testing"
	"time"

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

// The cursor row is marked and painted, and keeps its palette's colours.
func TestBrowserModel_MarksTheCursorRow(t *testing.T) {
	m := newBrowserModel(theme.Theme{}, true)
	t.Log("\n" + ansi.Strip(m.View().Content))
	first, ok := m.tbl.Selected()
	if !ok {
		t.Fatal("no row under the cursor")
	}
	row := rowOf(t, m, first.Value)
	if !strings.Contains(ansi.Strip(row), "▶ ") {
		t.Errorf("the cursor row has no marker: %q", ansi.Strip(row))
	}
	if !strings.Contains(row, m.th.Cursor.Render(first.Value)) {
		t.Errorf("the cursor row's name is not in the cursor's style: %q", row)
	}
	if !strings.Contains(row, sgr(theme.ResolveColors(theme.All()[first.Value], true).Accent)) {
		t.Errorf("the cursor row lost its palette: %q", row)
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
	return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
}

// On 80 columns each row shows six blocks, in the colours that tell one theme
// from another: accent, selection, border, dim, success, danger.
func TestBrowserModel_ShowsEachThemesPalette(t *testing.T) {
	m := newBrowserModel(theme.Theme{}, true)
	row := rowOf(t, m, "gruvbox")
	if n := strings.Count(ansi.Strip(row), "██"); n != 6 {
		t.Fatalf("want six blocks in the gruvbox row, got %d: %q", n, ansi.Strip(row))
	}
	at := 0
	for _, hex := range []string{"#FE8019", "#FABD2F", "#7C6F64", "#A89984", "#B8BB26", "#FB4934"} {
		i := strings.Index(row[at:], sgr(hex))
		if i < 0 {
			t.Fatalf("gruvbox's %s is missing or out of order: %q", hex, row)
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

// sized is a browser after the terminal reported w by h.
func sized(t *testing.T, w, h int) *browserModel {
	t.Helper()
	m := newBrowserModel(theme.Theme{}, true)
	m.th.CaretBlink = time.Millisecond // each key arms a blink the test waits on
	m.filter = m.newFilter()
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(*browserModel)
}

// press sends one key and runs the command it returns, feeding its message
// back, the way the program would.
func press(m *browserModel, k tea.KeyPressMsg) *browserModel {
	next, cmd := m.Update(k)
	m = next.(*browserModel)
	if cmd != nil {
		if msg := cmd(); msg != nil {
			next, _ = m.Update(msg)
			m = next.(*browserModel)
		}
	}
	return m
}

func typed(m *browserModel, s string) *browserModel {
	for _, r := range s {
		m = press(m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

// The panel is as wide as its widest row and a margin, never the terminal's
// whole width, and shows at most 20 themes.
func TestBrowserModel_FitsItsRows(t *testing.T) {
	m := sized(t, 200, 60)
	top := strings.Split(m.View().Content, "\n")[0]
	if got, want := ansi.StringWidth(top), m.needs(m.layout)+4+margin; got >= 200 || got != want {
		t.Errorf("the panel is %d wide, want %d: its widest row and a margin", got, want)
	}
	if n := m.tbl.Scroll().Visible; n != maxRows {
		t.Errorf("%d themes show, want %d", n, maxRows)
	}
}

// A short terminal shows fewer rows, leaving a line for the prompt.
func TestBrowserModel_FitsAShortTerminal(t *testing.T) {
	lines := strings.Split(sized(t, 120, 20).View().Content, "\n")
	if len(lines) != 19 {
		t.Errorf("the view is %d lines, want 19 on a 20-line terminal", len(lines))
	}
}

// A wide terminal names all ten colours in full; a narrow one abbreviates.
func TestBrowserModel_NamesEachColour(t *testing.T) {
	wide := ansi.Strip(sized(t, 140, 30).View().Content)
	for _, name := range []string{"accent", "select", "border", "dim", "text", "success", "warning", "danger", "info", "bg"} {
		if !strings.Contains(wide, name) {
			t.Errorf("a wide terminal does not name %q:\n%s", name, wide)
		}
	}
	if n := strings.Count(ansi.Strip(rowOf(t, sized(t, 140, 30), "gruvbox")), "██"); n != 10 {
		t.Errorf("want ten blocks on a wide terminal, got %d", n)
	}
	narrow := ansi.Strip(sized(t, 80, 30).View().Content)
	if strings.Contains(narrow, "select") || !strings.Contains(narrow, "sel") {
		t.Errorf("an 80-column terminal should abbreviate the colours:\n%s", narrow)
	}
}

// The keys sit under the panel, quit included.
func TestBrowserModel_ShowsItsKeys(t *testing.T) {
	view := ansi.Strip(sized(t, 100, 30).View().Content)
	for _, want := range []string{"move", "filter", "copy name", "quit"} {
		if !strings.Contains(view, want) {
			t.Errorf("the key line lacks %q:\n%s", want, view)
		}
	}
}

// enter copies the name under the cursor and says so.
func TestBrowserModel_EnterCopiesTheName(t *testing.T) {
	m := sized(t, 100, 30)
	var got string
	m.write = func(s string) error { got = s; return nil }
	m = press(m, tea.KeyPressMsg{Code: tea.KeyDown})
	m = press(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if got != "plain" {
		t.Errorf("copied %q, want plain", got)
	}
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, `copied "plain"`) {
		t.Errorf("the copy is not reported:\n%s", view)
	}
}

// A clipboard that fails is reported, not swallowed.
func TestBrowserModel_ReportsAFailedCopy(t *testing.T) {
	m := sized(t, 100, 30)
	m.write = func(string) error { return errors.New("no clipboard") }
	m = press(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "no clipboard") {
		t.Errorf("the failure is not reported:\n%s", view)
	}
}

// / filters by name, q is text while typing, enter copies the best match and
// esc brings every theme back.
func TestBrowserModel_FiltersByName(t *testing.T) {
	m := sized(t, 100, 30)
	var got string
	m.write = func(s string) error { got = s; return nil }
	m = typed(press(m, tea.KeyPressMsg{Code: '/', Text: "/"}), "grap")
	if m.tbl.Len() == 0 || m.tbl.Len() == len(theme.All()) {
		t.Fatalf("the filter kept %d of %d themes", m.tbl.Len(), len(theme.All()))
	}
	if !strings.Contains(selValue(m.tbl), "grap") {
		t.Errorf("best match = %q, want a name with grap", selValue(m.tbl))
	}
	m = press(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if got != selValue(m.tbl) {
		t.Errorf("copied %q, want %q", got, selValue(m.tbl))
	}
	m = press(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.tbl.Len() != len(theme.All()) {
		t.Errorf("esc left %d themes, want all %d", m.tbl.Len(), len(theme.All()))
	}
}

// A resize that changes the colour columns keeps the cursor on its theme.
func TestBrowserModel_ResizeKeepsTheCursor(t *testing.T) {
	m := sized(t, 140, 30)
	for range 5 {
		m = press(m, tea.KeyPressMsg{Code: tea.KeyDown})
	}
	want := selValue(m.tbl)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	if got := selValue(next.(*browserModel).tbl); got != want {
		t.Errorf("after the resize the cursor is on %q, want %q", got, want)
	}
}

// The colour columns share one width, so the blocks sit evenly spaced.
func TestBrowserModel_SpacesTheColoursEvenly(t *testing.T) {
	for _, w := range []int{80, 200} {
		row := []rune(ansi.Strip(rowOf(t, sized(t, w, 30), "gruvbox")))
		var starts []int
		for i, r := range row {
			if r == '█' && (i == 0 || row[i-1] != '█') && i < len(row)-1 && row[i+1] == '█' {
				starts = append(starts, i)
			}
		}
		for i := 2; i < len(starts); i++ {
			if starts[i]-starts[i-1] != starts[1]-starts[0] {
				t.Errorf("at %d columns the blocks start at %v, not evenly", w, starts)
				break
			}
		}
	}
}

// selValue is the theme name under the table's cursor, "" when none is.
func selValue(t table.Model[string]) string {
	it, _ := t.Selected()
	return it.Value
}
