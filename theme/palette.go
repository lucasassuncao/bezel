// Package theme is the palette and the lipgloss styles every bezel app draws with.
package theme

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
)

// Colors names what a colour is for, not where one app happens to use it.
// Each value is a lipgloss colour string: a hex value ("#7C3AED") or an ANSI
// code ("63"). Empty means the adaptive default for that role.
type Colors struct {
	// Background is the whole screen, behind every pane; it decides light or
	// dark itself. Empty leaves the terminal's own.
	Background string
	Accent     string // focused borders, section headings, legend keys
	Selection  string // the cursor row, the focused panel's title
	Border     string // unfocused borders, status and hint text
	Dim        string // secondary text, items not yet present
	Text       string // body text
	Success    string // present items, healthy results
	Warning    string // drafts, soft failures
	Danger     string // errors, unknown items
	Info       string // counts, badges, neutral highlights
	OnAccent   string // text drawn on a filled accent, success, warning or danger
	// Series tells apart things of one kind, such as a dashboard's panels;
	// Ramp climbs from low to high, as a meter or a graph fills.
	Series [6]string
	Ramp   [3]string
}

// Theme layers an appearance: Base is an optional preset (nil → adaptive
// default) that may have a Base of its own; Colors override every layer
// under them.
type Theme struct {
	Base   *Theme
	Colors Colors
}

// adaptive is a default colour with one value for a light terminal and one
// for a dark one; which applies is known only once the terminal answers.
type adaptive struct{ light, dark string }

var defaults = struct {
	accent, selection, border, dim, text, success, warning, danger, info, onAccent adaptive
	series                                                                         [6]adaptive
}{
	accent:    adaptive{"#5f00af", "#c39bff"},
	selection: adaptive{"#5f00af", "#c39bff"},
	border:    adaptive{"#d0d0d0", "#4e4e4e"},
	dim:       adaptive{"#6c6c6c", "#8a8a8a"},
	text:      adaptive{"#1c1c1c", "#e4e4e4"},
	success:   adaptive{"#005f00", "#87d787"},
	warning:   adaptive{"#af5f00", "#ffb86c"},
	danger:    adaptive{"#af0000", "#ff6b6b"},
	info:      adaptive{"#5f00af", "#c39bff"},
	onAccent:  adaptive{"#ffffff", "#1c1c1c"},
	series: [6]adaptive{
		{"#2e7d32", "#87d787"}, {"#00838f", "#5fd7d7"}, {"#d35400", "#ffaf5f"},
		{"#1565c0", "#87afff"}, {"#9a7b00", "#ffd75f"}, {"#8e24aa", "#d787d7"},
	},
}

// defaultColors is the adaptive default resolved for one background.
func defaultColors(dark bool) Colors {
	pick := func(a adaptive) string {
		if dark {
			return a.dark
		}
		return a.light
	}
	d := defaults
	c := Colors{
		Accent: pick(d.accent), Selection: pick(d.selection), Border: pick(d.border),
		Dim: pick(d.dim), Text: pick(d.text), Success: pick(d.success),
		Warning: pick(d.warning), Danger: pick(d.danger), Info: pick(d.info),
		OnAccent: pick(d.onAccent),
	}
	for i, a := range d.series {
		c.Series[i] = pick(a)
	}
	return c
}

// resolveColors merges t into concrete Colors over the adaptive default for
// a dark or light terminal, or for its own Background when it has one. A theme
// that sets Accent but not Selection or Info uses its accent for them.
func resolveColors(t Theme, dark bool) Colors {
	var c Colors
	for _, layer := range t.chain() {
		c = mergeColors(c, layer.Colors)
	}
	if c.Background != "" {
		dark = isDark(c.Background)
	}
	if c.Accent != "" {
		c = mergeColors(Colors{Selection: c.Accent, Info: c.Accent}, c)
	}
	c = mergeColors(defaultColors(dark), c)
	states := [3]string{c.Success, c.Warning, c.Danger}
	for i := range c.Ramp {
		if c.Ramp[i] == "" {
			c.Ramp[i] = states[i]
		}
	}
	return c
}

// chain is t and its Bases, deepest first. A Base met twice ends the walk,
// so a cycle resolves instead of recursing forever.
func (t Theme) chain() []Theme {
	layers := []Theme{t}
	seen := map[*Theme]bool{}
	for b := t.Base; b != nil && !seen[b]; b = b.Base {
		seen[b] = true
		layers = append(layers, *b)
	}
	slices.Reverse(layers)
	return layers
}

func mergeColors(base, over Colors) Colors {
	set := func(dst *string, v string) {
		if v != "" {
			*dst = v
		}
	}
	set(&base.Background, over.Background)
	set(&base.Accent, over.Accent)
	set(&base.Selection, over.Selection)
	set(&base.Border, over.Border)
	set(&base.Dim, over.Dim)
	set(&base.Text, over.Text)
	set(&base.Success, over.Success)
	set(&base.Warning, over.Warning)
	set(&base.Danger, over.Danger)
	set(&base.Info, over.Info)
	set(&base.OnAccent, over.OnAccent)
	for i := range base.Series {
		set(&base.Series[i], over.Series[i])
	}
	for i := range base.Ramp {
		set(&base.Ramp[i], over.Ramp[i])
	}
	return base
}

