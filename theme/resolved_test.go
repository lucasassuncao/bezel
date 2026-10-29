package theme

import (
	"encoding/json"
	"os"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveThemeZeroValueIsAdaptive(t *testing.T) {
	dark, light := Resolve(Theme{}, true), Resolve(Theme{}, false)
	assert.NotEmpty(t, dark.Colors.Accent)
	assert.NotEmpty(t, dark.Colors.OnAccent)
	assert.NotEqual(t, dark.Colors, light.Colors, "the default must differ between light and dark")
}

func TestResolveThemeBaseOverride(t *testing.T) {
	require.Equal(t, "#9B59B6", Resolve(Theme{Base: &Grape}, true).Colors.Accent)
}

func TestResolveThemeColorOverride(t *testing.T) {
	rt := Resolve(Theme{Base: &Grape, Colors: Colors{Accent: "#FF0000"}}, true)
	assert.Equal(t, "#FF0000", rt.Colors.Accent, "Colors.Accent should override Base")
	assert.Equal(t, "#5DBB63", rt.Colors.Success, "Success should inherit from Grape")
}

func TestResolveAccentFillsSelectionAndInfo(t *testing.T) {
	rt := Resolve(Theme{Colors: Colors{Accent: "#123456"}}, true)
	assert.Equal(t, "#123456", rt.Colors.Selection)
	assert.Equal(t, "#123456", rt.Colors.Info)
}

// A button is painted from the palette like everything else, never a fixed grey.
func TestButtonsTakeTheirColoursFromThePalette(t *testing.T) {
	rt := Resolve(Theme{Colors: Colors{Accent: "#112233", Border: "#445566", Text: "#778899", OnAccent: "#000000"}}, true)
	assert.Equal(t, lipgloss.Color("#112233"), rt.Button.Focused.GetBackground())
	assert.Equal(t, lipgloss.Color("#000000"), rt.Button.Focused.GetForeground())
	assert.Equal(t, lipgloss.Color("#445566"), rt.Button.Blurred.GetBackground())
	assert.Equal(t, lipgloss.Color("#778899"), rt.Button.Blurred.GetForeground())
}

func TestResolveThemeStyleOverride(t *testing.T) {
	custom := lipgloss.NewStyle().Bold(true)
	rt := Resolve(Theme{Styles: Styles{Danger: &custom}}, true)
	require.Equal(t, custom, rt.Danger)
}

// The rename to the current role names kept every preset's colours: the golden
// file was dumped from the old API as [Accent, Selection, Border, Dim, Success, Danger].
func TestPresetsKeepTheirColours(t *testing.T) {
	b, err := os.ReadFile("testdata/presets_golden.json")
	require.NoError(t, err)
	var golden map[string][6]string
	require.NoError(t, json.Unmarshal(b, &golden))

	for name, th := range All() {
		if name == "default" {
			continue // new: the adaptive default has no old colours
		}
		c := ResolveColors(th, true)
		want, ok := golden[name]
		require.True(t, ok, name)
		assert.Equal(t, want, [6]string{c.Accent, c.Selection, c.Border, c.Dim, c.Success, c.Danger}, name)
	}
}

func TestResolveFillsTheChromeSlots(t *testing.T) {
	r := Resolve(Mint, true)
	require.Equal(t, "#3EB489", r.Colors.Accent)
	require.NotEqual(t, r.Panel.Border, r.PanelFocused.Border)
	require.NotEqual(t, r.PanelFocused.Border.GetForeground(), r.PanelFocused.Thumb.GetForeground(), "focused, the thumb stands out from its track")
	require.NotEmpty(t, r.Chrome.TabActive.Render("x"))
	require.NotEmpty(t, r.Legend.Key.Render("x"))
	require.NotEmpty(t, r.Modal.Render("x"))
}

func TestResolveWalksTheWholeBaseChain(t *testing.T) {
	bold := lipgloss.NewStyle().Bold(true)
	root := Theme{Colors: Colors{Border: "#111111"}, Styles: Styles{Muted: &bold}}
	mid := Theme{Base: &root, Colors: Colors{Accent: "#222222"}}
	rt := Resolve(Theme{Base: &mid, Colors: Colors{Dim: "#333333"}}, true)

	assert.Equal(t, "#111111", rt.Colors.Border, "the base's base was skipped")
	assert.Equal(t, "#222222", rt.Colors.Accent)
	assert.Equal(t, "#333333", rt.Colors.Dim)
	assert.Equal(t, bold, rt.Muted, "a base's Styles were skipped")
}

func TestResolveSurvivesABaseCycle(t *testing.T) {
	a := Theme{Colors: Colors{Accent: "#AAAAAA"}}
	b := Theme{Base: &a, Colors: Colors{Border: "#BBBBBB"}}
	a.Base = &b
	rt := Resolve(a, true)
	assert.Equal(t, "#AAAAAA", rt.Colors.Accent)
	assert.Equal(t, "#BBBBBB", rt.Colors.Border)
}

func TestSeriesDefaultsToSixDistinctColours(t *testing.T) {
	for _, dark := range []bool{true, false} {
		seen := map[string]bool{}
		for _, s := range ResolveColors(Theme{}, dark).Series {
			require.NotEmpty(t, s)
			seen[s] = true
		}
		assert.Len(t, seen, 6, "dark %v", dark)
	}
}

func TestAThemeOverridesPartOfItsSeries(t *testing.T) {
	c := ResolveColors(Theme{Base: &Gruvbox, Colors: Colors{Series: [6]string{2: "#FF0000"}}}, true)
	assert.Equal(t, "#FF0000", c.Series[2])
	assert.Equal(t, Gruvbox.Colors.Series[0], c.Series[0], "the rest is the base's")
}

// A theme that names no ramp climbs through its own success, warning and danger.
func TestRampFollowsTheThemesStates(t *testing.T) {
	c := ResolveColors(Mint, true)
	assert.Equal(t, [3]string{c.Success, c.Warning, c.Danger}, c.Ramp)
}

func TestResolveStylesTheSeriesAndTheRamp(t *testing.T) {
	r := Resolve(Gruvbox, true)
	assert.Equal(t, lipgloss.Color(Gruvbox.Colors.Series[3]), r.Series[3].GetForeground())
	assert.Equal(t, lipgloss.Color(Gruvbox.Colors.Ramp[2]), r.Ramp[2].GetForeground())
}

func TestPlainKeepsItsSeriesInTheTerminalsColours(t *testing.T) {
	for _, s := range ResolveColors(Plain, true).Series {
		assert.NotContains(t, s, "#", "Plain uses ANSI codes only")
	}
}

func TestLookupFindsGruvbox(t *testing.T) {
	got, err := Lookup("Gruvbox")
	require.NoError(t, err)
	assert.Equal(t, Gruvbox.Colors, got.Colors)
}

func TestResolveDerivesTheMatchAndSelectionStyles(t *testing.T) {
	r := Resolve(Mint, true)
	c := r.Colors
	require.Equal(t, lipgloss.NewStyle().Foreground(lipgloss.Color(c.OnAccent)).Background(lipgloss.Color(c.Info)), r.Match)
	require.Equal(t, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(c.OnAccent)).Background(lipgloss.Color(c.Warning)), r.MatchCurrent)
	require.Equal(t, lipgloss.NewStyle().Foreground(lipgloss.Color(c.Text)).Background(lipgloss.Color(c.Border)), r.Selected)
}

