package filepicker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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
