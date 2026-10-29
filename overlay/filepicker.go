package overlay

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lucasassuncao/bezel/component/filepicker"
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
	title  string
	m      filepicker.Model
	onPick func(path string) tea.Cmd
}

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
		Up:       key.New("up", "k", "ctrl+p").WithHelp("↑/↓", "move"),
		Down:     key.New("down", "j", "ctrl+n"),
		PageUp:   key.New("pgup", "K"),
		PageDown: key.New("pgdown", "J"),
		Top:      key.New("home", "g"),
		Bottom:   key.New("end", "G"),
		Back:     key.New("left", "h", "backspace").WithHelp("←", "up"),
		Open:     key.New("right", "l").WithHelp("→", "open"),
		Pick:     key.New("enter").WithHelp("enter", "pick"),
	}
}

// pickerChrome is the lines a FilePicker spends around its entries: the
// title, the directory and the hint, and a gap after each of the first two.
const pickerChrome = 6

// volumes lists the drives for every picker; nil is the system's. A
// variable so tests can stand in their own.
var volumes func() []string

// NewFilePicker browses dir, files only, without hidden entries. onPick,
// which may be nil, gets the absolute path.
func NewFilePicker(title, dir string, onPick func(path string) tea.Cmd, th theme.Resolved) FilePicker {
	m := filepicker.New(dir).WithKeys(filepicker.Keys(DefaultFilePickerKeys()))
	return FilePicker{look: lookOf(th), title: title, m: m, onPick: onPick}
}

// WithKeys rebinds the picker.
func (p FilePicker) WithKeys(k FilePickerKeys) FilePicker {
	p.m = p.m.WithKeys(filepicker.Keys(k))
	return p
}

// Resize shows as many entries as fit in body.
func (p FilePicker) Resize(body layout.Rect) Overlay {
	p.m = p.m.SetHeight(p.fit(body))
	return p
}

// fit is how many entries body leaves room for, at least one.
func (p FilePicker) fit(body layout.Rect) int {
	return max(1, body.H-p.style().GetVerticalFrameSize()-pickerChrome)
}

// WithAllowedTypes limits the pick to names ending in one of exts, such as
// ".go"; other files stay listed but enter refuses them and says why.
func (p FilePicker) WithAllowedTypes(exts ...string) FilePicker {
	p.m = p.m.WithAllowedTypes(exts...)
	return p
}

// WithDirs lets enter pick a directory; → still opens one.
func (p FilePicker) WithDirs() FilePicker {
	p.m = p.m.WithDirs()
	return p
}

// WithHidden lists hidden entries too: dotfiles, and on Windows the ones
// marked hidden. Set it before Init.
func (p FilePicker) WithHidden() FilePicker {
	p.m = p.m.WithHidden()
	return p
}

// Init reads the first directory, answering through TopMsg. The picker must
// already be on top of the stack to receive it: push it first, as Open does.
func (p FilePicker) Init() tea.Cmd { return ToTop(p.m.Init()) }

// Open pushes the picker and then reads its directory, in that order.
func (p FilePicker) Open() tea.Cmd { return tea.Sequence(Open(p), p.Init()) }

// Dir is the directory being browsed.
func (p FilePicker) Dir() string { return p.m.Dir() }

// Highlighted is the path under the cursor, or "" in an empty directory.
func (p FilePicker) Highlighted() string { return p.m.Highlighted() }

func (p FilePicker) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if m, ok := msg.(TopMsg); ok {
		msg = m.Msg
	}
	var cmd tea.Cmd
	p.m, cmd = p.m.WithVolumes(volumes).Update(msg)
	path, picked := p.m.Picked()
	switch {
	case !picked:
		if cmd != nil {
			cmd = ToTop(cmd)
		}
		return p, cmd
	case p.onPick == nil:
		return p, Close()
	}
	return p, tea.Batch(p.onPick(path), Close())
}

func (p FilePicker) Legend() []legend.Entry {
	return append(p.m.Legend(), legend.New("esc", "cancel"))
}

func (p FilePicker) View(body layout.Rect) string {
	ink, hint := accent(p.style()), p.hint().Text
	where := tail(p.Dir(), Inner(body))
	if !p.m.Browsing() {
		where = "drives"
	}
	rows := p.m.SetHeight(min(p.fit(body), p.m.Height())).WithStyles(func(filepicker.Styles) filepicker.Styles {
		return filepicker.Styles{Cursor: ink.Bold(true), Mark: ink, Dir: ink, File: lipgloss.NewStyle(), Disabled: hint, Meta: hint}
	}).View(*p.styles())
	lines := []string{ink.Bold(true).Render(p.title), hint.Render(where), ""}
	if rows != "" {
		lines = append(lines, rows)
	}
	h := legend.HintLine(p.Legend(), p.hint())
	if problem := p.m.Problem(); problem != "" {
		h = hint.Render(problem) + "   " + h
	}
	return Box(append(lines, "", h), body, p.style())
}

// tail cuts a path from the left: its end is what tells directories apart.
func tail(path string, width int) string {
	if over := draw.Width(path) - width; over > 0 {
		return ansi.TruncateLeft(path, over+1, "…")
	}
	return path
}
