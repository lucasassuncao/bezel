// Package filepicker browses directories inside a pane: directories first,
// sizes and permissions in columns, enter to pick a file, ← to climb and
// from a drive's root on to the other drives. overlay.FilePicker is built on it.
package filepicker

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/internal/frames"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Keys are the bindings the picker answers to. Pick shares enter with Open:
// on a directory it opens it, unless WithDirs.
type Keys struct {
	Up, Down, PageUp, PageDown, Top, Bottom key.Binding
	Back, Open, Pick                        key.Binding
}

// DefaultKeys are the arrows, pgup/pgdown, home/end, backspace to go up and
// enter to pick.
func DefaultKeys() Keys {
	return Keys{
		Up:       key.New("up", "ctrl+p").WithHelp("↑/↓", "move"),
		Down:     key.New("down", "ctrl+n"),
		PageUp:   key.New("pgup"),
		PageDown: key.New("pgdown"),
		Top:      key.New("home"),
		Bottom:   key.New("end"),
		Back:     key.New("left", "backspace").WithHelp("←", "up"),
		Open:     key.New("right").WithHelp("→", "open"),
		Pick:     key.New("enter").WithHelp("enter", "pick"),
	}
}

// VimKeys are DefaultKeys with vim's letters added: k/j to move, K/J to page,
// g/G to the ends, h to go up and l to open.
func VimKeys() Keys {
	k := DefaultKeys()
	k.Up, k.Down = key.New("up", "k", "ctrl+p").WithHelp("↑/↓ k/j", "move"), key.New("down", "j", "ctrl+n")
	k.PageUp, k.PageDown = key.New("pgup", "K"), key.New("pgdown", "J")
	k.Top, k.Bottom = key.New("home", "g"), key.New("end", "G")
	k.Back = key.New("left", "h", "backspace").WithHelp("h", "up")
	k.Open = key.New("right", "l").WithHelp("l", "open")
	return k
}

// Styles are how the rows are painted.
type Styles struct {
	Cursor   lipgloss.Style // the row under the cursor, after its mark
	Mark     lipgloss.Style // the ">" before it
	Dir      lipgloss.Style // directories and links
	File     lipgloss.Style
	Disabled lipgloss.Style // a file of a type that cannot be picked
	Meta     lipgloss.Style // the size and permission columns, and notes
}

// DefaultStyles ink directories and the cursor in the accent and the columns
// muted.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{
		Cursor: th.Accent.Bold(true), Mark: th.Accent, Dir: th.Accent,
		File: lipgloss.NewStyle(), Disabled: th.Muted, Meta: th.Muted,
	}
}

// entry is one row of a listing.
type entry struct {
	name string
	dir  bool   // a directory, or a link to one
	link string // where a symbolic link points, "" for anything else
	size int64
	mode os.FileMode
}

// drives is the drive list a ← at a root opened, shown in place of a
// directory until one is opened. The zero value is closed.
type drives struct {
	roots []string
	win   layout.Window
}

func (d drives) open() bool { return d.roots != nil }

// listedMsg answers the read of dir for the picker id.
type listedMsg struct {
	id      int64
	dir     string
	entries []entry
	err     error
}

// volumesMsg answers the drive list a ← at a root asked for.
type volumesMsg struct {
	id    int64
	roots []string
}

// sizeWidth is the column file sizes are right-aligned in, as "999.9MB".
const sizeWidth = 7

// Model is a value: every method returns the new picker.
type Model struct {
	id       int64
	dir      string  // absolute, so ← can always climb
	entries  []entry // nil until the first listing arrives
	win      layout.Window
	reselect string // the entry to land on when the listing arrives
	allowed  []string
	dirs     bool // enter picks a directory instead of opening it
	hidden   bool
	noSizes  bool
	perms    bool
	drives   drives
	keys     Keys
	problem  string
	picked   string // the path the last Update picked
	volumes  func() []string
	edit     func(Styles) Styles
}

// New browses dir, files only, without hidden entries, ten rows tall. Call
// Init for the first listing.
func New(dir string) Model {
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	return Model{id: frames.NextID(), dir: dir, win: layout.Window{Height: 10}, keys: DefaultKeys()}
}

