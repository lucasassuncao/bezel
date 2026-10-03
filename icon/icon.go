// Package icon is the set of marks that say what state something is in: done,
// failed, running, waiting. One set across an app, and across apps, so a check
// mark means the same thing in a list, a status line and plain CLI output.
package icon

import (
	"os"

	"github.com/lucasassuncao/bezel/theme"
)

// State is what a mark reports.
type State int

const (
	OK State = iota
	Fail
	Warn
	Info
	Pending
	Running
	Skipped
)

// Set is one glyph per State, plus the marks lists, hints, steppers and trees
// lean on. Branch, LastBranch and Pipe are three cells each, the label's space
// included, so a child lines up under its parent's label.
type Set struct {
	OK, Fail, Warn, Info, Pending, Running, Skipped string
	Bullet, Arrow, Line                             string
	Branch, LastBranch, Pipe                        string
}

// Unicode is the default set.
var Unicode = Set{
	OK: "✓", Fail: "✗", Warn: "!", Info: "i", Pending: "○", Running: "●", Skipped: "–",
	Bullet: "•", Arrow: "→", Line: "─",
	Branch: "├─ ", LastBranch: "└─ ", Pipe: "│  ",
}

// ASCII is for terminals and logs that cannot show the Unicode set.
var ASCII = Set{
	OK: "+", Fail: "x", Warn: "!", Info: "i", Pending: "o", Running: "*", Skipped: "-",
	Bullet: "*", Arrow: ">", Line: "-",
	Branch: "|- ", LastBranch: "`- ", Pipe: "|  ",
}

// Connector is the prefix that hangs an item off its parent: a branch, or the
// closing one when the item is the last of its siblings.
func (s Set) Connector(last bool) string {
	if last {
		return s.LastBranch
	}
	return s.Branch
}

// Indent is what an item's children carry under its connector: the pipe while
// siblings follow it, blank once it was the last.
func (s Set) Indent(last bool) string {
	if last {
		return "   "
	}
	return s.Pipe
}

// Default is ASCII when BEZEL_ASCII is set or TERM is "dumb", Unicode otherwise.
func Default() Set {
	if os.Getenv("BEZEL_ASCII") != "" || os.Getenv("TERM") == "dumb" {
		return ASCII
	}
	return Unicode
}

// Mark is the bare glyph for st.
func (s Set) Mark(st State) string {
	switch st {
	case OK:
		return s.OK
	case Fail:
		return s.Fail
	case Warn:
		return s.Warn
	case Info:
		return s.Info
	case Running:
		return s.Running
	case Skipped:
		return s.Skipped
	}
	return s.Pending
}

// Render is the glyph for st in the colour th gives that kind of state.
func (s Set) Render(st State, th theme.Resolved) string {
	mark := s.Mark(st)
	switch st {
	case OK:
		return th.Success.Render(mark)
	case Fail:
		return th.Danger.Render(mark)
	case Warn:
		return th.Warning.Render(mark)
	case Info:
		return th.Info.Render(mark)
	case Running:
		return th.Accent.Render(mark)
	}
	return th.Dim.Render(mark)
}
