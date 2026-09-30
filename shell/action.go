package shell

import (
	"errors"
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/legend"
)

// Action is one key for the legend, the dispatch and the palette. Run runs
// inside Shell.Update, so it must not change the app's copy of the shell (the
// Update's result would overwrite it): emit a message instead.
type Action interface {
	Entry() legend.Entry
	Run(ctx ActionContext) tea.Cmd
}

// ActionContext is what an action runs with. Arg is the text after the name
// when it ran from the palette (":goto kv/app"), empty otherwise.
type ActionContext struct {
	Context
	Arg string
}

// kind tells the prebuilt actions apart, for the dispatch's own work.
type kind int

const (
	kindHelp kind = iota
	kindCommands
	kindChangeTab
	kindChangePane
	kindMove
	kindScroll
	kindQuit
	kindCustom
)

type action struct {
	kind  kind
	keys  []string
	label string // the key as the legend prints it
	desc  string
	name  string // the palette name; empty keeps it out of the palette
	arg   string
	hint  string // a key the app handles itself, printed by the palette only
	needs Capability
	group int
	when  func(Context) bool
	run   func(ActionContext) tea.Cmd

	display  bool // printed only: the focused component handles the key
	override bool // a prebuilt that runs run instead of the shell's own work
}

// Option adjusts an action at construction.
type Option func(*action)

func (a *action) Entry() legend.Entry {
	b := key.NewBinding(key.WithKeys(a.keys...), key.WithHelp(a.label, a.desc))
	return legend.Entry{Binding: b, Needs: a.needs, Group: a.group}
}

func (a *action) Run(ctx ActionContext) tea.Cmd {
	if a.run == nil {
		return nil
	}
	return a.run(ctx)
}

// displayOnly actions are printed but never handled: the focused component
// is what moves or scrolls.
func (a *action) displayOnly() bool { return a.display || a.kind == kindMove || a.kind == kindScroll }

func build(a *action, opts []Option) Action {
	for _, o := range opts {
		o(a)
	}
	return a
}

// Help opens the key list for where the user is.
func Help(opts ...Option) Action {
	return build(&action{kind: kindHelp, keys: []string{"?"}, label: "?", desc: "help", name: "help"}, opts)
}

// Commands opens the command palette.
func Commands(opts ...Option) Action {
	return build(&action{kind: kindCommands, keys: []string{":"}, label: ":", desc: "commands"}, opts)
}

// ChangeTab cycles the tabs; the first key walks forward, the rest back.
func ChangeTab(opts ...Option) Action {
	return build(&action{kind: kindChangeTab, keys: []string{"tab", "shift+tab"}, label: "tab", desc: "change tab"}, opts)
}

// ChangePane cycles the panes on screen, for an app without tabs.
func ChangePane(opts ...Option) Action {
	return build(&action{kind: kindChangePane, keys: []string{"tab", "shift+tab"}, label: "tab", desc: "change pane"}, opts)
}

// Move says the arrows move the cursor. Display only.
func Move(opts ...Option) Action {
	return build(&action{kind: kindMove, keys: []string{"up", "down"}, label: "↑/↓", desc: "move"}, opts)
}

// Scroll says the arrows scroll a document pane. Display only.
func Scroll(opts ...Option) Action {
	return build(&action{kind: kindScroll, keys: []string{"up", "down"}, label: "↑/↓", desc: "scroll"}, opts)
}

// Quit leaves the program.
func Quit(opts ...Option) Action {
	quit := func(ActionContext) tea.Cmd { return tea.Quit }
	return build(&action{kind: kindQuit, keys: []string{"q"}, label: "q", desc: "quit", name: "quit", run: quit}, opts)
}

// Custom is an app key: keys is bound literally and printed as is. Use
// WithKey for several keys under one label ("←/→").
func Custom(keys, desc string, run func(ActionContext) tea.Cmd, opts ...Option) Action {
	return build(&action{kind: kindCustom, keys: []string{keys}, label: keys, desc: desc, run: run}, opts)
}

// Command is palette-only: no key, reached as :name.
func Command(name, desc string, run func(ActionContext) tea.Cmd, opts ...Option) Action {
	return build(&action{kind: kindCustom, desc: desc, name: name, run: run}, opts)
}

// Send is a run that emits msg, for value models: the app handles msg in its
// own Update instead of closing over a copy of itself.
func Send(msg tea.Msg) func(ActionContext) tea.Cmd {
	return func(ActionContext) tea.Cmd { return func() tea.Msg { return msg } }
}

// Named puts the action in the palette as :name.
func Named(name string) Option { return func(a *action) { a.name = name } }

// Arg makes the palette show ":name <placeholder>" and pass what follows.
func Arg(placeholder string) Option { return func(a *action) { a.arg = placeholder } }

// Needs drops the action where the session refuses c.
func Needs(c Capability) Option { return func(a *action) { a.needs = c } }

// Group puts the action in legend row group n: each group starts a new row
// while Config.LegendLines allows. Ungrouped actions are group 0.
func Group(n int) Option { return func(a *action) { a.group = n } }

// When keeps the action only while f reports true.
func When(f func(Context) bool) Option { return func(a *action) { a.when = f } }

// WithKey rebinds the action. A prebuilt, or a Command that had no key,
// prints the first key; a custom keeps the label it was given.
func WithKey(keys ...string) Option {
	return func(a *action) {
		a.keys = keys
		if len(keys) > 0 && (a.kind != kindCustom || a.label == "") {
			a.label = keys[0]
		}
	}
}

// KeyHint prints k beside the command in the palette without binding it or
// listing it in the legend: for an app that still routes that key itself.
func KeyHint(k string) Option { return func(a *action) { a.hint = k } }

// DisplayOnly prints the action but never handles its key, so the key
// reaches the focused component that owns it, as Move does for the arrows.
func DisplayOnly() Option { return func(a *action) { a.display = true } }

// RunWith replaces what the action does, keeping its key, text and rank: a
// Quit that asks first, a ChangePane that moves the app's own focus.
func RunWith(run func(ActionContext) tea.Cmd) Option {
	return func(a *action) { a.run, a.override = run, true }
}

// kindOf is an action's kind; foreign Action types count as custom.
func kindOf(a Action) kind {
	if ac, ok := a.(*action); ok {
		return ac.kind
	}
	return kindCustom
}

// Check reports every key two actions share, display-only ones included.
// Apps call it in a test over each screen's set; at runtime the first match
// runs.
func Check(actions ...Action) error {
	owner := map[string]string{}
	var errs []error
	for _, a := range actions {
		e := a.Entry()
		for _, k := range e.Keys() {
			if prev, taken := owner[k]; taken {
				errs = append(errs, fmt.Errorf("key %q: %q and %q", k, prev, e.Help().Desc))
				continue
			}
			owner[k] = e.Help().Desc
		}
	}
	return errors.Join(errs...)
}
