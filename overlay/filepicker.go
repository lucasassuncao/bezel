package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// FilePicker browses directories in a modal. Enter picks a file and closes,
// running onPick; enter or → opens a directory, ← or backspace goes up, and
// from a drive's root lists the drives. esc is the stack's. Push it with Open.
type FilePicker struct {
	look
	title    string
	dir      string      // absolute, so ← can always climb
	entries  []fileEntry // nil until the first listing arrives
	win      layout.Window
	reselect string // the entry to land on when the listing arrives
	allowed  []string
	dirs     bool // enter picks a directory instead of opening it
	hidden   bool
	onPick   func(path string) tea.Cmd
	problem  string
	drives   driveList
	keys     FilePickerKeys
}

// fileEntry is one row of a listing.
type fileEntry struct {
	name string
	dir  bool   // a directory, or a link to one
	link string // where a symbolic link points, "" for anything else
	size int64
}

// driveList is the drives a ← at a root opened, shown in place of a directory
// until one is opened. The zero value is closed.
type driveList struct {
	roots []string
	win   layout.Window
}

func (d driveList) open() bool { return d.roots != nil }

// FilePickerKeys are the bindings a FilePicker answers to besides esc. Pick
// shares enter with Open: on a directory it opens it, unless WithDirs.
type FilePickerKeys struct {
	Up, Down, PageUp, PageDown, Top, Bottom key.Binding
	Back, Open, Pick                        key.Binding
}

// DefaultFilePickerKeys are the arrows and hjkl, pgup/pgdown, home/end with
// g/G, backspace to go up and enter to pick.
func DefaultFilePickerKeys() FilePickerKeys {
	return FilePickerKeys{
		Up:       key.NewBinding(key.WithKeys("up", "k", "ctrl+p"), key.WithHelp("↑/↓", "move")),
		Down:     key.NewBinding(key.WithKeys("down", "j", "ctrl+n")),
		PageUp:   key.NewBinding(key.WithKeys("pgup", "K")),
		PageDown: key.NewBinding(key.WithKeys("pgdown", "J")),
		Top:      key.NewBinding(key.WithKeys("home", "g")),
		Bottom:   key.NewBinding(key.WithKeys("end", "G")),
		Back:     key.NewBinding(key.WithKeys("left", "h", "backspace"), key.WithHelp("←", "up")),
		Open:     key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→", "open")),
		Pick:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "pick")),
	}
}

// pickerChrome is the lines a FilePicker spends around its entries: the
// title, the directory and the hint, and a gap after each of the first two.
const pickerChrome = 6

// pickerRows is how many entries a FilePicker shows before its first Resize.
const pickerRows = 10

// sizeWidth is the column file sizes are right-aligned in, as "999.9MB".
const sizeWidth = 7

// listedMsg answers the read of dir.
type listedMsg struct {
	dir     string
	entries []fileEntry
	err     error
}

// volumesMsg answers the drive list a ← at a root asked for.
type volumesMsg struct{ roots []string }

// volumes lists the drives; a variable so tests can stand in their own.
var volumes = listVolumes

// NewFilePicker browses dir, files only, without hidden entries. onPick,
// which may be nil, gets the absolute path.
func NewFilePicker(title, dir string, onPick func(path string) tea.Cmd, th theme.Resolved) FilePicker {
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	return FilePicker{
		look: lookOf(th), title: title, dir: dir, onPick: onPick,
		win: layout.Window{Height: pickerRows}, keys: DefaultFilePickerKeys(),
	}
}

// WithKeys rebinds the picker.
func (p FilePicker) WithKeys(k FilePickerKeys) FilePicker {
	p.keys = k
	return p
}

// Resize shows as many entries as fit in body.
func (p FilePicker) Resize(body layout.Rect) Overlay {
	rows := p.fit(body)
	p.win.Height, p.drives.win.Height = rows, rows
	p.win = p.win.Clamp(len(p.entries))
	p.drives.win = p.drives.win.Clamp(len(p.drives.roots))
	return p
}