// namedTheme pairs a theme's CLI name with the preset that builds it, the
// unit All() and Categories() both project, so the two never drift apart.
type namedTheme struct {
	name  string
	theme func() Theme
}

// themeRegistry is the one place a built-in theme is listed, with its CLI
// name and category.
var themeRegistry = []struct {
	category string
	themes   []namedTheme
}{
	{"Miscellaneous", []namedTheme{
		{"default", Default}, {"plain", Plain}, {"terminal", Terminal}, {"mint", Mint},
	}},
}

// allNamedThemes flattens themeRegistry into a single slice, dropping the
// category grouping - the projection All() needs.
func allNamedThemes() []namedTheme {
	var out []namedTheme
	for _, cat := range themeRegistry {
		out = append(out, cat.themes...)
	}
	return out
}

// namesOf extracts just the names from a slice of namedTheme, in order - the
// projection Categories() needs for each group.
func namesOf(nts []namedTheme) []string {
	names := make([]string, len(nts))
	for i, nt := range nts {
		names[i] = nt.name
	}
	return names
}

// All returns all built-in theme presets keyed by their CLI name.
// Useful for --theme flag validation and --list-themes output in host CLIs.
func All() map[string]Theme {
	m := make(map[string]Theme)
	for _, nt := range allNamedThemes() {
		m[nt.name] = nt.theme()
	}
	return m
}

// Lookup is the built-in theme called name, in any case; empty is Default.
// The error names what was asked for; the caller adds where its list is.
func Lookup(name string) (Theme, error) {
	if name == "" {
		return Default(), nil
	}
	t, ok := All()[strings.ToLower(name)]
	if !ok {
		return Theme{}, fmt.Errorf("unknown theme %q", name)
	}
	return t, nil
}

// Category groups a set of related built-in theme names for display purposes -
// e.g. a --list-themes command that wants headings instead of one flat list.
type Category struct {
	Name   string
	Themes []string // names, in display order; each is also a key in All()
}

// Categories returns the built-in themes grouped for display. Every theme
// in All() belongs to exactly one category.
func Categories() []Category {
	cats := make([]Category, len(themeRegistry))
	for i, cat := range themeRegistry {
		cats[i] = Category{Name: cat.category, Themes: namesOf(cat.themes)}
	}
	return cats
}

// Default leaves every colour to the adaptive default, light or dark as the
// terminal is. Each preset builds a fresh Theme: a change never reaches the next.
func Default() Theme { return Theme{} }

// Plain is the terminal's own 16 ANSI colours.
func Plain() Theme {
	return Theme{Colors: Colors{
		Accent:    "4", // ANSI blue
		Selection: "6", // ANSI cyan
		Border:    "8", // ANSI dark grey
		Dim:       "8", // ANSI dark grey
		Success:   "2", // ANSI green
		Danger:    "1", // ANSI red
		Warning:   "3", // ANSI yellow
		Info:      "6", // ANSI cyan
		Series:    [6]string{"2", "6", "3", "4", "5", "1"},
	}}
}

// Terminal is the 256-colour palette most terminals share.
func Terminal() Theme {
	return Theme{Colors: Colors{
		Accent:    "13",  // bright magenta: selection, active tab
		Selection: "252", // near-white text on the cursor row
		Border:    "240", // grey
		Dim:       "240",
		Success:   "2",
		Danger:    "9",
		Warning:   "3",
		Info:      "6",
		Series:    [6]string{"10", "14", "11", "12", "13", "9"},
	}}
}

// Mint is a green accent on a dark background.
func Mint() Theme {
	return Theme{Colors: Colors{
		Accent: "#3EB489", Selection: "#98DFAF", Border: "#4A7B6F", Dim: "#2E4F46", Success: "#2ECC71", Danger: "#E74C3C", Background: "#0B1B15",
	}}
}

// isDark reports a colour below half lightness, as a terminal judges its own.
func isDark(c string) bool {
	r, g, b, _ := lipgloss.Color(c).RGBA()
	return max(r, g, b)+min(r, g, b) < 0xffff
}

// DarkTerminal asks the terminal on stdin/stdout whether its background is
// dark, for code that resolves a theme outside a running bubbletea program.
// A redirected stdin or stdout is not asked, and reads as dark.
func DarkTerminal() bool { return darkTerminal(os.Stdin, os.Stdout) }

// darkTerminal asks only when both ends are terminals: with either one
// redirected, lipgloss reads the Windows console, which may never answer.
func darkTerminal(in, out *os.File) bool {
	if !term.IsTerminal(in.Fd()) || !term.IsTerminal(out.Fd()) {
		return true
	}
	return lipgloss.HasDarkBackground(in, out)
}