// WithKeys rebinds the picker.
func (m Model) WithKeys(k Keys) Model {
	m.keys = k
	return m
}

// WithStyles changes the styles derived from the theme through edit.
func (m Model) WithStyles(edit func(Styles) Styles) Model {
	m.edit = edit
	return m
}

// WithAllowedTypes limits the pick to names ending in one of exts, such as
// ".go"; other files stay listed but enter refuses them and says why.
func (m Model) WithAllowedTypes(exts ...string) Model {
	m.allowed = slices.Clone(exts)
	return m
}

// WithDirs lets enter pick a directory; → still opens one.
func (m Model) WithDirs(on bool) Model {
	m.dirs = on
	return m
}

// WithHidden lists hidden entries too: dotfiles, and on Windows the ones
// marked hidden. Set it before Init.
func (m Model) WithHidden(on bool) Model {
	m.hidden = on
	return m
}

// WithSizes shows or hides the column of file sizes; it shows by default.
func (m Model) WithSizes(on bool) Model {
	m.noSizes = !on
	return m
}

// WithPermissions shows or hides a column of permissions, as ls -l writes
// them; it is hidden by default.
func (m Model) WithPermissions(on bool) Model {
	m.perms = on
	return m
}

// WithVolumes lists the drives through fn when ← leaves a root, for a test
// that must not read the machine's; nil is the system's.
func (m Model) WithVolumes(fn func() []string) Model {
	m.volumes = fn
	return m
}

// SetHeight shows as many entries as h rows hold, at least one.
func (m Model) SetHeight(h int) Model {
	h = max(h, 1)
	m.win.Height, m.drives.win.Height = h, h
	m.win = m.win.Clamp(len(m.entries))
	m.drives.win = m.drives.win.Clamp(len(m.drives.roots))
	return m
}

// Height is how many entries the picker shows.
func (m Model) Height() int { return m.win.Height }

// Init reads the first directory.
func (m Model) Init() tea.Cmd { return m.list(m.dir) }

// Dir is the directory being browsed.
func (m Model) Dir() string { return m.dir }

// Browsing reports whether a directory shows, false while the drive list does.
func (m Model) Browsing() bool { return !m.drives.open() }

// Problem is why the last key did nothing, such as a file of the wrong type;
// "" when it did.
func (m Model) Problem() string { return m.problem }

// Picked is the absolute path the last Update picked, and whether it picked.
func (m Model) Picked() (string, bool) { return m.picked, m.picked != "" }

// Highlighted is the path under the cursor, or "" in an empty directory.
func (m Model) Highlighted() string {
	if m.drives.open() {
		return m.drives.roots[m.drives.win.Cursor]
	}
	if e, ok := m.current(); ok {
		return filepath.Join(m.dir, e.name)
	}
	return ""
}

func (m Model) current() (entry, bool) {
	if m.win.Cursor >= len(m.entries) {
		return entry{}, false
	}
	return m.entries[m.win.Cursor], true
}

// list reads dir: directories first, then files, each by name.
func (m Model) list(dir string) tea.Cmd {
	id, hidden := m.id, m.hidden
	return func() tea.Msg {
		des, err := os.ReadDir(dir)
		if err != nil {
			return listedMsg{id: id, dir: dir, err: err}
		}
		entries := make([]entry, 0, len(des))
		for _, de := range des {
			path := filepath.Join(dir, de.Name())
			info, err := de.Info()
			if !hidden && isHidden(de.Name(), info) {
				continue
			}
			e := entry{name: de.Name(), dir: de.IsDir()}
			if err == nil {
				e.size, e.mode = info.Size(), info.Mode()
				if info.Mode()&os.ModeSymlink != 0 {
					e.link, _ = filepath.EvalSymlinks(path)
					if target, err := os.Stat(path); err == nil {
						e.dir = target.IsDir()
					}
				}
			}
			entries = append(entries, e)
		}
		slices.SortStableFunc(entries, func(a, b entry) int {
			if a.dir != b.dir {
				if a.dir {
					return -1
				}
				return 1
			}
			return strings.Compare(a.name, b.name)
		})
		return listedMsg{id: id, dir: dir, entries: entries}
	}
}

