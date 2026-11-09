package filepicker

import (
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestChosenCarriesThePath(t *testing.T) {
	require.Equal(t, "notes.txt", filepath.Base(Chosen{Path: "/tmp/notes.txt"}.Path))
}

func TestEnterAndAClickReportChosen(t *testing.T) {
	dir := tree(t)
	m := New(dir)
	m = load(t, m, m.Init())
	m, r := m.Update(clickAt(2))
	require.Nil(t, r.Event, "the first click only moves the cursor")
	_, r = m.Update(clickAt(2))
	require.Equal(t, Chosen{Path: filepath.Join(dir, "notes.txt")}, r.Event)

	m, _ = press(m, 'x')
	_, r = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Equal(t, Chosen{Path: filepath.Join(dir, "notes.txt")}, r.Event)
}

func TestSelectedIsThePathUnderTheCursor(t *testing.T) {
	dir := tree(t)
	m := New(dir)
	_, ok := m.Selected()
	require.False(t, ok, "nothing before the listing comes")
	m = load(t, m, m.Init())
	path, ok := m.Selected()
	require.True(t, ok)
	require.Equal(t, filepath.Join(dir, "sub"), path)
}
