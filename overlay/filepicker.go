package overlay

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// FilePicker is the bubbles file picker in a modal. Enter picks a file and
// closes, running onPick; enter or → opens a directory, ← or backspace goes up,
// and from a drive's root lists the drives. esc is the stack's. Push it with Open.
type FilePicker struct {
	look
	title   string
	picker  filepicker.Model
	onPick  func(path string) tea.Cmd
	problem string
	// volumes is the drive list shown above a root, nil while browsing.
	volumes []string
	volume  int
	rows    int // the entries that fit at the last Resize, pickerRows before one
}

// pickerChrome is the lines a FilePicker spends around its entries: the
// title, the directory and the hint, and a gap after each of the first two.
const pickerChrome = 6

// pickerRows is how many entries a FilePicker shows before its first Resize.
const pickerRows = 10

// volumesMsg answers the drive list a ← at a root asked for.
type volumesMsg struct{ roots []string }

// volumes lists the drives; a variable so tests can stand in their own.
var volumes = listVolumes

// NewFilePicker browses dir, files only, without hidden entries or
// permissions. onPick, which may be nil, gets the absolute path.
func NewFilePicker(title, dir string, onPick func(path string) tea.Cmd, th theme.Resolved) FilePicker {
	fp := filepicker.New()
	fp.CurrentDirectory = dir
	fp.ShowPermissions = false
	p := FilePicker{look: look{th: th}, title: title, onPick: onPick, rows: pickerRows}
	fp.Styles = pickerStyles(p.look)
	return p.WithPicker(fp)
}

// WithPicker replaces the bubbles model, for a caller that sets its own
// fields. Esc is taken out of its Back keys: here esc closes the modal.
func (p FilePicker) WithPicker(fp filepicker.Model) FilePicker {
	// Browsed absolute: filepath.Dir(".") is ".", so ← could not climb above it.
	if abs, err := filepath.Abs(fp.CurrentDirectory); err == nil {
		fp.CurrentDirectory = abs
	}
	fp.AutoHeight = false
	fp.KeyMap.Back = withoutKey(fp.KeyMap.Back, "esc")
	fp.SetHeight(p.rows)
	p.picker = fp
	return p
}

// Resize shows as many entries as fit in body, which moves the picker's
// window to keep its cursor in view.
func (p FilePicker) Resize(body layout.Rect) Overlay {
	if p.rows = p.fit(body); p.rows != p.picker.Height() {
		p.picker.SetHeight(p.rows)
		// Without AutoHeight a size message only re-anchors the window.
		p.picker, _ = p.picker.Update(tea.WindowSizeMsg{})
	}
	return p
}

// fit is how many entries body leaves room for, at least one.
func (p FilePicker) fit(body layout.Rect) int {
	return max(1, body.H-p.style().GetVerticalFrameSize()-pickerChrome)
}

// WithAllowedTypes limits the pick to names ending in one of exts, such as
// ".go"; other files stay listed but enter refuses them and says why.
func (p FilePicker) WithAllowedTypes(exts ...string) FilePicker {
	p.picker.AllowedTypes = slices.Clone(exts)
	return p
}

// WithDirs lets enter pick a directory as well as open it.
func (p FilePicker) WithDirs() FilePicker {
	p.picker.DirAllowed = true
	return p
}

// WithHidden lists hidden entries too; set it before Init. On Windows bubbles
// checks the hidden attribute in the working directory, not the one browsed.
func (p FilePicker) WithHidden() FilePicker {
	p.picker.ShowHidden = true
	return p
}

// Init reads the first directory, answering through TopMsg. The picker must
// already be on top of the stack to receive it: push it first, as Open does.
func (p FilePicker) Init() tea.Cmd { return ToTop(p.picker.Init()) }

// Open pushes the picker and then reads its directory, in that order.
func (p FilePicker) Open() tea.Cmd { return tea.Sequence(Open(p), p.Init()) }

// Dir is the directory being browsed.
func (p FilePicker) Dir() string { return p.picker.CurrentDirectory }

// Highlighted is the path under the cursor, or "" in an empty directory.
func (p FilePicker) Highlighted() string {
	if p.volumes != nil {
		return p.volumes[p.volume]
	}
	return p.picker.HighlightedPath()
}

// atRoot reports a ← that has nowhere above to go: a drive's root.
func (p FilePicker) atRoot(msg tea.Msg) bool {
	k, ok := msg.(tea.KeyPressMsg)
	dir := p.picker.CurrentDirectory
	return ok && key.Matches(k, p.picker.KeyMap.Back) && filepath.Dir(dir) == dir
}

// updateVolumes moves through the drive list; opening one browses it afresh.
func (p FilePicker) updateVolumes(msg tea.Msg) (Overlay, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	km := p.picker.KeyMap
	switch {
	case !ok:
	case key.Matches(k, km.Up):
		p.volume = max(p.volume-1, 0)
	case key.Matches(k, km.Down):
		p.volume = min(p.volume+1, len(p.volumes)-1)
	case key.Matches(k, km.Open, km.Select):
		return p.browse(p.volumes[p.volume])
	}
	return p, nil
}

