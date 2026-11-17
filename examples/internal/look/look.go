// Package look is the colour scheme the recordings are made in: the 16 colours
// of a plain terminal, as RGB. An effect can only mix a colour it has numbers
// for, and theme.Plain leaves them to the terminal.
package look

import (
	"image/color"
	"regexp"
	"strconv"

	"charm.land/lipgloss/v2"

	"github.com/lucasassuncao/bezel/theme"
)

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

// Text is the theme's text colour, for a scene or a light that wants a plain
// colour. Bold's foreground is no colour at all, and a scene reads that as black.
func Text(th theme.Resolved) color.Color { return lipgloss.Color(th.Colors.Text) }

var truecolor = regexp.MustCompile(`[34]8;2;(\d+);(\d+);(\d+)`)

// Brightest is the strongest channel of any colour s sets, 0 for none. A test
// asks it whether something drawn in a colour can be seen on black.
func Brightest(s string) int {
	best := 0
	for _, m := range truecolor.FindAllStringSubmatch(s, -1) {
		for _, c := range m[1:] {
			n, _ := strconv.Atoi(c)
			best = max(best, n)
		}
	}
	return best
}
