package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/theme"
)

type filePickedMsg struct{ path string }

// pickerTree lays out dir/{notes.txt, main.go, sub/inner.md} and returns dir.
func pickerTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hi"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "sub"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "inner.md"), []byte("# x"), 0o600))
	return dir
}

func newPicker(dir string) FilePicker {
	return NewFilePicker("Open file", dir, func(path string) tea.Cmd {
		return func() tea.Msg { return filePickedMsg{path} }
	}, theme.Resolved{})
}

// load runs a command the picker returned and feeds its message back, the
// way the stack delivers it to the top overlay.
func load(t *testing.T, p FilePicker, cmd tea.Cmd) FilePicker {
	t.Helper()
	require.NotNil(t, cmd)
	o, _ := p.Update(cmd())
	return o.(FilePicker)
}

func pressPicker(t *testing.T, p FilePicker, key tea.KeyPressMsg) (FilePicker, tea.Cmd) {
	t.Helper()
	o, cmd := p.Update(key)
	return o.(FilePicker), cmd
}

func pickerView(p FilePicker) string { return ansi.Strip(p.View(layout.Rect{W: 80, H: 30})) }

// Directories come first, as the bubbles picker sorts them.
func TestFilePickerListsTheDirectory(t *testing.T) {
	dir := pickerTree(t)
	p := newPicker(dir)
	p = load(t, p, p.Init())

	v := pickerView(p)
	require.Contains(t, v, "Open file")
	require.Contains(t, v, filepath.Base(dir), "the directory being browsed is shown")
	require.Contains(t, v, "sub")
	require.Contains(t, v, "main.go")
	require.Contains(t, v, "notes.txt")
	require.Equal(t, filepath.Join(dir, "sub"), p.Highlighted())
}

func TestFilePickerPicksAFileAndCloses(t *testing.T) {
	dir := pickerTree(t)
	p := newPicker(dir)
	p = load(t, p, p.Init())
	p, _ = pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyDown})

	_, cmd := pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	msgs := drain(cmd())
	require.Contains(t, msgs, tea.Msg(filePickedMsg{filepath.Join(dir, "main.go")}))
	require.Contains(t, msgs, tea.Msg(CloseMsg{}), "picking closes the picker")
}

func TestFilePickerEntersADirectoryAndGoesBack(t *testing.T) {
	dir := pickerTree(t)
	p := newPicker(dir)
	p = load(t, p, p.Init())

	p, cmd := pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyEnter})
	p = load(t, p, cmd)
	require.Equal(t, filepath.Join(dir, "sub"), p.Dir())
	require.Contains(t, pickerView(p), "inner.md")

	p, cmd = pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyLeft})
	p = load(t, p, cmd)
	require.Equal(t, dir, p.Dir())
	require.Equal(t, filepath.Join(dir, "sub"), p.Highlighted(), "back returns to the row it left")
}

// A file of the wrong type stays open and says why, instead of closing on
// nothing or picking what the caller cannot use.
func TestFilePickerRefusesATypeItWasNotGiven(t *testing.T) {
	dir := pickerTree(t)
	p := newPicker(dir).WithAllowedTypes(".go")
	p = load(t, p, p.Init())
	p, _ = pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyDown})
	p, _ = pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, filepath.Join(dir, "notes.txt"), p.Highlighted())

	p, cmd := pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		require.NotContains(t, drain(cmd()), tea.Msg(CloseMsg{}))
	}
	require.Contains(t, pickerView(p), "notes.txt is not .go")
}

// A path wider than the modal keeps its end, the part that tells it apart.
func TestFilePickerCutsALongDirectoryFromTheLeft(t *testing.T) {
	p := newPicker(pickerTree(t))
	v := ansi.Strip(p.View(layout.Rect{W: 30, H: 20}))
	dirLine := strings.Split(v, "\n")[1]
	require.True(t, strings.HasPrefix(dirLine, "…"), dirLine)
	require.Contains(t, dirLine, string(filepath.Separator)+filepath.Base(p.Dir()))
}