// browse starts over in dir with a new bubbles model, whose cursor and
// history belong to no other directory, and reads it.
func (p FilePicker) browse(dir string) (Overlay, tea.Cmd) {
	old := p.picker
	fp := filepicker.New()
	fp.CurrentDirectory, fp.Cursor, fp.AllowedTypes = dir, old.Cursor, old.AllowedTypes
	fp.ShowPermissions, fp.ShowSize, fp.ShowHidden = old.ShowPermissions, old.ShowSize, old.ShowHidden
	fp.DirAllowed, fp.FileAllowed = old.DirAllowed, old.FileAllowed
	fp.KeyMap, fp.Styles = old.KeyMap, old.Styles
	p, p.volumes = p.WithPicker(fp), nil
	return p, p.Init()
}

func (p FilePicker) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if m, ok := msg.(TopMsg); ok {
		msg = m.Msg
	}
	if _, ok := msg.(tea.KeyPressMsg); ok {
		p.problem = ""
	}
	if m, ok := msg.(volumesMsg); ok && len(m.roots) > 0 {
		p.volumes, p.volume = m.roots, max(slices.Index(m.roots, p.picker.CurrentDirectory), 0)
		return p, nil
	}
	if p.volumes != nil {
		return p.updateVolumes(msg)
	}
	if p.atRoot(msg) {
		return p, ToTop(func() tea.Msg { return volumesMsg{volumes()} })
	}
	var cmd tea.Cmd
	p.picker, cmd = p.picker.Update(msg)
	if ok, path := p.picker.DidSelectFile(msg); ok {
		if p.onPick == nil {
			return p, Close()
		}
		return p, tea.Batch(p.onPick(path), Close())
	}
	if ok, path := p.picker.DidSelectDisabledFile(msg); ok {
		p.problem = fmt.Sprintf("%s is not %s", filepath.Base(path), strings.Join(p.picker.AllowedTypes, " or "))
	}
	return p, ToTop(cmd)
}

func (p FilePicker) Legend() []legend.Entry {
	if p.volumes != nil {
		return []legend.Entry{legend.New("↑/↓", "move"), legend.New("→", "open"), legend.New("esc", "cancel")}
	}
	return []legend.Entry{
		legend.New("↑/↓", "move"), legend.New("←", "up"), legend.New("→", "open"),
		legend.New("enter", "pick"), legend.New("esc", "cancel"),
	}
}

func (p FilePicker) View(body layout.Rect) string {
	rows := p.fit(body)
	// A shorter body than the last Resize gave: draw the window that fits.
	fp := p.picker
	fp.SetHeight(min(rows, fp.Height()))
	entries := strings.Split(strings.TrimRight(fp.View(), "\n"), "\n")
	where := tail(p.Dir(), Inner(body))
	if p.volumes != nil {
		entries, where = p.volumeRows(rows), "drives"
	}

	lines := []string{accent(p.style()).Bold(true).Render(p.title), p.hint().Text.Render(where), ""}
	lines = append(lines, entries[:min(len(entries), rows)]...)
	hint := legend.HintLine(p.Legend(), p.hint())
	if p.problem != "" {
		hint = p.hint().Text.Render(p.problem) + "   " + hint
	}
	return Box(append(lines, "", hint), body, p.style())
}

// volumeRows draws the drive list in the picker's styles, scrolled so the
// cursor stays within rows.
func (p FilePicker) volumeRows(rows int) []string {
	st := p.picker.Styles
	first := max(0, p.volume-rows+1)
	var out []string
	for i, v := range p.volumes[first:min(len(p.volumes), first+rows)] {
		if first+i == p.volume {
			out = append(out, st.Cursor.Render(p.picker.Cursor)+st.Selected.Render(" "+v))
			continue
		}
		out = append(out, strings.Repeat(" ", draw.Width(p.picker.Cursor))+st.Directory.Render(" "+v))
	}
	return out
}

// tail cuts a path from the left: its end is what tells directories apart.
func tail(path string, width int) string {
	if over := draw.Width(path) - width; over > 0 {
		return ansi.TruncateLeft(path, over+1, "…")
	}
	return path
}

// pickerStyles paints the bubbles picker in the dialog's colours instead of
// its own fixed palette.
func pickerStyles(l look) filepicker.Styles {
	ink, hint := accent(l.style()), l.hint()
	s := filepicker.DefaultStyles()
	s.Cursor, s.Selected, s.Directory, s.Symlink = ink, ink.Bold(true), ink, ink
	s.File = lipgloss.NewStyle()
	s.DisabledCursor, s.DisabledSelected, s.DisabledFile = hint.Text, hint.Text, hint.Text
	s.Permission = hint.Text
	s.FileSize = hint.Text.Width(s.FileSize.GetWidth()).Align(lipgloss.Right)
	s.EmptyDirectory = hint.Text.PaddingLeft(2).SetString("nothing here")
	return s
}

// withoutKey is b answering to every key it did except drop.
func withoutKey(b key.Binding, drop string) key.Binding {
	keys := slices.DeleteFunc(slices.Clone(b.Keys()), func(k string) bool { return k == drop })
	h := b.Help()
	return key.NewBinding(key.WithKeys(keys...), key.WithHelp(h.Key, h.Desc))
}
