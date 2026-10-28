// Package logger is a slog.Handler that prints for people: the time, the level
// and the message on one line, and every attribute under it as a child of a
// tree, wrapped at the terminal's width without losing a character.
package logger

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"

	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/theme"
)

// LevelTrace is below slog.LevelDebug, for the detail only a bug hunt wants.
const LevelTrace = slog.LevelDebug - 4

// DefaultTimeFormat is the time each line starts with; WithTimeFormat("") drops it.
const DefaultTimeFormat = "2006-01-02 15:04:05"

// Styles is how each part of a line is drawn.
type Styles struct {
	Time, Caller                    lipgloss.Style
	Trace, Debug, Info, Warn, Error lipgloss.Style
	Key, Equals, Branch             lipgloss.Style
}

// DefaultStyles is the levels in the theme's info, warning and danger colours,
// the keys in its accent, and the time and the tree in its quieter ones.
func DefaultStyles(th theme.Resolved) Styles {
	return Styles{
		Time:   th.Dim,
		Caller: th.Muted,
		Trace:  th.Muted.Bold(true),
		Debug:  th.Dim.Bold(true),
		Info:   th.Info.Bold(true),
		Warn:   th.Warning.Bold(true),
		Error:  th.Danger.Bold(true),
		Key:    th.Accent,
		Equals: th.Muted,
		Branch: th.Muted,
	}
}

// Handler is the slog.Handler; slog.New(logger.New(os.Stderr, th)) is a logger.
// Every copy writes through the same lock, so lines never interleave.
type Handler struct {
	out     *output
	th      *theme.Resolved // shared by every copy and never written
	st      *Styles
	edit    func(Styles) Styles
	level   slog.Leveler
	width   int
	timeFmt string
	caller  bool
	color   bool
	icons   icon.Set
	scopes  []scope // [0] is the root; WithGroup opens one more
}

type output struct {
	mu  sync.Mutex
	w   io.Writer
	fd  uintptr
	tty bool
}

// scope is a group opened by WithGroup and the attributes added inside it.
type scope struct {
	name  string
	attrs []slog.Attr
}

// New prints to w in th's colours from slog.LevelInfo up. On a terminal it
// draws in colour and wraps at the terminal's width; anywhere else it writes
// plain text, one line per value.
func New(w io.Writer, th theme.Resolved) Handler {
	o := &output{w: w}
	if f, ok := w.(interface{ Fd() uintptr }); ok && term.IsTerminal(f.Fd()) {
		o.fd, o.tty = f.Fd(), true
	}
	st := DefaultStyles(th)
	return Handler{
		out: o, th: &th, st: &st,
		level: slog.LevelInfo, timeFmt: DefaultTimeFormat, color: o.tty,
		icons: icon.FromEnv(), scopes: []scope{{}},
	}
}

// WithLevel sets the lowest level printed; a *slog.LevelVar changes it live.
func (h Handler) WithLevel(l slog.Leveler) Handler {
	h.level = l
	return h
}

// WithWidth wraps at n cells instead of the terminal's width; below zero never wraps.
func (h Handler) WithWidth(n int) Handler {
	h.width = n
	return h
}

// WithTimeFormat lays out the time with a time.Format layout; "" leaves it out.
func (h Handler) WithTimeFormat(layout string) Handler {
	h.timeFmt = layout
	return h
}

// WithCaller ends the message with the file and line that logged it.
func (h Handler) WithCaller(on bool) Handler {
	h.caller = on
	return h
}

// WithColor turns colour on or off, whatever w is. Off also strips the colour
// a message carries, so a log file stays plain text.
func (h Handler) WithColor(on bool) Handler {
	h.color = on
	return h
}

// WithIcons draws the tree with set's branches, ASCII for a dumb terminal.
func (h Handler) WithIcons(set icon.Set) Handler {
	h.icons = set
	return h
}

// WithStyles edits the default styles; a second call replaces the first edit.
func (h Handler) WithStyles(edit func(Styles) Styles) Handler {
	h.edit = edit
	st := edit(DefaultStyles(*h.th))
	h.st = &st
	return h
}

// Enabled reports whether l is at or above the level set.
func (h Handler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

// WithAttrs adds attrs to every record, inside the groups open so far.
func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	h.scopes = append([]scope(nil), h.scopes...)
	top := &h.scopes[len(h.scopes)-1]
	top.attrs = append(append([]slog.Attr(nil), top.attrs...), attrs...)
	return h
}

// WithGroup nests every attribute added from now on under name.
func (h Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h.scopes = append(append([]scope(nil), h.scopes...), scope{name: name})
	return h
}

// Handle writes r: its head line, then its attributes as a tree.
func (h Handler) Handle(_ context.Context, r slog.Record) error {
	width := h.lineWidth()
	var b strings.Builder
	indent := h.head(&b, r, width)

	attrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	h.tree(&b, h.fold(attrs), strings.Repeat(" ", indent), indent, width)

	h.out.mu.Lock()
	defer h.out.mu.Unlock()
	_, err := io.WriteString(h.out.w, b.String())
	return err
}