// fit is how many entries body leaves room for, at least one.
func (p FilePicker) fit(body layout.Rect) int {
	return max(1, body.H-p.style().GetVerticalFrameSize()-pickerChrome)
}

// WithAllowedTypes limits the pick to names ending in one of exts, such as
// ".go"; other files stay listed but enter refuses them and says why.
func (p FilePicker) WithAllowedTypes(exts ...string) FilePicker {
	p.allowed = slices.Clone(exts)
	return p
}

// WithDirs lets enter pick a directory; → still opens one.
func (p FilePicker) WithDirs() FilePicker {
	p.dirs = true
	return p
}

// WithHidden lists hidden entries too: dotfiles, and on Windows the ones
// marked hidden. Set it before Init.
func (p FilePicker) WithHidden() FilePicker {
	p.hidden = true
	return p
}

// Init reads the first directory, answering through TopMsg. The picker must
// already be on top of the stack to receive it: push it first, as Open does.
func (p FilePicker) Init() tea.Cmd { return ToTop(list(p.dir, p.hidden)) }

// Open pushes the picker and then reads its directory, in that order.
func (p FilePicker) Open() tea.Cmd { return tea.Sequence(Open(p), p.Init()) }

// Dir is the directory being browsed.
func (p FilePicker) Dir() string { return p.dir }

// Highlighted is the path under the cursor, or "" in an empty directory.
func (p FilePicker) Highlighted() string {
	if p.drives.open() {
		return p.drives.roots[p.drives.win.Cursor]
	}
	if e, ok := p.current(); ok {
		return filepath.Join(p.dir, e.name)
	}
	return ""
}

func (p FilePicker) current() (fileEntry, bool) {
	if p.win.Cursor >= len(p.entries) {
		return fileEntry{}, false
	}
	return p.entries[p.win.Cursor], true
}

// list reads dir: directories first, then files, each by name.
func list(dir string, hidden bool) tea.Cmd {
	return func() tea.Msg {
		des, err := os.ReadDir(dir)
		if err != nil {
			return listedMsg{dir: dir, err: err}
		}
		entries := make([]fileEntry, 0, len(des))
		for _, de := range des {
			path := filepath.Join(dir, de.Name())
			if !hidden && isHidden(path) {
				continue
			}
			e := fileEntry{name: de.Name(), dir: de.IsDir()}
			if info, err := de.Info(); err == nil {
				e.size = info.Size()
				if info.Mode()&os.ModeSymlink != 0 {
					e.link, _ = filepath.EvalSymlinks(path)
					if target, err := os.Stat(path); err == nil {
						e.dir = target.IsDir()
					}
				}
			}
			entries = append(entries, e)
		}
		slices.SortStableFunc(entries, func(a, b fileEntry) int {
			if a.dir != b.dir {
				if a.dir {
					return -1
				}
				return 1
			}
			return strings.Compare(a.name, b.name)
		})
		return listedMsg{dir: dir, entries: entries}
	}
}

// isHidden is a dotfile anywhere, or a file Windows marks hidden.
func isHidden(path string) bool {
	return strings.HasPrefix(filepath.Base(path), ".") || hiddenAttr(path)
}

// browse starts dir afresh, landing on reselect once its listing arrives.
func (p FilePicker) browse(dir, reselect string) (Overlay, tea.Cmd) {
	p.dir, p.entries, p.reselect, p.drives = dir, nil, reselect, driveList{}
	p.win = layout.Window{Height: p.win.Height}
	return p, ToTop(list(dir, p.hidden))
}

// allows reports whether a file of this name may be picked.
func (p FilePicker) allows(name string) bool {
	if len(p.allowed) == 0 {
		return true
	}
	return slices.ContainsFunc(p.allowed, func(ext string) bool { return strings.HasSuffix(name, ext) })
}

