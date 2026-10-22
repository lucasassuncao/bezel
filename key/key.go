// Package key binds keys to what they do: the keys that fire a binding, the
// help the legend shows for it, and whether it is on. A Binding is a value,
// built with NewBinding and its options and switched off by its Disabled field.
package key

import "fmt"

// Binding is a set of keys and their help. A Binding with no keys matches
// nothing and reads as not enabled. Disabled switches it off, keys and help
// kept: k.Save.Disabled = true. There are no setters to call and lose.
type Binding struct {
	Disabled bool
	keys     []string
	help     Help
}

// Help is what the legend prints for a binding: its key and what it does.
type Help struct {
	Key  string
	Desc string
}

// BindingOpt sets one part of a binding being built.
type BindingOpt func(*Binding)

// NewBinding is a binding built from opts. Its keys or help change by
// building another: a binding read elsewhere never changes under its reader.
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
	return func(b *Binding) { b.Disabled = true }
}

// Keys are the keys that fire b.
func (b Binding) Keys() []string { return b.keys }

// Help is what the legend prints for b.
func (b Binding) Help() Help { return b.help }

// Enabled reports whether b can fire: it has keys and is not switched off.
func (b Binding) Enabled() bool { return !b.Disabled && b.keys != nil }

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
