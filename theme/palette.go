// Package theme is the palette and the lipgloss styles every bezel app draws with.
package theme

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

// Colors names what a colour is for, not where one app happens to use it.
// Each value is a lipgloss colour string: a hex value ("#7C3AED") or an ANSI
// code ("63"). Empty means the adaptive default for that role.
type Colors struct {
	Accent    string // focused borders, section headings, legend keys
	Selection string // the cursor row, the focused panel's title
	Border    string // unfocused borders, status and hint text
	Dim       string // secondary text, items not yet present
	Text      string // body text
	Success   string // present items, healthy results
	Warning   string // drafts, soft failures
	Danger    string // errors, unknown items
	Info      string // counts, badges, neutral highlights
	OnAccent  string // text drawn on a filled accent, success, warning or danger
	// Series tells apart things of one kind, such as a dashboard's panels;
	// Ramp climbs from low to high, as a meter or a graph fills.
	Series [6]string
	Ramp   [3]string
}

// Styles holds optional per-element lipgloss overrides. Nil fields are ignored
// during theme resolution and the default derived from Colors is used instead.
type Styles struct {
	Cursor *lipgloss.Style
	Muted  *lipgloss.Style
	Danger *lipgloss.Style
}

// Theme layers an appearance: Base is an optional preset (nil → adaptive
// default) that may have a Base of its own; Colors and Styles override every
// layer under them.
type Theme struct {
	Base   *Theme
	Colors Colors
	Styles Styles
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

// ResolveColors merges t into concrete Colors over the adaptive default for
// a dark or light terminal. A theme that sets Accent but not Selection or
// Info uses its accent for them; one without a Ramp climbs its states.
func ResolveColors(t Theme, dark bool) Colors {
	var c Colors
	for _, layer := range t.chain() {
		c = mergeColors(c, layer.Colors)
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

// namedTheme pairs a theme's CLI name with its value - the single entry unit
// shared by both All() and Categories(), so the two can never drift apart.
type namedTheme struct {
	name  string
	theme *Theme // the preset variable itself, so a change to it shows here
}

// themeRegistry is the one place a built-in theme is listed, with its CLI
// name and category. All() and Categories() both project this slice, so a
// theme added here shows up in both.
var themeRegistry = []struct {
	category string
	themes   []namedTheme
}{
	{"Miscellaneous", []namedTheme{
		{"default", &Default},
		{"plain", &Plain},
		{"terminal", &Terminal},
		{"gruvbox", &Gruvbox},
	}},
	{"Fruit", []namedTheme{
		{"banana", &Banana}, {"mint", &Mint}, {"strawberry", &Strawberry},
		{"blueberry", &Blueberry}, {"mango", &Mango}, {"watermelon", &Watermelon},
		{"peach", &Peach}, {"kiwi", &Kiwi}, {"lemon", &Lemon},
		{"orange", &Orange}, {"grape", &Grape}, {"cherry", &Cherry},
		{"pineapple", &Pineapple}, {"raspberry", &Raspberry}, {"lime", &Lime},
		{"pomegranate", &Pomegranate}, {"apple", &Apple}, {"plum", &Plum},
		{"apricot", &Apricot}, {"dragonfruit", &Dragonfruit}, {"blackberry", &Blackberry},
		{"tangerine", &Tangerine}, {"fig", &Fig}, {"guava", &Guava},
		{"acai", &Acai}, {"coconut", &Coconut}, {"guarana", &Guarana},
		{"melon", &Melon},
	}},
	{"Horizon", []namedTheme{
		{"farzenith", &FarZenith}, {"banuk", &Banuk}, {"nora", &Nora},
		{"carja", &Carja}, {"oseram", &Oseram}, {"utaru", &Utaru},
		{"tenakth", &Tenakth}, {"quen", &Quen},
	}},
	{"Super Mario", []namedTheme{
		{"mario", &Mario}, {"luigi", &Luigi}, {"princesspeach", &PrincessPeach},
		{"daisy", &Daisy}, {"yoshi", &Yoshi}, {"toad", &Toad},
		{"rosalina", &Rosalina}, {"toadette", &Toadette}, {"wario", &Wario},
		{"waluigi", &Waluigi}, {"bowser", &Bowser},
	}},
	{"Sonic", []namedTheme{
		{"sonic", &Sonic}, {"tails", &Tails}, {"knuckles", &Knuckles},
		{"shadow", &Shadow}, {"amyrose", &AmyRose}, {"cream", &Cream},
		{"rouge", &Rouge}, {"eggman", &Eggman},
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
		m[nt.name] = *nt.theme
	}
	return m
}

// Lookup is the built-in theme called name, in any case; empty is Default.
// The error names what was asked for; the caller adds where its list is.
func Lookup(name string) (Theme, error) {
	if name == "" {
		return Default, nil
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
// in All() belongs to exactly one category - "plain" has no siblings of its
// own, so it lives under "Miscellaneous" rather than being left ungrouped.
func Categories() []Category {
	cats := make([]Category, len(themeRegistry))
	for i, cat := range themeRegistry {
		cats[i] = Category{Name: cat.category, Themes: namesOf(cat.themes)}
	}
	return cats
}

// Built-in theme presets. Use directly or as a Base for partial overrides.
var (
	Banana = Theme{Colors: Colors{
		Accent: "#F4D03F", Selection: "#E6FF79", Border: "#8D7B3A", Dim: "#5C4F20", Success: "#E6FF79", Danger: "#E74C3C",
	}}
	Mint = Theme{Colors: Colors{
		Accent: "#3EB489", Selection: "#98DFAF", Border: "#4A7B6F", Dim: "#2E4F46", Success: "#2ECC71", Danger: "#E74C3C",
	}}
	Strawberry = Theme{Colors: Colors{
		Accent: "#E83A59", Selection: "#FF7096", Border: "#8B3A52", Dim: "#5C2035", Success: "#4CAF50", Danger: "#C0392B",
	}}
	Blueberry = Theme{Colors: Colors{
		Accent: "#6C63FF", Selection: "#A89CFF", Border: "#4A4580", Dim: "#2E2A55", Success: "#4CAF50", Danger: "#E74C3C",
	}}
	Mango = Theme{Colors: Colors{
		Accent: "#FF9F1C", Selection: "#FFCF77", Border: "#9A6020", Dim: "#5C3A10", Success: "#5DBB63", Danger: "#E74C3C",
	}}
	Watermelon = Theme{Colors: Colors{
		Accent: "#FF4D6D", Selection: "#FF8FA3", Border: "#4A7C59", Dim: "#2D5240", Success: "#52B788", Danger: "#C9184A",
	}}
	Peach = Theme{Colors: Colors{
		Accent: "#FF8B64", Selection: "#FFCBA4", Border: "#9A6448", Dim: "#5C3A28", Success: "#5DBB63", Danger: "#E74C3C",
	}}
	Kiwi = Theme{Colors: Colors{
		Accent: "#8DB600", Selection: "#C5E84A", Border: "#5A6E2A", Dim: "#384418", Success: "#C5E84A", Danger: "#E74C3C",
	}}
	Lemon = Theme{Colors: Colors{
		Accent: "#FFE600", Selection: "#FFF176", Border: "#9A8A20", Dim: "#5C5010", Success: "#8BC34A", Danger: "#E74C3C",
	}}
	Orange = Theme{Colors: Colors{
		Accent: "#FF6B00", Selection: "#FFA040", Border: "#9A4A10", Dim: "#5C2C08", Success: "#5DBB63", Danger: "#E74C3C",
	}}
	Grape = Theme{Colors: Colors{
		Accent: "#9B59B6", Selection: "#C39BD3", Border: "#5C3A7A", Dim: "#3A2050", Success: "#5DBB63", Danger: "#E74C3C",
	}}
	Cherry = Theme{Colors: Colors{
		Accent: "#CC0000", Selection: "#FF6B9D", Border: "#7A1A30", Dim: "#4A0A1A", Success: "#4CAF50", Danger: "#8B0000",
	}}
	Pineapple = Theme{Colors: Colors{
		Accent: "#FFD700", Selection: "#FFF44F", Border: "#7A6A10", Dim: "#4A4010", Success: "#2E8B57", Danger: "#E74C3C",
	}}
	Raspberry = Theme{Colors: Colors{
		Accent: "#E91E8C", Selection: "#FF6EC7", Border: "#8B1A5A", Dim: "#5C1038", Success: "#4CAF50", Danger: "#C2185B",
	}}
	Lime = Theme{Colors: Colors{
		Accent: "#00C853", Selection: "#69FF47", Border: "#2E6B30", Dim: "#1A4020", Success: "#69FF47", Danger: "#E74C3C",
	}}
	Pomegranate = Theme{Colors: Colors{
		Accent: "#96002D", Selection: "#FF1654", Border: "#6B1020", Dim: "#3A0810", Success: "#C5E84A", Danger: "#FF1654",
	}}
	Apple = Theme{Colors: Colors{
		Accent: "#FF3B30", Selection: "#FF9F0A", Border: "#8B2020", Dim: "#4A1010", Success: "#34C759", Danger: "#FF3B30",
	}}
	Plum = Theme{Colors: Colors{
		Accent: "#8E4585", Selection: "#C490BD", Border: "#5A2A5A", Dim: "#361836", Success: "#5DBB63", Danger: "#E74C3C",
	}}
	Apricot = Theme{Colors: Colors{
		Accent: "#FBAE52", Selection: "#FDD5A0", Border: "#9A6A30", Dim: "#5C3A18", Success: "#5DBB63", Danger: "#E74C3C",
	}}
	Dragonfruit = Theme{Colors: Colors{
		Accent: "#FF2D78", Selection: "#FF6EAE", Border: "#8B1A5A", Dim: "#5C0A38", Success: "#4CAF50", Danger: "#E74C3C",
	}}
	Blackberry = Theme{Colors: Colors{
		Accent: "#5C3A6B", Selection: "#9B6FAE", Border: "#3A1E4A", Dim: "#200A30", Success: "#4CAF50", Danger: "#E74C3C",
	}}
	Tangerine = Theme{Colors: Colors{
		Accent: "#FF8C00", Selection: "#FFB347", Border: "#9A5A10", Dim: "#5C3008", Success: "#5DBB63", Danger: "#E74C3C",
	}}
	Fig = Theme{Colors: Colors{
		Accent: "#7B3F6E", Selection: "#B07AAA", Border: "#4A2048", Dim: "#2A0E30", Success: "#5DBB63", Danger: "#E74C3C",
	}}
	Guava = Theme{Colors: Colors{
		Accent: "#FF6B8A", Selection: "#FFB3C1", Border: "#8B3A50", Dim: "#5C1A30", Success: "#4CAF50", Danger: "#C0392B",
	}}
	Acai = Theme{Colors: Colors{
		Accent: "#4A1A6B", Selection: "#9B4FCC", Border: "#3A1050", Dim: "#200830", Success: "#5DBB63", Danger: "#E74C3C",
	}}
	Coconut = Theme{Colors: Colors{
		Accent: "#C4A882", Selection: "#EDD9B8", Border: "#7A6048", Dim: "#4A3828", Success: "#5DBB63", Danger: "#E74C3C",
	}}
	// Default is the adaptive default: every role left empty, so each
	// follows the terminal's light or dark background.
	Default = Theme{}
	// Plain uses only ANSI 16-colour codes, for terminals with limited colour.
	Plain = Theme{Colors: Colors{
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
	// Terminal takes every colour from the terminal's own palette, so
	// the interface matches whatever scheme the user already runs.
	Terminal = Theme{Colors: Colors{
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
	// Gruvbox is morhetz's retro palette, for a dark terminal.
	Gruvbox = Theme{Colors: Colors{
		Accent: "#FE8019", Selection: "#FABD2F", Border: "#7C6F64", Dim: "#A89984", Text: "#EBDBB2",
		Success: "#B8BB26", Warning: "#FABD2F", Danger: "#FB4934", Info: "#83A598", OnAccent: "#282828",
		Series: [6]string{"#B8BB26", "#8EC07C", "#FE8019", "#83A598", "#FABD2F", "#D3869B"},
		Ramp:   [3]string{"#B8BB26", "#FABD2F", "#FE8019"},
	}}
	Guarana = Theme{Colors: Colors{
		Accent: "#A83220", Selection: "#D4503C", Border: "#5C2A1A", Dim: "#3A1408", Success: "#4A7C2F", Danger: "#C0392B",
	}}
	// Melon: cantaloupe - salmon-orange flesh, sage-green rind borders.
	// The green sets it apart from peach, apricot and mango, which all pair an
	// orange accent with brown borders.
	Melon = Theme{Colors: Colors{
		Accent: "#E8845A", Selection: "#F7D9A8", Border: "#8A9A6B", Dim: "#5A6B45", Success: "#7BB661", Danger: "#C0392B",
	}}
	// FarZenith: the Far Zenith's ivory and bronze (Horizon Forbidden
	// West). Unfocused borders are white, the hull color; focus and error stay
	// gold/rust, so gold reads as an accent, not the base.
	FarZenith = Theme{Colors: Colors{
		Accent: "#D4AF37", Selection: "#FFFFFF", Border: "#FFFFFF", Dim: "#8A8368", Success: "#8A9A5B", Danger: "#B7472A",
	}}
	// Banuk: the Banuk's "Blue Light" - a neon cyan glow (the cables
	// shamans thread through their skin to channel it) against dark
	// weathered hide and machine-metal tones.
	Banuk = Theme{Colors: Colors{
		Accent: "#00D9FF", Selection: "#7DF9FF", Border: "#3E4A52", Dim: "#5A6670", Success: "#3ED9B0", Danger: "#FF4655",
	}}
	// Nora: earthy hide-and-forest tones, with the Nora's blue woad
	// face paint as the one cool accent against greens and browns.
	Nora = Theme{Colors: Colors{
		Accent: "#3F6B3F", Selection: "#4FB8D0", Border: "#6B5A45", Dim: "#4A3C2E", Success: "#6B8E4E", Danger: "#B33A3A",
	}}
	// Carja: the sun-worshipping Carja's royal crimson and gold, fire
	// and light against a dark ember base.
	Carja = Theme{Colors: Colors{
		Accent: "#C81E3A", Selection: "#F4A825", Border: "#7A3B3B", Dim: "#4A2020", Success: "#5DBB63", Danger: "#8B0000",
	}}
	// Oseram: forged metal and rust - no face paint, no ornamentation,
	// just iron, ember-orange heat, and industrial grey.
	Oseram = Theme{Colors: Colors{
		Accent: "#B35A2A", Selection: "#FF8C42", Border: "#4A4A48", Dim: "#3A3A38", Success: "#5DBB63", Danger: "#C0392B",
	}}
	// Utaru: woven-leaf green (their armor's dominant color) over
	// mustard sashes and tan straps, with the Utaru's white face paint as
	// the bright accent.
	Utaru = Theme{Colors: Colors{
		Accent: "#5C8A3A", Selection: "#F5F0E0", Border: "#B8860B", Dim: "#6B5A35", Success: "#D4A017", Danger: "#A63D2A",
	}}
	// Tenakth: warrior red armor on an ember base, blue clan body paint as
	// the calm accent. Border and Dim also color real text (status bar, gutter,
	// unchecked items), so they stay muted, not near-black, to stay legible.
	Tenakth = Theme{Colors: Colors{
		Accent: "#A6231F", Selection: "#3E9BC7", Border: "#8A5A4A", Dim: "#6B4A3A", Success: "#5C7A3A", Danger: "#C0392B",
	}}
	// Quen: the coastal Quen's turquoise and coral face paint against
	// blue-grey sea mist.
	Quen = Theme{Colors: Colors{
		Accent: "#1FA8A0", Selection: "#FF7F66", Border: "#5C7A82", Dim: "#3E525A", Success: "#3EBD93", Danger: "#C0392B",
	}}
	// Mario: cap-and-shirt red, white gloves as the bright accent, his
	// overalls blue demoted to unfocused borders.
	Mario = Theme{Colors: Colors{
		Accent: "#E52521", Selection: "#F0F0E8", Border: "#049CD8", Dim: "#7A6552", Success: "#43B047", Danger: "#A61B1B",
	}}
	// Luigi: his green over denim-overalls blue (the same vivid blue as
	// Mario's - previously too desaturated here and just read as grey),
	// white gloves as the bright accent.
	Luigi = Theme{Colors: Colors{
		Accent: "#43B047", Selection: "#F0F0E8", Border: "#049CD8", Dim: "#4A6B85", Success: "#8BC34A", Danger: "#C0392B",
	}}
	// PrincessPeach: dress pink, white gloves, golden hair as the second
	// vivid color (the Mario/Luigi pattern). Not named Peach, to avoid
	// colliding with the fruit preset.
	PrincessPeach = Theme{Colors: Colors{
		Accent: "#F06CA0", Selection: "#F5F5F0", Border: "#F0C419", Dim: "#B5527A", Success: "#5DBB63", Danger: "#C0392B",
	}}
	// Daisy: her dress orange (predominant) over her brown hair
	// (secondary); white sleeve, teal brooch gem, and a deeper orange trim
	// fill the rest, all pulled from the reference art rather than invented.
	Daisy = Theme{Colors: Colors{
		Accent: "#F39C12", Selection: "#F5F0E8", Border: "#8A5A3A", Dim: "#B8791E", Success: "#2E9C9C", Danger: "#C0392B",
	}}
	// Yoshi: his green body (predominant) over his white belly
	// (secondary); orange boots, red mouth, and cream spikes fill the rest.
	Yoshi = Theme{Colors: Colors{
		Accent: "#3CB043", Selection: "#FF8C1A", Border: "#F5F0E8", Dim: "#B8A888", Success: "#D64545", Danger: "#8B0000",
	}}
	// Toad: white cap/body (predominant) over his cap's red spot
	// (secondary); blue vest, gold trim, and brown shoes fill the rest.
	Toad = Theme{Colors: Colors{
		Accent: "#F5F0E8", Selection: "#2E6DA4", Border: "#E52521", Dim: "#8A6248", Success: "#F0C419", Danger: "#8B0000",
	}}
	// Rosalina: her cosmic teal gown (predominant) over her golden hair
	// (secondary); a silver crown, blue eyes, and a pink gem fill the rest.
	Rosalina = Theme{Colors: Colors{
		Accent: "#4FB8C0", Selection: "#E8E8EC", Border: "#D4AF37", Dim: "#7A9BA0", Success: "#4A9FD8", Danger: "#D6396B",
	}}
	// Toadette: her pink cap/dress (predominant) over her cap's white
	// spot (secondary); red vest trim, gold trim, and brown shoes fill the
	// rest.
	Toadette = Theme{Colors: Colors{
		Accent: "#E85AA0", Selection: "#D62839", Border: "#F5F0EC", Dim: "#8A6248", Success: "#F0C419", Danger: "#B33A5C",
	}}
	// Wario: his yellow shirt/cap (predominant) over his purple
	// overalls (secondary); white gloves, green shoes, and his pink nose
	// fill the rest.
	Wario = Theme{Colors: Colors{
		Accent: "#F0C419", Selection: "#F5F0E8", Border: "#7B2D8E", Dim: "#B87A6E", Success: "#3E8E41", Danger: "#B33A3A",
	}}
	// Waluigi: his purple shirt/cap (predominant) over his dark navy
	// overalls (secondary); white gloves, a gold "L", orange shoes, and his
	// pink nose fill the rest.
	Waluigi = Theme{Colors: Colors{
		Accent: "#7B2D8E", Selection: "#F5F0E8", Border: "#42425E", Dim: "#8A5A2E", Success: "#F0C419", Danger: "#B33A5C",
	}}
	// Bowser: his orange-tan hide (predominant) over his green shell
	// (secondary); red-orange spikes, a muted hide tone, cream claws, and
	// dark red danger fill the rest.
	Bowser = Theme{Colors: Colors{
		Accent: "#E8A33D", Selection: "#D2691E", Border: "#4A7A3A", Dim: "#B98A55", Success: "#F0E6D2", Danger: "#8B0000",
	}}
	// Sonic: his blue fur (predominant) over his tan muzzle/belly
	// (secondary); red shoes, a gold shoe buckle, and green eyes fill the rest.
	Sonic = Theme{Colors: Colors{
		Accent: "#1A50BC", Selection: "#E8291F", Border: "#FFD78F", Dim: "#C9A227", Success: "#00A845", Danger: "#8B0000",
	}}
	// Tails: his orange fur (predominant) over his white belly/tail-tips
	// (secondary); red shoes and blue eyes fill the rest.
	Tails = Theme{Colors: Colors{
		Accent: "#F1B000", Selection: "#E8291F", Border: "#F5F0E8", Dim: "#B8830A", Success: "#0FB3F0", Danger: "#8B0000",
	}}
	// Knuckles: his red fur (predominant) over his peach muzzle/chest
	// (secondary); his purple eyes and his shoe's grey, green, and yellow
	// parts fill the rest.
	Knuckles = Theme{Colors: Colors{
		Accent: "#FF1400", Selection: "#5F3FAA", Border: "#FFDDA0", Dim: "#8A8A8A", Success: "#01AA33", Danger: "#8B0000",
	}}
	// Shadow: red quill stripes (predominant) over black fur, rendered as
	// charcoal since true black vanishes on a black terminal; white chest tuft
	// and a gold rocket-shoe accent fill the rest.
	Shadow = Theme{Colors: Colors{
		Accent: "#DC0000", Selection: "#F5F0E8", Border: "#48484C", Dim: "#5A5A5C", Success: "#FFB528", Danger: "#8B0000",
	}}
	// AmyRose: her pink fur (predominant) over her red dress/shoes
	// (secondary); white gloves, her gold bracelets, and green eyes fill
	// the rest.
	AmyRose = Theme{Colors: Colors{
		Accent: "#FD95C6", Selection: "#F5F0E8", Border: "#D10000", Dim: "#C9A227", Success: "#01A900", Danger: "#8B0000",
	}}
	// Cream: her buff fur (predominant) over her orange dress
	// (secondary); her pink inner ears, brown eyes, and white muzzle fill
	// the rest.
	Cream = Theme{Colors: Colors{
		Accent: "#F5DFA0", Selection: "#F0A8C0", Border: "#E8821A", Dim: "#6B4A2E", Success: "#F5F0E8", Danger: "#C0392B",
	}}
	// Rouge: white fur (predominant) over a black catsuit (charcoal, as
	// with Shadow); pink heart chest plate, tan skin and teal-green eyes fill
	// the rest.
	Rouge = Theme{Colors: Colors{
		Accent: "#F5F0E8", Selection: "#E85AA0", Border: "#48484C", Dim: "#B89868", Success: "#2E9C7A", Danger: "#B33A3A",
	}}
	// Eggman: red coat (predominant) over black pants/boots (charcoal, as
	// with Shadow); orange mustache, gold buttons and white gloves fill the
	// rest.
	Eggman = Theme{Colors: Colors{
		Accent: "#CC2936", Selection: "#D2691E", Border: "#48484C", Dim: "#C9A227", Success: "#F0E6D2", Danger: "#8B0000",
	}}
)

// DarkTerminal asks the terminal on stdin/stdout whether its background is
// dark, for code that resolves a theme outside a running bubbletea program.
func DarkTerminal() bool { return lipgloss.HasDarkBackground(os.Stdin, os.Stdout) }