func (p FilePicker) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if m, ok := msg.(TopMsg); ok {
		msg = m.Msg
	}
	switch m := msg.(type) {
	case listedMsg:
		return p.listed(m), nil
	case volumesMsg:
		if len(m.roots) > 0 {
			at := max(slices.Index(m.roots, p.dir), 0)
			p.drives = driveList{roots: m.roots, win: layout.Window{Cursor: at, Height: p.win.Height}.Clamp(len(m.roots))}
		}
		return p, nil
	case tea.KeyPressMsg:
		p.problem = ""
		if p.drives.open() {
			return p.updateDrives(m)
		}
		return p.key(m)
	}
	return p, nil
}

// listed takes the listing of the directory still being browsed; a late one
// for a directory already left is dropped.
func (p FilePicker) listed(m listedMsg) FilePicker {
	if m.dir != p.dir {
		return p
	}
	if m.err != nil {
		p.problem = "cannot read it: " + m.err.Error()
	}
	p.entries = m.entries
	if p.entries == nil {
		p.entries = []fileEntry{}
	}
	p.win.Cursor = max(slices.IndexFunc(p.entries, func(e fileEntry) bool { return e.name == p.reselect }), 0)
	p.reselect = ""
	p.win = p.win.Clamp(len(p.entries))
	return p
}

func (p FilePicker) key(k tea.KeyPressMsg) (Overlay, tea.Cmd) {
	n, page := len(p.entries), max(1, p.win.Height-1)
	switch {
	case key.Matches(k, p.keys.Up):
		p.win = p.win.Move(-1, n, nil)
	case key.Matches(k, p.keys.Down):
		p.win = p.win.Move(1, n, nil)
	case key.Matches(k, p.keys.PageUp):
		p.win = p.win.Move(-page, n, nil)
	case key.Matches(k, p.keys.PageDown):
		p.win = p.win.Move(page, n, nil)
	case key.Matches(k, p.keys.Top):
		p.win = p.win.Move(-n, n, nil)
	case key.Matches(k, p.keys.Bottom):
		p.win = p.win.Move(n, n, nil)
	case key.Matches(k, p.keys.Back):
		if up := filepath.Dir(p.dir); up != p.dir {
			return p.browse(up, filepath.Base(p.dir))
		}
		return p, ToTop(func() tea.Msg { return volumesMsg{volumes()} })
	case key.Matches(k, p.keys.Open):
		if e, ok := p.current(); ok && e.dir {
			return p.browse(filepath.Join(p.dir, e.name), "")
		}
	case key.Matches(k, p.keys.Pick):
		return p.pick()
	}
	return p, nil
}

// pick is enter: a directory opens unless WithDirs, a file of an allowed
// type is picked, and any other file stays and says why not.
func (p FilePicker) pick() (Overlay, tea.Cmd) {
	e, ok := p.current()
	switch {
	case !ok:
		return p, nil
	case e.dir && !p.dirs:
		return p.browse(filepath.Join(p.dir, e.name), "")
	case !e.dir && !p.allows(e.name):
		p.problem = fmt.Sprintf("%s is not %s", e.name, strings.Join(p.allowed, " or "))
		return p, nil
	case p.onPick == nil:
		return p, Close()
	}
	return p, tea.Batch(p.onPick(filepath.Join(p.dir, e.name)), Close())
}

// updateDrives moves through the drive list; opening one browses it afresh.
func (p FilePicker) updateDrives(k tea.KeyPressMsg) (Overlay, tea.Cmd) {
	n := len(p.drives.roots)
	switch {
	case key.Matches(k, p.keys.Up):
		p.drives.win = p.drives.win.Move(-1, n, nil)
	case key.Matches(k, p.keys.Down):
		p.drives.win = p.drives.win.Move(1, n, nil)
	case key.Matches(k, p.keys.Open, p.keys.Pick):
		return p.browse(p.drives.roots[p.drives.win.Cursor], "")
	}
	return p, nil
}

func (p FilePicker) Legend() []legend.Entry {
	esc := legend.New("esc", "cancel")
	if p.drives.open() {
		return append(legend.FromBindings([]key.Binding{p.keys.Up, p.keys.Open}), esc)
	}
	return append(legend.FromBindings([]key.Binding{p.keys.Up, p.keys.Back, p.keys.Open, p.keys.Pick}), esc)
}

