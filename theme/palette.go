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
// a dark or light terminal, or for its own Background when it has one. A theme
// that sets Accent but not Selection or Info uses its accent for them.
func ResolveColors(t Theme, dark bool) Colors {
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
	}},
	{"Classic", []namedTheme{
		{"gruvbox", &Gruvbox}, {"dracula", &Dracula}, {"nord", &Nord},
		{"solarizeddark", &SolarizedDark}, {"solarizedlight", &SolarizedLight},
		{"onedark", &OneDark}, {"monokai", &Monokai}, {"tokyonight", &TokyoNight},
		{"catppuccinmocha", &CatppuccinMocha}, {"catppuccinlatte", &CatppuccinLatte},
		{"rosepine", &RosePine}, {"everforest", &Everforest}, {"kanagawa", &Kanagawa},
		{"ubuntu", &Ubuntu},
		{"ayu", &Ayu}, {"cobalt2", &Cobalt2}, {"github", &GitHub}, {"iceberg", &Iceberg},
		{"nightowl", &NightOwl}, {"nightfox", &Nightfox}, {"onehalf", &OneHalf},
		{"shadesofpurple", &ShadesOfPurple}, {"sonokai", &Sonokai}, {"tomorrow", &Tomorrow},
	}},
	{"Apps", []namedTheme{
		{"claudecode", &ClaudeCode}, {"codex", &Codex}, {"geminicli", &GeminiCLI},
		{"lazygit", &Lazygit}, {"k9s", &K9s}, {"btop", &Btop},
	}},
	{"Styles", []namedTheme{
		{"cyberpunk", &Cyberpunk}, {"ocean", &Ocean}, {"retro", &Retro}, {"minimal", &Minimal},
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
		Accent: "#F4D03F", Selection: "#E6FF79", Border: "#8D7B3A", Dim: "#5C4F20", Success: "#E6FF79", Danger: "#E74C3C", Background: "#1B180B",
	}}
	Mint = Theme{Colors: Colors{
		Accent: "#3EB489", Selection: "#98DFAF", Border: "#4A7B6F", Dim: "#2E4F46", Success: "#2ECC71", Danger: "#E74C3C", Background: "#0B1B15",
	}}
	Strawberry = Theme{Colors: Colors{
		Accent: "#E83A59", Selection: "#FF7096", Border: "#8B3A52", Dim: "#5C2035", Success: "#4CAF50", Danger: "#C0392B", Background: "#000000",
	}}
	Blueberry = Theme{Colors: Colors{
		Accent: "#6C63FF", Selection: "#A89CFF", Border: "#4A4580", Dim: "#2E2A55", Success: "#4CAF50", Danger: "#E74C3C", Background: "#0C0B1B",
	}}
	Mango = Theme{Colors: Colors{
		Accent: "#FF9F1C", Selection: "#FFCF77", Border: "#9A6020", Dim: "#5C3A10", Success: "#5DBB63", Danger: "#E74C3C", Background: "#000000",
	}}
	Watermelon = Theme{Colors: Colors{
		Accent: "#FF4D6D", Selection: "#FF8FA3", Border: "#4A7C59", Dim: "#2D5240", Success: "#52B788", Danger: "#C9184A", Background: "#000000",
	}}
	Peach = Theme{Colors: Colors{
		Accent: "#FF8B64", Selection: "#FFCBA4", Border: "#9A6448", Dim: "#5C3A28", Success: "#5DBB63", Danger: "#E74C3C", Background: "#000000",
	}}
	Kiwi = Theme{Colors: Colors{
		Accent: "#8DB600", Selection: "#C5E84A", Border: "#5A6E2A", Dim: "#384418", Success: "#C5E84A", Danger: "#E74C3C", Background: "#000000",
	}}
	Lemon = Theme{Colors: Colors{
		Accent: "#FFE600", Selection: "#FFF176", Border: "#9A8A20", Dim: "#5C5010", Success: "#8BC34A", Danger: "#E74C3C", Background: "#000000",
	}}
	Orange = Theme{Colors: Colors{
		Accent: "#FF6B00", Selection: "#FFA040", Border: "#9A4A10", Dim: "#5C2C08", Success: "#5DBB63", Danger: "#E74C3C", Background: "#000000",
	}}
	Grape = Theme{Colors: Colors{
		Accent: "#9B59B6", Selection: "#C39BD3", Border: "#5C3A7A", Dim: "#3A2050", Success: "#5DBB63", Danger: "#E74C3C", Background: "#000000",
	}}
	Cherry = Theme{Colors: Colors{
		Accent: "#CC0000", Selection: "#FF6B9D", Border: "#7A1A30", Dim: "#4A0A1A", Success: "#4CAF50", Danger: "#8B0000", Background: "#000000",
	}}
	Pineapple = Theme{Colors: Colors{
		Accent: "#FFD700", Selection: "#FFF44F", Border: "#7A6A10", Dim: "#4A4010", Success: "#2E8B57", Danger: "#E74C3C", Background: "#000000",
	}}
	Raspberry = Theme{Colors: Colors{
		Accent: "#E91E8C", Selection: "#FF6EC7", Border: "#8B1A5A", Dim: "#5C1038", Success: "#4CAF50", Danger: "#C2185B", Background: "#000000",
	}}
	Lime = Theme{Colors: Colors{
		Accent: "#00C853", Selection: "#69FF47", Border: "#2E6B30", Dim: "#1A4020", Success: "#69FF47", Danger: "#E74C3C", Background: "#0B1B12",
	}}
	Pomegranate = Theme{Colors: Colors{
		Accent: "#96002D", Selection: "#FF1654", Border: "#6B1020", Dim: "#3A0810", Success: "#C5E84A", Danger: "#FF1654", Background: "#000000",
	}}
	Apple = Theme{Colors: Colors{
		Accent: "#FF3B30", Selection: "#FF9F0A", Border: "#8B2020", Dim: "#4A1010", Success: "#34C759", Danger: "#FF3B30", Background: "#000000",
	}}
	Plum = Theme{Colors: Colors{
		Accent: "#8E4585", Selection: "#C490BD", Border: "#5A2A5A", Dim: "#361836", Success: "#5DBB63", Danger: "#E74C3C", Background: "#000000",
	}}
	Apricot = Theme{Colors: Colors{
		Accent: "#FBAE52", Selection: "#FDD5A0", Border: "#9A6A30", Dim: "#5C3A18", Success: "#5DBB63", Danger: "#E74C3C", Background: "#1B140B",
	}}
	Dragonfruit = Theme{Colors: Colors{
		Accent: "#FF2D78", Selection: "#FF6EAE", Border: "#8B1A5A", Dim: "#5C0A38", Success: "#4CAF50", Danger: "#E74C3C", Background: "#000000",
	}}
	Blackberry = Theme{Colors: Colors{
		Accent: "#5C3A6B", Selection: "#9B6FAE", Border: "#3A1E4A", Dim: "#200A30", Success: "#4CAF50", Danger: "#E74C3C", Background: "#000000",
	}}
	Tangerine = Theme{Colors: Colors{
		Accent: "#FF8C00", Selection: "#FFB347", Border: "#9A5A10", Dim: "#5C3008", Success: "#5DBB63", Danger: "#E74C3C", Background: "#1B140B",
	}}
	Fig = Theme{Colors: Colors{
		Accent: "#7B3F6E", Selection: "#B07AAA", Border: "#4A2048", Dim: "#2A0E30", Success: "#5DBB63", Danger: "#E74C3C", Background: "#000000",
	}}
	Guava = Theme{Colors: Colors{
		Accent: "#FF6B8A", Selection: "#FFB3C1", Border: "#8B3A50", Dim: "#5C1A30", Success: "#4CAF50", Danger: "#C0392B", Background: "#000000",
	}}
	Acai = Theme{Colors: Colors{
		Accent: "#4A1A6B", Selection: "#9B4FCC", Border: "#3A1050", Dim: "#200830", Success: "#5DBB63", Danger: "#E74C3C", Background: "#000000",
	}}
	Coconut = Theme{Colors: Colors{
		Accent: "#C4A882", Selection: "#EDD9B8", Border: "#7A6048", Dim: "#4A3828", Success: "#5DBB63", Danger: "#E74C3C", Background: "#1A140C",
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
		Accent: "#FE8019", Selection: "#FABD2F", Border: "#7C6F64", Dim: "#A89984", Text: "#EBDBB2", Background: "#282828",
		Success: "#B8BB26", Warning: "#FABD2F", Danger: "#FB4934", Info: "#83A598", OnAccent: "#282828",
		Series: [6]string{"#B8BB26", "#8EC07C", "#FE8019", "#83A598", "#FABD2F", "#D3869B"},
		Ramp:   [3]string{"#B8BB26", "#FABD2F", "#FE8019"},
	}}
	// Dracula is Zeno Rocha's purple and pink on a blue-grey night.
	Dracula = Theme{Colors: Colors{
		Accent: "#BD93F9", Selection: "#FF79C6", Border: "#6272A4", Dim: "#6272A4", Text: "#F8F8F2", Background: "#282A36",
		Success: "#50FA7B", Warning: "#FFB86C", Danger: "#FF5555", Info: "#8BE9FD", OnAccent: "#282A36",
		Series: [6]string{"#50FA7B", "#8BE9FD", "#FFB86C", "#BD93F9", "#F1FA8C", "#FF79C6"},
	}}
	// Nord is Arctic Ice Studio's frost blues over polar night.
	Nord = Theme{Colors: Colors{
		Accent: "#88C0D0", Selection: "#8FBCBB", Border: "#616E88", Dim: "#616E88", Text: "#D8DEE9", Background: "#2E3440",
		Success: "#A3BE8C", Warning: "#EBCB8B", Danger: "#BF616A", Info: "#81A1C1", OnAccent: "#2E3440",
		Series: [6]string{"#A3BE8C", "#88C0D0", "#D08770", "#81A1C1", "#EBCB8B", "#B48EAD"},
	}}
	// SolarizedDark is Ethan Schoonover's palette on its dark base.
	SolarizedDark = Theme{Colors: Colors{
		Accent: "#268BD2", Selection: "#2AA198", Border: "#586E75", Dim: "#586E75", Text: "#839496", Background: "#002B36",
		Success: "#859900", Warning: "#B58900", Danger: "#DC322F", Info: "#6C71C4", OnAccent: "#002B36",
		Series: [6]string{"#859900", "#2AA198", "#CB4B16", "#268BD2", "#B58900", "#D33682"},
	}}
	// SolarizedLight is the same accents on Solarized's base2, a step off white.
	SolarizedLight = Theme{Colors: Colors{
		Accent: "#268BD2", Selection: "#2AA198", Border: "#93A1A1", Dim: "#93A1A1", Text: "#657B83", Background: "#EEE8D5",
		Success: "#859900", Warning: "#B58900", Danger: "#DC322F", Info: "#6C71C4", OnAccent: "#FDF6E3",
		Series: [6]string{"#859900", "#2AA198", "#CB4B16", "#268BD2", "#B58900", "#D33682"},
	}}
	// OneDark is Atom's dark syntax theme.
	OneDark = Theme{Colors: Colors{
		Accent: "#61AFEF", Selection: "#C678DD", Border: "#5C6370", Dim: "#5C6370", Text: "#ABB2BF", Background: "#282C34",
		Success: "#98C379", Warning: "#E5C07B", Danger: "#E06C75", Info: "#56B6C2", OnAccent: "#282C34",
		Series: [6]string{"#98C379", "#56B6C2", "#D19A66", "#61AFEF", "#E5C07B", "#C678DD"},
	}}
	// Monokai is Wimer Hazenberg's neon on warm charcoal.
	Monokai = Theme{Colors: Colors{
		Accent: "#66D9EF", Selection: "#A6E22E", Border: "#75715E", Dim: "#75715E", Text: "#F8F8F2", Background: "#272822",
		Success: "#A6E22E", Warning: "#E6DB74", Danger: "#F92672", Info: "#AE81FF", OnAccent: "#272822",
		Series: [6]string{"#A6E22E", "#66D9EF", "#FD971F", "#AE81FF", "#E6DB74", "#F92672"},
	}}
	// TokyoNight is enkia's night variant: neon blues on deep indigo.
	TokyoNight = Theme{Colors: Colors{
		Accent: "#7AA2F7", Selection: "#BB9AF7", Border: "#565F89", Dim: "#565F89", Text: "#C0CAF5", Background: "#1A1B26",
		Success: "#9ECE6A", Warning: "#E0AF68", Danger: "#F7768E", Info: "#7DCFFF", OnAccent: "#1A1B26",
		Series: [6]string{"#9ECE6A", "#7DCFFF", "#FF9E64", "#7AA2F7", "#E0AF68", "#BB9AF7"},
	}}
	// CatppuccinMocha is Catppuccin's darkest flavour, mauve on base.
	CatppuccinMocha = Theme{Colors: Colors{
		Accent: "#CBA6F7", Selection: "#F5C2E7", Border: "#6C7086", Dim: "#7F849C", Text: "#CDD6F4", Background: "#1E1E2E",
		Success: "#A6E3A1", Warning: "#F9E2AF", Danger: "#F38BA8", Info: "#89B4FA", OnAccent: "#1E1E2E",
		Series: [6]string{"#A6E3A1", "#94E2D5", "#FAB387", "#89B4FA", "#F9E2AF", "#F5C2E7"},
	}}
	// CatppuccinLatte is Catppuccin's light flavour, on its mantle rather than
	// the brighter base.
	CatppuccinLatte = Theme{Colors: Colors{
		Accent: "#8839EF", Selection: "#EA76CB", Border: "#9CA0B0", Dim: "#8C8FA1", Text: "#4C4F69", Background: "#E6E9EF",
		Success: "#40A02B", Warning: "#DF8E1D", Danger: "#D20F39", Info: "#1E66F5", OnAccent: "#EFF1F5",
		Series: [6]string{"#40A02B", "#179299", "#FE640B", "#1E66F5", "#DF8E1D", "#EA76CB"},
	}}
	// RosePine is Rosé Pine's main variant: iris and rose on a dusk base.
	RosePine = Theme{Colors: Colors{
		Accent: "#C4A7E7", Selection: "#EBBCBA", Border: "#6E6A86", Dim: "#908CAA", Text: "#E0DEF4", Background: "#191724",
		Success: "#9CCFD8", Warning: "#F6C177", Danger: "#EB6F92", Info: "#C4A7E7", OnAccent: "#191724",
		Series: [6]string{"#9CCFD8", "#31748F", "#F6C177", "#C4A7E7", "#EBBCBA", "#EB6F92"},
	}}
	// Everforest is sainnhe's soft greens, dark and medium.
	Everforest = Theme{Colors: Colors{
		Accent: "#A7C080", Selection: "#83C092", Border: "#7A8478", Dim: "#859289", Text: "#D3C6AA", Background: "#2D353B",
		Success: "#A7C080", Warning: "#DBBC7F", Danger: "#E67E80", Info: "#7FBBB3", OnAccent: "#2D353B",
		Series: [6]string{"#A7C080", "#83C092", "#E69875", "#7FBBB3", "#DBBC7F", "#D699B6"},
	}}
	// Kanagawa is rebelot's wave variant, after Hokusai's Great Wave.
	Kanagawa = Theme{Colors: Colors{
		Accent: "#7E9CD8", Selection: "#957FB8", Border: "#727169", Dim: "#727169", Text: "#DCD7BA", Background: "#1F1F28",
		Success: "#98BB6C", Warning: "#E6C384", Danger: "#E46876", Info: "#7FB4CA", OnAccent: "#1F1F28",
		Series: [6]string{"#98BB6C", "#7AA89F", "#FFA066", "#7E9CD8", "#E6C384", "#D27E99"},
	}}
	// Ubuntu is Ubuntu's terminal: its orange over the aubergine background,
	// with the Yaru palette's bright colours for the states.
	Ubuntu = Theme{Colors: Colors{
		Accent: "#E95420", Selection: "#FFFFFF", Border: "#AEA79F", Dim: "#8E8780", Text: "#EEEEEC", Background: "#300A24",
		Success: "#33D17A", Warning: "#E9AD0C", Danger: "#F66151", Info: "#33C7DE", OnAccent: "#300A24",
		Series: [6]string{"#33D17A", "#33C7DE", "#E9AD0C", "#2A7BDE", "#C061CB", "#F66151"},
	}}
	// Ayu is ayu dark: warm orange and sky blue on an ink black.
	Ayu = Theme{Colors: Colors{
		Accent: "#FFB454", Selection: "#59C2FF", Border: "#686868", Dim: "#686868", Text: "#BFBDB6", Background: "#0B0E14",
		Success: "#AAD94C", Warning: "#F9AF4F", Danger: "#F07178", Info: "#95E6CB", OnAccent: "#0B0E14",
		Series: [6]string{"#AAD94C", "#95E6CB", "#FFB454", "#59C2FF", "#F9AF4F", "#D2A6FF"},
	}}
	// Cobalt2 is Wes Bos's yellow on deep cobalt blue.
	Cobalt2 = Theme{Colors: Colors{
		Accent: "#FFC600", Selection: "#0088FF", Border: "#0050A4", Dim: "#0088FF", Text: "#FFFFFF", Background: "#122738",
		Success: "#3AD900", Warning: "#FFC600", Danger: "#FF628C", Info: "#80FCFF", OnAccent: "#122738",
		Series: [6]string{"#3AD900", "#80FCFF", "#FFC600", "#0088FF", "#FB94FF", "#FF628C"},
	}}
	// GitHub is GitHub's dark default: its link blue on near black.
	GitHub = Theme{Colors: Colors{
		Accent: "#58A6FF", Selection: "#BC8CFF", Border: "#6E7681", Dim: "#6E7681", Text: "#E6EDF3", Background: "#010409",
		Success: "#3FB950", Warning: "#D29922", Danger: "#FF7B72", Info: "#39C5CF", OnAccent: "#010409",
		Series: [6]string{"#3FB950", "#39C5CF", "#D29922", "#58A6FF", "#E3B341", "#BC8CFF"},
	}}
	// Iceberg is cocopon's cold, muted blues.
	Iceberg = Theme{Colors: Colors{
		Accent: "#84A0C6", Selection: "#89B8C2", Border: "#6B7089", Dim: "#6B7089", Text: "#C6C8D1", Background: "#161821",
		Success: "#B4BE82", Warning: "#E2A478", Danger: "#E27878", Info: "#A093C7", OnAccent: "#161821",
		Series: [6]string{"#B4BE82", "#89B8C2", "#E2A478", "#84A0C6", "#E9B189", "#A093C7"},
	}}
	// NightOwl is Sarah Drasner's blue and lilac for working at night.
	NightOwl = Theme{Colors: Colors{
		Accent: "#82AAFF", Selection: "#C792EA", Border: "#575656", Dim: "#575656", Text: "#CCCCCC", Background: "#011627",
		Success: "#22DA6E", Warning: "#FFEB95", Danger: "#EF5350", Info: "#7FDBCA", OnAccent: "#011627",
		Series: [6]string{"#22DA6E", "#21C7A8", "#FFEB95", "#82AAFF", "#C5E478", "#C792EA"},
	}}
	// Nightfox is EdenEast's default fox: soft blue on slate.
	Nightfox = Theme{Colors: Colors{
		Accent: "#719CD6", Selection: "#9D79D6", Border: "#575860", Dim: "#575860", Text: "#CDCECF", Background: "#192330",
		Success: "#81B29A", Warning: "#DBC074", Danger: "#C94F6D", Info: "#63CDCF", OnAccent: "#192330",
		Series: [6]string{"#81B29A", "#63CDCF", "#DBC074", "#719CD6", "#E0C989", "#9D79D6"},
	}}
	// OneHalf is One Half dark; its bright black is the background, so the
	// greys are One's comment grey.
	OneHalf = Theme{Colors: Colors{
		Accent: "#61AFEF", Selection: "#C678DD", Border: "#5C6370", Dim: "#5C6370", Text: "#DCDFE4", Background: "#282C34",
		Success: "#98C379", Warning: "#E5C07B", Danger: "#E06C75", Info: "#56B6C2", OnAccent: "#282C34",
		Series: [6]string{"#98C379", "#56B6C2", "#E5C07B", "#61AFEF", "#C678DD", "#E06C75"},
	}}
	// ShadesOfPurple is Ahmad Awais's yellow on purple.
	ShadesOfPurple = Theme{Colors: Colors{
		Accent: "#FAD000", Selection: "#FB94FF", Border: "#7857FE", Dim: "#7857FE", Text: "#FFFFFF", Background: "#1E1E3F",
		Success: "#3AD900", Warning: "#FAD000", Danger: "#E43937", Info: "#80FCFF", OnAccent: "#1E1E3F",
		Series: [6]string{"#3AD900", "#80FCFF", "#FAD000", "#6943FF", "#FB94FF", "#FF2C70"},
	}}
	// Sonokai is sainnhe's Monokai-inspired default: cyan and lilac on grey.
	Sonokai = Theme{Colors: Colors{
		Accent: "#76CCE0", Selection: "#B39DF3", Border: "#7F8490", Dim: "#7F8490", Text: "#E2E2E3", Background: "#2C2E34",
		Success: "#9ED072", Warning: "#E7C664", Danger: "#FC5D7C", Info: "#F39660", OnAccent: "#2C2E34",
		Series: [6]string{"#9ED072", "#76CCE0", "#F39660", "#B39DF3", "#E7C664", "#FC5D7C"},
	}}
	// Tomorrow is Chris Kempson's Tomorrow Night; its bright black is black,
	// so the greys are the scheme's comment grey.
	Tomorrow = Theme{Colors: Colors{
		Accent: "#81A2BE", Selection: "#B294BB", Border: "#969896", Dim: "#969896", Text: "#C5C8C6", Background: "#1D1F21",
		Success: "#B5BD68", Warning: "#F0C674", Danger: "#CC6666", Info: "#8ABEB7", OnAccent: "#1D1F21",
		Series: [6]string{"#B5BD68", "#8ABEB7", "#F0C674", "#81A2BE", "#B294BB", "#CC6666"},
	}}
	// ClaudeCode is Claude Code's terracotta and hot pink on near black.
	ClaudeCode = Theme{Colors: Colors{
		Accent: "#D77757", Selection: "#FD5DB1", Border: "#888888", Dim: "#888888", Text: "#FFFFFF", Background: "#1A1A1A",
		Success: "#4EBA65", Warning: "#FFC107", Danger: "#FF6B80", Info: "#B1B9F9", OnAccent: "#1A1A1A",
		Series: [6]string{"#4EBA65", "#B1B9F9", "#D77757", "#AF87FF", "#FFC107", "#FD5DB1"},
	}}
	// Codex is the Codex CLI's monochrome, green only for what succeeded.
	Codex = Theme{Colors: Colors{
		Accent: "#E0E0E0", Selection: "#22C55E", Border: "#555555", Dim: "#808080", Text: "#E0E0E0", Background: "#000000",
		Success: "#22C55E", Warning: "#EAB308", Danger: "#EF4444", OnAccent: "#000000",
	}}
	// GeminiCLI is the Gemini CLI's soft blue and lilac on a navy black.
	GeminiCLI = Theme{Colors: Colors{
		Accent: "#87AFFF", Selection: "#D7AFFF", Border: "#808080", Dim: "#808080", Text: "#E0E0E0", Background: "#1A1A2E",
		Success: "#D7FFD7", Warning: "#FFFFAF", Danger: "#FF87AF", Info: "#87D7D7", OnAccent: "#1A1A2E",
		Series: [6]string{"#D7FFD7", "#87D7D7", "#FFFFAF", "#87AFFF", "#D7AFFF", "#FF87AF"},
	}}
	// Lazygit is lazygit's green active border and blue selection on black.
	Lazygit = Theme{Colors: Colors{
		Accent: "#22C55E", Selection: "#3B82F6", Border: "#6B7280", Dim: "#6B7280", Text: "#D4D4D4", Background: "#000000",
		Success: "#22C55E", Warning: "#EAB308", Danger: "#EF4444", Info: "#06B6D4", OnAccent: "#000000",
		Series: [6]string{"#22C55E", "#06B6D4", "#EAB308", "#3B82F6", "#D946EF", "#EF4444"},
	}}
	// K9s is k9s's aqua and dodger blue, with cadet blue text, on black.
	K9s = Theme{Colors: Colors{
		Accent: "#00FFFF", Selection: "#1E90FF", Border: "#1E90FF", Dim: "#808080", Text: "#5F9EA0", Background: "#000000",
		Success: "#ADFF2F", Warning: "#FF8C00", Danger: "#FF4500", Info: "#FF00FF", OnAccent: "#000000",
		Series: [6]string{"#ADFF2F", "#00FFFF", "#FFA500", "#1E90FF", "#87CEFA", "#9370DB"},
	}}
	// Btop is btop's healthy green and alert red on pure black.
	Btop = Theme{Colors: Colors{
		Accent: "#77CA9B", Selection: "#EEEEEE", Border: "#555555", Dim: "#888888", Text: "#CCCCCC", Background: "#000000",
		Success: "#77CA9B", Warning: "#CBC06C", Danger: "#DC4C4C", OnAccent: "#000000",
	}}
	// Cyberpunk is neon cyan and magenta in dark rain.
	Cyberpunk = Theme{Colors: Colors{
		Accent: "#00FFFF", Selection: "#FF00FF", Border: "#4A4A6A", Dim: "#4A4A6A", Text: "#E0E0FF", Background: "#0D0D1A",
		Success: "#00FF88", Warning: "#FFAA00", Danger: "#FF0044", Info: "#FFFF00", OnAccent: "#0D0D1A",
		Series: [6]string{"#00FFFF", "#00FF88", "#FFFF00", "#FF00FF", "#FFAA00", "#FF0044"},
	}}
	// Ocean is sky blue and sea foam over the deep.
	Ocean = Theme{Colors: Colors{
		Accent: "#48DBFB", Selection: "#55EFC4", Border: "#576574", Dim: "#576574", Text: "#C8D6E5", Background: "#0B1929",
		Success: "#00D2D3", Warning: "#FECA57", Danger: "#FF6B6B", Info: "#0ABDE3", OnAccent: "#0B1929",
	}}
	// Retro is an amber CRT: every colour glows out of the same phosphor.
	Retro = Theme{Colors: Colors{
		Accent: "#FFCC00", Selection: "#FF6600", Border: "#665500", Dim: "#CC8800", Text: "#FFB000", Background: "#1A1200",
		Success: "#33FF00", Warning: "#FFCC00", Danger: "#FF0033", Info: "#CC8800", OnAccent: "#1A1200",
	}}
	// Minimal is white on near black, with one blue for what matters.
	Minimal = Theme{Colors: Colors{
		Accent: "#FFFFFF", Selection: "#0070F3", Border: "#555555", Dim: "#888888", Text: "#EDEDED", Background: "#0A0A0A",
		Success: "#00C853", Warning: "#F5A623", Danger: "#EE0000", Info: "#0070F3", OnAccent: "#0A0A0A",
	}}
	Guarana = Theme{Colors: Colors{
		Accent: "#A83220", Selection: "#D4503C", Border: "#5C2A1A", Dim: "#3A1408", Success: "#4A7C2F", Danger: "#C0392B", Background: "#000000",
	}}
	// Melon: cantaloupe - salmon-orange flesh, sage-green rind borders.
	// The green sets it apart from peach, apricot and mango, which all pair an
	// orange accent with brown borders.
	Melon = Theme{Colors: Colors{
		Accent: "#E8845A", Selection: "#F7D9A8", Border: "#8A9A6B", Dim: "#5A6B45", Success: "#7BB661", Danger: "#C0392B", Background: "#000000",
	}}
	// FarZenith: the Far Zenith's ivory and bronze (Horizon Forbidden
	// West). Unfocused borders are white, the hull color; focus and error stay
	// gold/rust, so gold reads as an accent, not the base.
	FarZenith = Theme{Colors: Colors{
		Accent: "#D4AF37", Selection: "#FFFFFF", Border: "#FFFFFF", Dim: "#8A8368", Success: "#8A9A5B", Danger: "#B7472A", Background: "#1B170B",
	}}
	// Banuk: the Banuk's "Blue Light" - a neon cyan glow (the cables
	// shamans thread through their skin to channel it) against dark
	// weathered hide and machine-metal tones.
	Banuk = Theme{Colors: Colors{
		Accent: "#00D9FF", Selection: "#7DF9FF", Border: "#3E4A52", Dim: "#5A6670", Success: "#3ED9B0", Danger: "#FF4655", Background: "#0B181B",
	}}
	// Nora: earthy hide-and-forest tones, with the Nora's blue woad
	// face paint as the one cool accent against greens and browns.
	Nora = Theme{Colors: Colors{
		Accent: "#3F6B3F", Selection: "#4FB8D0", Border: "#6B5A45", Dim: "#4A3C2E", Success: "#6B8E4E", Danger: "#B33A3A", Background: "#0E180E",
	}}
	// Carja: the sun-worshipping Carja's royal crimson and gold, fire
	// and light against a dark ember base.
	Carja = Theme{Colors: Colors{
		Accent: "#C81E3A", Selection: "#F4A825", Border: "#7A3B3B", Dim: "#4A2020", Success: "#5DBB63", Danger: "#8B0000", Background: "#1B0B0E",
	}}
	// Oseram: forged metal and rust - no face paint, no ornamentation,
	// just iron, ember-orange heat, and industrial grey.
	Oseram = Theme{Colors: Colors{
		Accent: "#B35A2A", Selection: "#FF8C42", Border: "#4A4A48", Dim: "#3A3A38", Success: "#5DBB63", Danger: "#C0392B", Background: "#1B110B",
	}}
	// Utaru: woven-leaf green (their armor's dominant color) over
	// mustard sashes and tan straps, with the Utaru's white face paint as
	// the bright accent.
	Utaru = Theme{Colors: Colors{
		Accent: "#5C8A3A", Selection: "#F5F0E0", Border: "#B8860B", Dim: "#6B5A35", Success: "#D4A017", Danger: "#A63D2A", Background: "#121B0B",
	}}
	// Tenakth: warrior red armor on an ember base, blue clan body paint as
	// the calm accent. Border and Dim also color real text (status bar, gutter,
	// unchecked items), so they stay muted, not near-black, to stay legible.
	Tenakth = Theme{Colors: Colors{
		Accent: "#A6231F", Selection: "#3E9BC7", Border: "#8A5A4A", Dim: "#6B4A3A", Success: "#5C7A3A", Danger: "#C0392B", Background: "#1B0C0B",
	}}
	// Quen: the coastal Quen's turquoise and coral face paint against
	// blue-grey sea mist.
	Quen = Theme{Colors: Colors{
		Accent: "#1FA8A0", Selection: "#FF7F66", Border: "#5C7A82", Dim: "#3E525A", Success: "#3EBD93", Danger: "#C0392B", Background: "#0B1B1A",
	}}
	// Mario: cap-and-shirt red, white gloves as the bright accent, his
	// overalls blue demoted to unfocused borders.
	Mario = Theme{Colors: Colors{
		Accent: "#E52521", Selection: "#F0F0E8", Border: "#049CD8", Dim: "#7A6552", Success: "#43B047", Danger: "#A61B1B", Background: "#1B0C0B",
	}}
	// Luigi: his green over denim-overalls blue (the same vivid blue as
	// Mario's - previously too desaturated here and just read as grey),
	// white gloves as the bright accent.
	Luigi = Theme{Colors: Colors{
		Accent: "#43B047", Selection: "#F0F0E8", Border: "#049CD8", Dim: "#4A6B85", Success: "#8BC34A", Danger: "#C0392B", Background: "#0B1B0C",
	}}
	// PrincessPeach: dress pink, white gloves, golden hair as the second
	// vivid color (the Mario/Luigi pattern). Not named Peach, to avoid
	// colliding with the fruit preset.
	PrincessPeach = Theme{Colors: Colors{
		Accent: "#F06CA0", Selection: "#F5F5F0", Border: "#F0C419", Dim: "#B5527A", Success: "#5DBB63", Danger: "#C0392B", Background: "#1B0B12",
	}}
	// Daisy: her dress orange (predominant) over her brown hair
	// (secondary); white sleeve, teal brooch gem, and a deeper orange trim
	// fill the rest, all pulled from the reference art rather than invented.
	Daisy = Theme{Colors: Colors{
		Accent: "#F39C12", Selection: "#F5F0E8", Border: "#8A5A3A", Dim: "#B8791E", Success: "#2E9C9C", Danger: "#C0392B", Background: "#1B150B",
	}}
	// Yoshi: his green body (predominant) over his white belly
	// (secondary); orange boots, red mouth, and cream spikes fill the rest.
	Yoshi = Theme{Colors: Colors{
		Accent: "#3CB043", Selection: "#FF8C1A", Border: "#F5F0E8", Dim: "#B8A888", Success: "#D64545", Danger: "#8B0000", Background: "#0B1B0C",
	}}
	// Toad: white cap/body (predominant) over his cap's red spot
	// (secondary); blue vest, gold trim, and brown shoes fill the rest.
	Toad = Theme{Colors: Colors{
		Accent: "#F5F0E8", Selection: "#2E6DA4", Border: "#E52521", Dim: "#8A6248", Success: "#F0C419", Danger: "#8B0000", Background: "#0B141B",
	}}
	// Rosalina: her cosmic teal gown (predominant) over her golden hair
	// (secondary); a silver crown, blue eyes, and a pink gem fill the rest.
	Rosalina = Theme{Colors: Colors{
		Accent: "#4FB8C0", Selection: "#E8E8EC", Border: "#D4AF37", Dim: "#7A9BA0", Success: "#4A9FD8", Danger: "#D6396B", Background: "#0B1A1B",
	}}
	// Toadette: her pink cap/dress (predominant) over her cap's white
	// spot (secondary); red vest trim, gold trim, and brown shoes fill the
	// rest.
	Toadette = Theme{Colors: Colors{
		Accent: "#E85AA0", Selection: "#D62839", Border: "#F5F0EC", Dim: "#8A6248", Success: "#F0C419", Danger: "#B33A5C", Background: "#1B0B13",
	}}
	// Wario: his yellow shirt/cap (predominant) over his purple
	// overalls (secondary); white gloves, green shoes, and his pink nose
	// fill the rest.
	Wario = Theme{Colors: Colors{
		Accent: "#F0C419", Selection: "#F5F0E8", Border: "#7B2D8E", Dim: "#B87A6E", Success: "#3E8E41", Danger: "#B33A3A", Background: "#1B180B",
	}}
	// Waluigi: his purple shirt/cap (predominant) over his dark navy
	// overalls (secondary); white gloves, a gold "L", orange shoes, and his
	// pink nose fill the rest.
	Waluigi = Theme{Colors: Colors{
		Accent: "#7B2D8E", Selection: "#F5F0E8", Border: "#42425E", Dim: "#8A5A2E", Success: "#F0C419", Danger: "#B33A5C", Background: "#180B1B",
	}}
	// Bowser: his orange-tan hide (predominant) over his green shell
	// (secondary); red-orange spikes, a muted hide tone, cream claws, and
	// dark red danger fill the rest.
	Bowser = Theme{Colors: Colors{
		Accent: "#E8A33D", Selection: "#D2691E", Border: "#4A7A3A", Dim: "#B98A55", Success: "#F0E6D2", Danger: "#8B0000", Background: "#1B150B",
	}}
	// Sonic: his blue fur (predominant) over his tan muzzle/belly
	// (secondary); red shoes, a gold shoe buckle, and green eyes fill the rest.
	Sonic = Theme{Colors: Colors{
		Accent: "#1A50BC", Selection: "#E8291F", Border: "#FFD78F", Dim: "#C9A227", Success: "#00A845", Danger: "#8B0000", Background: "#0B111B",
	}}
	// Tails: his orange fur (predominant) over his white belly/tail-tips
	// (secondary); red shoes and blue eyes fill the rest.
	Tails = Theme{Colors: Colors{
		Accent: "#F1B000", Selection: "#E8291F", Border: "#F5F0E8", Dim: "#B8830A", Success: "#0FB3F0", Danger: "#8B0000", Background: "#1B170B",
	}}
	// Knuckles: his red fur (predominant) over his peach muzzle/chest
	// (secondary); his purple eyes and his shoe's grey, green, and yellow
	// parts fill the rest.
	Knuckles = Theme{Colors: Colors{
		Accent: "#FF1400", Selection: "#5F3FAA", Border: "#FFDDA0", Dim: "#8A8A8A", Success: "#01AA33", Danger: "#8B0000", Background: "#1B0D0B",
	}}
	// Shadow: red quill stripes (predominant) over black fur, rendered as
	// charcoal since true black vanishes on a black terminal; white chest tuft
	// and a gold rocket-shoe accent fill the rest.
	Shadow = Theme{Colors: Colors{
		Accent: "#DC0000", Selection: "#F5F0E8", Border: "#48484C", Dim: "#5A5A5C", Success: "#FFB528", Danger: "#8B0000", Background: "#141416",
	}}
	// AmyRose: her pink fur (predominant) over her red dress/shoes
	// (secondary); white gloves, her gold bracelets, and green eyes fill
	// the rest.
	AmyRose = Theme{Colors: Colors{
		Accent: "#FD95C6", Selection: "#F5F0E8", Border: "#D10000", Dim: "#C9A227", Success: "#01A900", Danger: "#8B0000", Background: "#1B0B13",
	}}
	// Cream: her buff fur (predominant) over her orange dress
	// (secondary); her pink inner ears, brown eyes, and white muzzle fill
	// the rest.
	Cream = Theme{Colors: Colors{
		Accent: "#F5DFA0", Selection: "#F0A8C0", Border: "#E8821A", Dim: "#6B4A2E", Success: "#F5F0E8", Danger: "#C0392B", Background: "#1B170B",
	}}
	// Rouge: white fur (predominant) over a black catsuit (charcoal, as
	// with Shadow); pink heart chest plate, tan skin and teal-green eyes fill
	// the rest.
	Rouge = Theme{Colors: Colors{
		Accent: "#F5F0E8", Selection: "#E85AA0", Border: "#48484C", Dim: "#B89868", Success: "#2E9C7A", Danger: "#B33A3A", Background: "#17141A",
	}}
	// Eggman: red coat (predominant) over black pants/boots (charcoal, as
	// with Shadow); orange mustache, gold buttons and white gloves fill the
	// rest.
	Eggman = Theme{Colors: Colors{
		Accent: "#CC2936", Selection: "#D2691E", Border: "#48484C", Dim: "#C9A227", Success: "#F0E6D2", Danger: "#8B0000", Background: "#1B0B0D",
	}}
)

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
