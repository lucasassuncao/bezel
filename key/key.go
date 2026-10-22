// Package key binds keys to what they do: the keys that fire a binding, the
// help the legend shows for it, and whether it is on. Its API is bubbles'
// key package's, so moving a program over is a change of import path.
package key

import "fmt"

// Binding is a set of keys and their help. A Binding with no keys matches
// nothing and reads as not enabled.
type Binding struct {
	keys     []string
	help     Help
	disabled bool
}

// Help is what the legend prints for a binding: its key and what it does.
type Help struct {
	Key  string
	Desc string
}

// BindingOpt sets one part of a binding being built.
type BindingOpt func(*Binding)

// NewBinding is a binding built from opts.
func NewBinding(opts ...BindingOpt) Binding {
	var b Binding
	for _, opt := range opts {
		opt(&b)
	}
	return b
}

// WithKeys sets the keys, as tea.KeyPressMsg.String() spells them: "ctrl+s",
// "up", "?".
func WithKeys(keys ...string) BindingOpt {
	return func(b *Binding) { b.keys = keys }
}

// WithHelp sets the key and the description the legend prints.
func WithHelp(key, desc string) BindingOpt {
	return func(b *Binding) { b.help = Help{Key: key, Desc: desc} }
}

// WithDisabled builds the binding switched off.
func WithDisabled() BindingOpt {
	return func(b *Binding) { b.disabled = true }
}

// Keys are the keys that fire b.
func (b Binding) Keys() []string { return b.keys }

// Help is what the legend prints for b.
func (b Binding) Help() Help { return b.help }

// Enabled reports whether b can fire: it has keys and is not switched off.
func (b Binding) Enabled() bool { return !b.disabled && b.keys != nil }

// SetKeys replaces b's keys.
func (b *Binding) SetKeys(keys ...string) { b.keys = keys }

// SetHelp replaces b's help.
func (b *Binding) SetHelp(key, desc string) { b.help = Help{Key: key, Desc: desc} }

// SetEnabled switches b on or off; its keys and help stay.
func (b *Binding) SetEnabled(v bool) { b.disabled = !v }

// Unbind drops b's keys and help, so it fires on nothing and shows nowhere.
func (b *Binding) Unbind() {
	b.keys = nil
	b.help = Help{}
}

// Matches reports whether k is one of the keys of an enabled binding in b.
func Matches[Key fmt.Stringer](k Key, b ...Binding) bool {
	pressed := k.String()
	for _, binding := range b {
		if !binding.Enabled() {
			continue
		}
		for _, v := range binding.keys {
			if pressed == v {
				return true
			}
		}
	}
	return false
}