// esc belongs to the stack, which closes the picker; inside it, esc must not
// mean "parent directory" the way it does in a bare bubbles picker.
func TestFilePickerLeavesEscToTheStack(t *testing.T) {
	dir := pickerTree(t)
	p := newPicker(dir)
	p = load(t, p, p.Init())
	s := Stack{}.Push(p)

	_, handled, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.True(t, handled)
	require.Equal(t, tea.Msg(CloseMsg{}), cmd())
}

// The list fits the body: a short terminal scrolls the entries instead of
// pushing the modal's border off screen.
func TestFilePickerFitsAShortBody(t *testing.T) {
	dir := t.TempDir()
	for i := range 40 {
		require.NoError(t, os.WriteFile(filepath.Join(dir, filepath.Base(dir)+string(rune('a'+i%26))+string(rune('a'+i/26))), nil, 0o600))
	}
	p := newPicker(dir)
	p = load(t, p, p.Init())
	body := layout.Rect{W: 80, H: 16}
	p = p.Resize(body).(FilePicker)
	for range 30 {
		p, _ = pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyDown})
	}
	v := ansi.Strip(p.View(body))
	_, h := lipgloss.Size(v)
	require.LessOrEqual(t, h, body.H)
	require.Contains(t, v, filepath.Base(p.Highlighted()), "the cursor row stays in view")
}

// Open pushes the picker before it reads, so the listing reaches it on top.
func TestFilePickerOpenPushesThenReads(t *testing.T) {
	dir := pickerTree(t)
	s := Stack{}
	for _, c := range sequence(t, newPicker(dir).Open()()) {
		var handled bool
		s, handled, _ = s.Update(c())
		require.True(t, handled)
	}
	require.Contains(t, ansi.Strip(s.Top().View(layout.Rect{W: 80, H: 30})), "notes.txt")
}

// sequence unpacks tea.Sequence's message, an unexported []tea.Cmd, in order.
func sequence(t *testing.T, msg tea.Msg) []tea.Cmd {
	t.Helper()
	v := reflect.ValueOf(msg)
	require.Equal(t, reflect.Slice, v.Kind(), "Open must return a sequence")
	cmds := make([]tea.Cmd, v.Len())
	for i := range cmds {
		cmds[i] = v.Index(i).Interface().(tea.Cmd)
	}
	return cmds
}

func TestFilePickerWithDirsPicksADirectory(t *testing.T) {
	dir := pickerTree(t)
	p := newPicker(dir).WithDirs()
	p = load(t, p, p.Init())

	_, cmd := pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	require.Contains(t, drain(cmd()), tea.Msg(filePickedMsg{filepath.Join(dir, "sub")}))
}

// On Windows bubbles asks for the hidden attribute by bare name, relative to
// the working directory, so only Unix dotfiles are testable here.
func TestFilePickerWithHiddenListsDotfiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bubbles' IsHidden on Windows ignores the directory being browsed")
	}
	dir := pickerTree(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"), nil, 0o600))

	p := newPicker(dir)
	p = load(t, p, p.Init())
	require.NotContains(t, pickerView(p), ".env")

	p = newPicker(dir).WithHidden()
	p = load(t, p, p.Init())
	require.Contains(t, pickerView(p), ".env")
}

// Without onPick, picking still closes: the caller reads nothing back.
func TestFilePickerWithoutOnPickJustCloses(t *testing.T) {
	dir := pickerTree(t)
	p := NewFilePicker("Open", dir, nil, theme.Resolved{})
	p = load(t, p, p.Init())
	p, _ = pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyDown})

	_, cmd := pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Equal(t, tea.Msg(CloseMsg{}), cmd())
}

// The directory listing is the picker's own answer: it must come back as a
// TopMsg, or a shell would hand it to the app and the picker would stay empty.
func TestFilePickerReadsComeBackToTheTop(t *testing.T) {
	dir := pickerTree(t)
	p := newPicker(dir)
	_, ok := p.Init()().(TopMsg)
	require.True(t, ok, "Init answers through TopMsg")

	p = load(t, p, p.Init())
	_, cmd := pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyEnter})
	_, ok = cmd().(TopMsg)
	require.True(t, ok, "opening a directory answers through TopMsg")
}