// head writes the time, level and message, and returns the cells before the
// message, where the tree and the message's own continuation lines start.
func (h Handler) head(b *strings.Builder, r slog.Record, width int) int {
	indent := 0
	if h.timeFmt != "" && !r.Time.IsZero() {
		t := r.Time.Format(h.timeFmt)
		b.WriteString(h.paint(h.st.Time, t) + " ")
		indent += ansi.StringWidth(t) + 1
	}
	name, st := h.label(r.Level)
	b.WriteString(h.paint(st, name) + strings.Repeat(" ", 6-len(name)))
	indent += 6

	msg := r.Message
	if !h.color {
		msg = ansi.Strip(msg)
	}
	if h.caller && r.PC != 0 {
		f, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
		msg += " " + h.paint(h.st.Caller, filepath.Base(f.File)+":"+strconv.Itoa(f.Line))
	}
	if width > indent {
		msg = ansi.Wrap(msg, width-indent, "")
	}
	b.WriteString(strings.ReplaceAll(msg, "\n", "\n"+strings.Repeat(" ", indent)) + "\n")
	return indent
}

// label is the level's name and style: the nearest slog level at or below it.
func (h Handler) label(l slog.Level) (string, lipgloss.Style) {
	switch {
	case l < slog.LevelDebug:
		return "TRACE", h.st.Trace
	case l < slog.LevelInfo:
		return "DEBUG", h.st.Debug
	case l < slog.LevelWarn:
		return "INFO", h.st.Info
	case l < slog.LevelError:
		return "WARN", h.st.Warn
	}
	return "ERROR", h.st.Error
}

// fold puts the record's attributes inside the open groups and those inside
// their parents, the way slog nests them.
func (h Handler) fold(attrs []slog.Attr) []slog.Attr {
	for i := len(h.scopes) - 1; i >= 0; i-- {
		s := h.scopes[i]
		attrs = clean(append(append([]slog.Attr(nil), s.attrs...), attrs...))
		if i > 0 {
			if len(attrs) == 0 {
				continue
			}
			attrs = []slog.Attr{{Key: s.name, Value: slog.GroupValue(attrs...)}}
		}
	}
	return attrs
}

// clean resolves each value and applies slog's rules: an empty attribute and
// an empty group are dropped, a group with no key gives its attributes to its parent.
func clean(attrs []slog.Attr) []slog.Attr {
	out := make([]slog.Attr, 0, len(attrs))
	for _, a := range attrs {
		a.Value = a.Value.Resolve()
		if a.Equal(slog.Attr{}) {
			continue
		}
		if a.Value.Kind() != slog.KindGroup {
			out = append(out, a)
			continue
		}
		inner := clean(a.Value.Group())
		switch {
		case len(inner) == 0:
		case a.Key == "":
			out = append(out, inner...)
		default:
			out = append(out, slog.Attr{Key: a.Key, Value: slog.GroupValue(inner...)})
		}
	}
	return out
}

// tree writes attrs as children of the line above, lead (leadW cells) before
// each. A value too long for the width goes on under its own start.
func (h Handler) tree(b *strings.Builder, attrs []slog.Attr, lead string, leadW, width int) {
	for i, a := range attrs {
		last := i == len(attrs)-1
		branch := lead + h.paint(h.st.Branch, h.icons.Connector(last))
		under := lead + h.paint(h.st.Branch, h.icons.Indent(last))
		underW := leadW + ansi.StringWidth(h.icons.Indent(last))
		if a.Value.Kind() == slog.KindGroup {
			b.WriteString(branch + h.paint(h.st.Key, a.Key) + "\n")
			h.tree(b, a.Value.Group(), under, underW, width)
			continue
		}
		keyW := ansi.StringWidth(a.Key) + 1
		room := 0
		if width > 0 {
			room = width - underW - keyW
		}
		lines := wrap(h.value(a.Value), room)
		b.WriteString(branch + h.paint(h.st.Key, a.Key) + h.paint(h.st.Equals, "=") + lines[0] + "\n")
		pad := under + strings.Repeat(" ", keyW)
		for _, l := range lines[1:] {
			b.WriteString(pad + l + "\n")
		}
	}
}

// value is v as text, without the colour it may carry.
func (h Handler) value(v slog.Value) string {
	if v.Kind() == slog.KindTime {
		layout := h.timeFmt
		if layout == "" {
			layout = time.RFC3339
		}
		return v.Time().Format(layout)
	}
	return ansi.Strip(v.String())
}

func (h Handler) paint(st lipgloss.Style, s string) string {
	if !h.color {
		return s
	}
	return st.Render(s)
}

// lineWidth is the width set, else the terminal's; 0 means no wrapping.
func (h Handler) lineWidth() int {
	if h.width != 0 {
		return max(h.width, 0)
	}
	if h.out.tty {
		if w, _, err := term.GetSize(h.out.fd); err == nil {
			return w
		}
	}
	return 0
}

// ReplaceAttr names LevelTrace "TRACE" in slog's own handlers, which print
// "DEBUG-4": set it as slog.HandlerOptions.ReplaceAttr for --json.
func ReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	if len(groups) == 0 && a.Key == slog.LevelKey {
		if l, ok := a.Value.Any().(slog.Level); ok && l < slog.LevelDebug {
			return slog.String(slog.LevelKey, "TRACE")
		}
	}
	return a
}