// A theme with its own background picks light or dark from it, whatever the
// terminal answered: light text on a light background would vanish.
func TestABackgroundDecidesLightOrDark(t *testing.T) {
	rt := Resolve(Theme{Colors: Colors{Background: "#FBFAFF"}}, true)
	assert.Equal(t, defaultColors(false).Text, rt.Colors.Text)
	assert.Equal(t, lipgloss.Color("#FBFAFF"), rt.Background)
	assert.Equal(t, lipgloss.Color(rt.Colors.Text), rt.Foreground)

	assert.Equal(t, defaultColors(true).Text, Resolve(Theme{Colors: Colors{Background: "#101010"}}, false).Colors.Text)
}

func TestNoBackgroundKeepsTheTerminals(t *testing.T) {
	rt := Resolve(Theme{}, true)
	assert.Nil(t, rt.Background)
	assert.Nil(t, rt.Foreground)
}

// Default, Plain and Terminal follow the terminal; every other preset brings
// a background, so what the terminal answers changes nothing.
func TestPresetsBringTheirOwnBackground(t *testing.T) {
	for name, th := range All() {
		c := ResolveColors(th, false)
		if name == "default" || name == "plain" || name == "terminal" {
			assert.Empty(t, c.Background, name)
			continue
		}
		require.NotEmpty(t, c.Background, name)
		assert.Equal(t, c, ResolveColors(th, true), name)
	}
}