// "." is browsed as the absolute directory it names: filepath.Dir(".") is
// ".", so a relative start could never go above itself.
func TestFilePickerBrowsesAbsolutePaths(t *testing.T) {
	dir := pickerTree(t)
	t.Chdir(filepath.Join(dir, "sub"))

	p := newPicker(".")
	require.Equal(t, filepath.Join(dir, "sub"), p.Dir())

	p = load(t, p, p.Init())
	p, cmd := pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyLeft})
	p = load(t, p, cmd)
	require.Equal(t, dir, p.Dir(), "back climbs above where it started")
}

// What onPick receives is the full path; the caller shapes it, since only it
// knows what a relative path should be relative to.
func TestFilePickerHandsOverTheFullPath(t *testing.T) {
	dir := pickerTree(t)
	t.Chdir(dir)

	p := newPicker(".")
	p = load(t, p, p.Init())
	p, _ = pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyDown})
	_, cmd := pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Contains(t, drain(cmd()), tea.Msg(filePickedMsg{filepath.Join(dir, "main.go")}))
}

// rootOf is the root of the volume dir is on: `C:\` on Windows, "/" elsewhere.
func rootOf(dir string) string { return filepath.VolumeName(dir) + string(filepath.Separator) }

// ← at a drive's root lists the drives, and opening one browses it: the only
// way to reach another disk on Windows.
func TestFilePickerLeavesARootThroughTheDrives(t *testing.T) {
	other := pickerTree(t)
	start := rootOf(other)
	volumes = func() []string { return []string{start, other} }
	t.Cleanup(func() { volumes = listVolumes })

	p := newPicker(start)
	p, cmd := pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyLeft})
	p = load(t, p, cmd)
	v := pickerView(p)
	require.Contains(t, v, "drives")
	require.Contains(t, v, "> "+start)
	require.Equal(t, start, p.Highlighted(), "the cursor starts on the drive it came from")

	p, _ = pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyDown})
	require.Equal(t, other, p.Highlighted())
	p, cmd = pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyRight})
	p = load(t, p, cmd)
	require.Equal(t, other, p.Dir())
	require.Equal(t, filepath.Join(other, "sub"), p.Highlighted())
	require.Contains(t, pickerView(p), "notes.txt")
}

// With no drive list, as off Windows, ← at the root just stays there.
func TestFilePickerStaysAtARootWithoutDrives(t *testing.T) {
	volumes = func() []string { return nil }
	t.Cleanup(func() { volumes = listVolumes })

	start := rootOf(t.TempDir())
	p := newPicker(start)
	p, cmd := pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyLeft})
	p = load(t, p, cmd)
	require.Equal(t, start, p.Dir())
	require.NotContains(t, pickerView(p), "drives")
}

// More drives than fit scroll as a list does: the cursor in view, the rest counted.
func TestFilePickerScrollsALongDriveList(t *testing.T) {
	start := rootOf(t.TempDir())
	drives := []string{start}
	for i := range 12 {
		drives = append(drives, fmt.Sprintf("D%02d:%c", i, filepath.Separator))
	}
	volumes = func() []string { return drives }
	t.Cleanup(func() { volumes = listVolumes })

	body := layout.Rect{W: 60, H: 14}
	p := newPicker(start).Resize(body).(FilePicker)
	p, cmd := pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyLeft})
	p = load(t, p, cmd)
	require.Contains(t, ansi.Strip(p.View(body)), "↓ 6 more")

	for range 8 {
		p, _ = pressPicker(t, p, tea.KeyPressMsg{Code: tea.KeyDown})
	}
	require.Equal(t, drives[8], p.Highlighted())
	require.Contains(t, ansi.Strip(p.View(body)), drives[8])
}