func (p FilePicker) View(body layout.Rect) string {
	rows := p.fit(body)
	entries, where := p.entryRows(rows), tail(p.Dir(), Inner(body))
	if p.drives.open() {
		entries, where = p.driveRows(rows), "drives"
	}
	lines := []string{accent(p.style()).Bold(true).Render(p.title), p.hint().Text.Render(where), ""}
	lines = append(lines, entries...)
	hint := legend.HintLine(p.Legend(), p.hint())
	if p.problem != "" {
		hint = p.hint().Text.Render(p.problem) + "   " + hint
	}
	return Box(append(lines, "", hint), body, p.style())
}

// window is w fit to rows and n entries, as View draws it: a body shorter
// than the last Resize gave still keeps the cursor in view.
func window(w layout.Window, rows, n int) layout.Window {
	w.Height = min(rows, max(w.Height, 1))
	return w.Clamp(n)
}

// entryRows draws the listing in a window that keeps the cursor in view and
// counts the entries below it.
func (p FilePicker) entryRows(rows int) []string {
	if p.entries == nil {
		return nil
	}
	if len(p.entries) == 0 {
		return []string{p.hint().Text.Render("  nothing here")}
	}
	ink, hint := accent(p.style()), p.hint().Text
	w := window(p.win, rows, len(p.entries))
	start, end, below := w.Bounds(len(p.entries))
	out := make([]string, 0, end-start+1)
	for i := start; i < end; i++ {
		e := p.entries[i]
		size := strings.Repeat(" ", sizeWidth)
		if !e.dir {
			size = fmt.Sprintf("%*s", sizeWidth, humanBytes(e.size))
		}
		name := e.name
		if e.link != "" {
			name += " → " + e.link
		}
		disabled := !e.dir && !p.allows(e.name)
		switch {
		case i == w.Cursor && disabled:
			out = append(out, hint.Render(">"+size+" "+name))
		case i == w.Cursor:
			out = append(out, ink.Render(">")+ink.Bold(true).Render(size+" "+name))
		default:
			out = append(out, " "+hint.Render(size)+" "+p.entryStyle(e, disabled, ink).Render(name))
		}
	}
	if below > 0 {
		out = append(out, hint.Render(fmt.Sprintf("  ↓ %d more", below)))
	}
	return out
}

// entryStyle inks directories and links, and greys out what cannot be picked.
func (p FilePicker) entryStyle(e fileEntry, disabled bool, ink lipgloss.Style) lipgloss.Style {
	switch {
	case e.dir, e.link != "":
		return ink
	case disabled:
		return p.hint().Text
	}
	return lipgloss.NewStyle()
}

// driveRows draws the drive list the way entryRows draws a directory.
func (p FilePicker) driveRows(rows int) []string {
	ink, roots := accent(p.style()), p.drives.roots
	w := window(p.drives.win, rows, len(roots))
	start, end, below := w.Bounds(len(roots))
	var out []string
	for i := start; i < end; i++ {
		if i == w.Cursor {
			out = append(out, ink.Render(">")+ink.Bold(true).Render(" "+roots[i]))
			continue
		}
		out = append(out, " "+ink.Render(" "+roots[i]))
	}
	if below > 0 {
		out = append(out, p.hint().Text.Render(fmt.Sprintf("  ↓ %d more", below)))
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

// humanBytes is a size in SI units the way a listing shows one: "12B",
// "1.5kB", "640MB", one decimal below ten. A value that would round up to
// 1000 of a unit is shown as the next one: 999,950 is "1.0MB".
func humanBytes(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%dB", max(n, 0))
	}
	units := []string{"kB", "MB", "GB", "TB", "PB", "EB"}
	v, u := float64(n)/1000, 0
	for v >= 999.5 && u < len(units)-1 {
		v, u = v/1000, u+1
	}
	if v < 9.95 {
		return fmt.Sprintf("%.1f%s", v, units[u])
	}
	return fmt.Sprintf("%.0f%s", v, units[u])
}