// isHidden is a dotfile anywhere, or a file Windows marks hidden; info is
// nil when it could not be read.
func isHidden(name string, info os.FileInfo) bool {
	return strings.HasPrefix(name, ".") || info != nil && hiddenAttr(info)
}

// browse starts dir afresh, landing on reselect once its listing arrives.
func (m Model) browse(dir, reselect string) (Model, tea.Cmd) {
	m.dir, m.entries, m.reselect, m.drives = dir, nil, reselect, drives{}
	m.win = layout.Window{Height: m.win.Height}
	return m, m.list(dir)
}

// allows reports whether a file of this name may be picked.
func (m Model) allows(name string) bool {
	if len(m.allowed) == 0 {
		return true
	}
	return slices.ContainsFunc(m.allowed, func(ext string) bool { return strings.HasSuffix(name, ext) })
}

// Update takes the picker's own listings and its keys; Picked then says
// whether enter picked something.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	m.picked = ""
	switch msg := msg.(type) {
	case listedMsg:
		if msg.id == m.id {
			return m.listed(msg), nil
		}
	case volumesMsg:
		if msg.id == m.id && len(msg.roots) > 0 {
			at := max(slices.Index(msg.roots, m.dir), 0)
			m.drives = drives{roots: msg.roots, win: layout.Window{Cursor: at, Height: m.win.Height}.Clamp(len(msg.roots))}
		}
	case tea.KeyPressMsg:
		m.problem = ""
		if m.drives.open() {
			return m.updateDrives(msg)
		}
		return m.key(msg)
	}
	return m, nil
}

// listed takes the listing of the directory still being browsed; a late one
// for a directory already left is dropped.
func (m Model) listed(msg listedMsg) Model {
	if msg.dir != m.dir {
		return m
	}
	if msg.err != nil {
		m.problem = "cannot read it: " + msg.err.Error()
	}
	m.entries = msg.entries
	if m.entries == nil {
		m.entries = []entry{}
	}
	m.win.Cursor = max(slices.IndexFunc(m.entries, func(e entry) bool { return e.name == m.reselect }), 0)
	m.reselect = ""
	m.win = m.win.Clamp(len(m.entries))
	return m
}

func (m Model) key(k tea.KeyPressMsg) (Model, tea.Cmd) {
	n, page := len(m.entries), max(1, m.win.Height-1)
	switch {
	case key.Matches(k, m.keys.Up):
		m.win = m.win.Move(-1, n, nil)
	case key.Matches(k, m.keys.Down):
		m.win = m.win.Move(1, n, nil)
	case key.Matches(k, m.keys.PageUp):
		m.win = m.win.Move(-page, n, nil)
	case key.Matches(k, m.keys.PageDown):
		m.win = m.win.Move(page, n, nil)
	case key.Matches(k, m.keys.Top):
		m.win = m.win.Move(-n, n, nil)
	case key.Matches(k, m.keys.Bottom):
		m.win = m.win.Move(n, n, nil)
	case key.Matches(k, m.keys.Back):
		if up := filepath.Dir(m.dir); up != m.dir {
			return m.browse(up, filepath.Base(m.dir))
		}
		id, vols := m.id, m.volumes
		if vols == nil {
			vols = listVolumes
		}
		return m, func() tea.Msg { return volumesMsg{id: id, roots: vols()} }
	case key.Matches(k, m.keys.Open):
		if e, ok := m.current(); ok && e.dir {
			return m.browse(filepath.Join(m.dir, e.name), "")
		}
	case key.Matches(k, m.keys.Pick):
		return m.pick()
	}
	return m, nil
}

// pick is enter: a directory opens unless WithDirs, a file of an allowed
// type is picked, and any other file stays and says why not.
func (m Model) pick() (Model, tea.Cmd) {
	e, ok := m.current()
	switch {
	case !ok:
	case e.dir && !m.dirs:
		return m.browse(filepath.Join(m.dir, e.name), "")
	case !e.dir && !m.allows(e.name):
		m.problem = fmt.Sprintf("%s is not %s", e.name, strings.Join(m.allowed, " or "))
	default:
		m.picked = filepath.Join(m.dir, e.name)
	}
	return m, nil
}

