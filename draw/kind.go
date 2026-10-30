package draw

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Kind is what a notice reports. Its colour is fixed, the same under every
// theme, so a warning reads as one whatever the app looks like.
type Kind int

const (
	Info Kind = iota
	Success
	Warning
	Danger
)

// Color is k's fixed colour; a Kind out of range is Info's blue.
func (k Kind) Color() color.Color {
	switch k {
	case Success:
		return lipgloss.Color("#22C55E")
	case Warning:
		return lipgloss.Color("#EAB308")
	case Danger:
		return lipgloss.Color("#EF4444")
	default:
		return lipgloss.Color("#3B82F6")
	}
}
