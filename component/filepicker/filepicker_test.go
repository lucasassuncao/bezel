package filepicker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Plain(), true)

// tree lays out dir/{notes.txt, main.go, sub/inner.md} and returns dir.
func tree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hi"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "sub"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "inner.md"), []byte("# x"), 0o600))
	return dir
}

// load runs cmd and hands its message back to m.
func load(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	require.NotNil(t, cmd)
	m, _ = m.Update(cmd())
	return m
}

func press(m Model, code rune) (Model, tea.Cmd) {
	return withCmd(m.Update(tea.KeyPressMsg{Code: code}))
}

func rows(m Model) []string {
	var out []string
	for _, l := range strings.Split(ansi.Strip(m.View(th)), "\n") {
		out = append(out, strings.TrimRight(l, " "))
	}
	return out
}

func TestTheListingPutsDirectoriesFirst(t *testing.T) {
	m := New(tree(t))
	require.Empty(t, m.View(th), "nothing until the listing comes")
	m = load(t, m, m.Init())
	require.Equal(t, []string{">        sub", "     12B main.go", "      2B notes.txt"}, rows(m))
}

func TestEnterPicksAFileAndOpensADirectory(t *testing.T) {
	dir := tree(t)
	m := New(dir)
	m = load(t, m, m.Init())
	m, cmd := press(m, tea.KeyEnter)
	m = load(t, m, cmd)
	require.Equal(t, filepath.Join(dir, "sub"), m.Dir())
	m, cmd = press(m, tea.KeyEnter)
	require.Nil(t, cmd)
	path, ok := m.Picked()
	require.True(t, ok)
	require.Equal(t, filepath.Join(dir, "sub", "inner.md"), path)
	m, _ = press(m, tea.KeyDown)
	_, ok = m.Picked()
	require.False(t, ok, "only the Update that picked says so")
}

func TestAFileOfTheWrongTypeIsRefused(t *testing.T) {
	m := New(tree(t)).WithAllowedTypes(".go")
	m = load(t, m, m.Init())
	m, _ = press(m, tea.KeyDown)
	m, _ = press(m, tea.KeyDown)
	m, _ = press(m, tea.KeyEnter)
	_, ok := m.Picked()
	require.False(t, ok)
	require.EqualError(t, m.Err(), "notes.txt is not .go")
}

func TestTheColumnsComeAndGo(t *testing.T) {
	m := New(tree(t)).WithSizes(false).WithPermissions(true)
	m = load(t, m, m.Init())
	r := rows(m)
	require.True(t, strings.HasPrefix(r[0], "> d"), r[0])
	require.Contains(t, r[2], "-rw")
	require.True(t, strings.HasSuffix(r[2], " notes.txt"))
	require.NotContains(t, r[2], "2B")
}

func TestAnotherPickersListingIsNotTaken(t *testing.T) {
	dir := tree(t)
	a, b := New(dir), New(dir)
	a = load(t, a, b.Init())
	require.Empty(t, a.View(th))
}

func TestBackFromARootListsTheDrives(t *testing.T) {
	root := filepath.VolumeName(t.TempDir()) + string(filepath.Separator)
	m := New(root).WithVolumes(func() []string { return []string{root, "Z:" + string(filepath.Separator)} })
	m, cmd := press(m, tea.KeyLeft)
	m = load(t, m, cmd)
	require.False(t, m.Browsing())
	require.Equal(t, root, m.Highlighted())
	require.Contains(t, ansi.Strip(m.View(th)), "> "+root)
}

func TestSetHeightScrollsTheListing(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"a", "b", "c", "d", "e"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, n), nil, 0o600))
	}
	m := New(dir).SetHeight(3)
	m = load(t, m, m.Init())
	require.Contains(t, ansi.Strip(m.View(th)), "↓ 3 more")
	require.Equal(t, 3, m.Height())
}

func TestHumanBytesReadsLikeAListing(t *testing.T) {
	for n, want := range map[int64]string{
		0: "0B", 999: "999B", 1000: "1.0kB", 1500: "1.5kB", 12_345: "12kB",
		640_000_000: "640MB", 3_200_000_000: "3.2GB",
		999_950: "1.0MB", 999_999_999: "1.0GB", 9_960: "10kB", 1_000_000_000_000: "1.0TB",
	} {
		require.Equal(t, want, humanBytes(n), "%d", n)
	}
}

