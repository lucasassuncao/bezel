// Package look is the colour scheme the recordings are made in: the 16 colours
// of a plain terminal, as RGB. An effect can only mix a colour it has numbers
// for, and theme.Plain leaves them to the terminal.
package look

import "github.com/lucasassuncao/bezel/theme"

// Plain is theme.Plain with each ANSI colour as the RGB the recordings'
// terminal shows for it, and no background: the terminal's own black stays.
func Plain() theme.Theme {
	return theme.Theme{Colors: theme.Colors{
		Accent:    "#0037DA", // ANSI blue
		Selection: "#3A96DD", // ANSI cyan
		Border:    "#767676", // ANSI bright black
		Dim:       "#767676",
		Text:      "#CCCCCC", // ANSI white
		Success:   "#13A10E", // ANSI green
		Danger:    "#C50F1F", // ANSI red
		Warning:   "#C19C00", // ANSI yellow
		Info:      "#3A96DD", // ANSI cyan
		Series:    [6]string{"#13A10E", "#3A96DD", "#C19C00", "#0037DA", "#881798", "#C50F1F"},
	}}
}