// updateDrives moves through the drive list; opening one browses it afresh.
func (m Model) updateDrives(k tea.KeyPressMsg) (Model, tea.Cmd) {
	n := len(m.drives.roots)
	switch {
	case key.Matches(k, m.keys.Up):
		m.drives.win = m.drives.win.Move(-1, n, nil)
	case key.Matches(k, m.keys.Down):
		m.drives.win = m.drives.win.Move(1, n, nil)
	case key.Matches(k, m.keys.Open, m.keys.Pick):
		return m.browse(m.drives.roots[m.drives.win.Cursor], "")
	}
	return m, nil
}

// Legend lists the keys the picker answers to now.
func (m Model) Legend() []legend.Entry {
	if m.drives.open() {
		return legend.FromBindings([]key.Binding{m.keys.Up, m.keys.Open})
	}
	return legend.FromBindings([]key.Binding{m.keys.Up, m.keys.Back, m.keys.Open, m.keys.Pick})
}

// View is the entries, or the drives, in a window that keeps the cursor in
// view and counts the rest below it; nothing until the first listing.
func (m Model) View(th theme.Resolved) string {
	st := DefaultStyles(th)
	if m.edit != nil {
		st = m.edit(st)
	}
	if m.drives.open() {
		return strings.Join(m.driveRows(&st), "\n")
	}
	return strings.Join(m.entryRows(&st), "\n")
}

func (m Model) entryRows(st *Styles) []string {
	if m.entries == nil {
		return nil
	}
	if len(m.entries) == 0 {
		return []string{st.Meta.Render("  nothing here")}
	}
	w := m.win.Clamp(len(m.entries))
	start, end, below := w.Bounds(len(m.entries))
	out := make([]string, 0, end-start+1)
	for i := start; i < end; i++ {
		e := m.entries[i]
		meta := m.columns(e)
		name := e.name
		if e.link != "" {
			name += " → " + e.link
		}
		disabled := !e.dir && !m.allows(e.name)
		switch {
		case i == w.Cursor && disabled:
			out = append(out, st.Disabled.Render(">"+meta+name))
		case i == w.Cursor:
			out = append(out, st.Mark.Render(">")+st.Cursor.Render(meta+name))
		default:
			out = append(out, " "+st.Meta.Render(meta)+m.entryStyle(e, disabled, st).Render(name))
		}
	}
	if below > 0 {
		out = append(out, draw.More(below, 2, st.Meta))
	}
	return out
}

// columns is what goes between the cursor's mark and the name: the
// permissions and the size, each followed by a space.
func (m Model) columns(e entry) string {
	var b strings.Builder
	if m.perms {
		b.WriteString(" " + e.mode.String())
	}
	if !m.noSizes {
		size := ""
		if !e.dir {
			size = humanBytes(e.size)
		}
		fmt.Fprintf(&b, "%*s", sizeWidth, size)
	}
	b.WriteString(" ")
	return b.String()
}

// entryStyle inks directories and links, and greys out what cannot be picked.
func (m Model) entryStyle(e entry, disabled bool, st *Styles) lipgloss.Style {
	switch {
	case e.dir, e.link != "":
		return st.Dir
	case disabled:
		return st.Disabled
	}
	return st.File
}

// driveRows draws the drive list the way entryRows draws a directory.
func (m Model) driveRows(st *Styles) []string {
	roots := m.drives.roots
	w := m.drives.win.Clamp(len(roots))
	start, end, below := w.Bounds(len(roots))
	var out []string
	for i := start; i < end; i++ {
		if i == w.Cursor {
			out = append(out, st.Mark.Render(">")+st.Cursor.Render(" "+roots[i]))
			continue
		}
		out = append(out, " "+st.Dir.Render(" "+roots[i]))
	}
	if below > 0 {
		out = append(out, draw.More(below, 2, st.Meta))
	}
	return out
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