func TestItsOwnListingAndItsKeysAreHandled(t *testing.T) {
	m := New(tree(t))
	m, r := m.Update(m.Init()())
	require.True(t, r.Handled, "its own listing")
	_, r = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.True(t, r.Handled)
	_, r = m.Update(tea.KeyPressMsg{Code: tea.KeyF5})
	require.False(t, r.Handled)
	other := New(tree(t))
	_, r = m.Update(other.Init()())
	require.False(t, r.Handled, "another picker's listing")
}

// withCmd keeps the command of an Update, for the tests that run it.
func withCmd[T any](m T, r bezel.Result) (T, tea.Cmd) { return m, r.Cmd }

// many lays out dir/f00.txt to f19.txt and returns dir.
func many(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for i := range 20 {
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("f%02d.txt", i)), nil, 0o600))
	}
	return dir
}

func TestPagesAndEndsMoveTheCursor(t *testing.T) {
	m := New(many(t)).SetHeight(6)
	m = load(t, m, m.Init())
	for _, step := range []struct {
		key  rune
		want string
	}{
		{tea.KeyEnd, "f19.txt"}, {tea.KeyHome, "f00.txt"}, {tea.KeyPgDown, "f05.txt"},
		{tea.KeyDown, "f06.txt"}, {tea.KeyPgUp, "f01.txt"}, {tea.KeyUp, "f00.txt"},
	} {
		m, _ = press(m, step.key)
		require.Equal(t, step.want, filepath.Base(m.Highlighted()))
	}
}

func TestVimKeysMoveWithLetters(t *testing.T) {
	dir := tree(t)
	m := New(filepath.Join(dir, "sub")).WithKeys(VimKeys())
	m = load(t, m, m.Init())
	m, cmd := withCmd(m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"}))
	m = load(t, m, cmd)
	require.Equal(t, dir, m.Dir(), "h goes up")
	require.Equal(t, "sub", filepath.Base(m.Highlighted()), "onto the directory it came from")
	m, _ = withCmd(m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"}))
	require.NotEqual(t, "sub", filepath.Base(m.Highlighted()), "j moves down")
}

func TestHiddenEntriesShowOnlyWhenAsked(t *testing.T) {
	dir := tree(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".secret"), nil, 0o600))
	shut := New(dir)
	shut = load(t, shut, shut.Init())
	require.NotContains(t, strings.Join(rows(shut), "\n"), ".secret")
	open := New(dir).WithHidden(true)
	open = load(t, open, open.Init())
	require.Contains(t, strings.Join(rows(open), "\n"), ".secret")
}

func TestWithDirsLetsEnterPickADirectory(t *testing.T) {
	dir := tree(t)
	m := New(dir).WithDirs(true)
	m = load(t, m, m.Init())
	require.Equal(t, "sub", filepath.Base(m.Highlighted()), "directories come first")
	m, _ = press(m, tea.KeyEnter)
	got, ok := m.Picked()
	require.True(t, ok)
	require.Equal(t, filepath.Join(dir, "sub"), got)
}

func TestTheDrivesMoveAndOpen(t *testing.T) {
	root := filepath.VolumeName(t.TempDir()) + string(filepath.Separator)
	other := "Z:" + string(filepath.Separator)
	m := New(root).WithVolumes(func() []string { return []string{root, other} })
	m, cmd := press(m, tea.KeyLeft)
	m = load(t, m, cmd)
	m, _ = press(m, tea.KeyDown)
	require.Equal(t, other, m.Highlighted())
	m, _ = press(m, tea.KeyUp)
	require.Equal(t, root, m.Highlighted())
	m, cmd = press(m, tea.KeyEnter)
	m = load(t, m, cmd)
	require.True(t, m.Browsing())
	require.Equal(t, root, m.Dir())
}

func TestTheLegendNamesTheKeys(t *testing.T) {
	var got []string
	for _, e := range New(".").Legend() {
		got = append(got, e.Help.Desc)
	}
	require.Equal(t, []string{"move", "up", "open", "pick"}, got)
}

// A file of a type that cannot be picked draws in the Disabled style, and an
// edit to the styles reaches it.
func TestWithStylesPaintsTheRefusedFiles(t *testing.T) {
	marked := lipgloss.NewStyle().Strikethrough(true)
	m := New(tree(t)).WithAllowedTypes(".md").WithStyles(func(s Styles) Styles { s.Disabled = marked; return s })
	m = load(t, m, m.Init())
	require.Contains(t, m.View(th), marked.Render("notes.txt"))
}
