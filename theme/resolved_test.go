package theme

import (
	"encoding/json"
	"os"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"
)

func TestResolveThemeZeroValueIsAdaptive(t *testing.T) {
	dark, light := Resolve(Theme{}, true), Resolve(Theme{}, false)
	require.NotEmpty(t, dark.Colors.Accent)
	require.NotEmpty(t, dark.Colors.OnAccent)
	require.NotEqual(t, dark.Colors, light.Colors, "the default must differ between light and dark")
}

func TestResolveThemeBaseOverride(t *testing.T) {
	require.Equal(t, "#3EB489", Resolve(Theme{Base: ptr(Mint())}, true).Colors.Accent)
}

func TestResolveThemeColorOverride(t *testing.T) {
	rt := Resolve(Theme{Base: ptr(Mint()), Colors: Colors{Accent: "#FF0000"}}, true)
	require.Equal(t, "#FF0000", rt.Colors.Accent, "Colors.Accent should override Base")
	require.Equal(t, "#2ECC71", rt.Colors.Success, "Success should inherit from Mint")
}

func TestResolveAccentFillsSelectionAndInfo(t *testing.T) {
	rt := Resolve(Theme{Colors: Colors{Accent: "#123456"}}, true)
	require.Equal(t, "#123456", rt.Colors.Selection)
	require.Equal(t, "#123456", rt.Colors.Info)
}

// A button is painted from the palette like everything else, never a fixed grey.
func TestButtonsTakeTheirColoursFromThePalette(t *testing.T) {
	rt := Resolve(Theme{Colors: Colors{Accent: "#112233", Border: "#445566", Text: "#778899", OnAccent: "#000000"}}, true)
	require.Equal(t, lipgloss.Color("#112233"), rt.Button.Focused.GetBackground())
	require.Equal(t, lipgloss.Color("#000000"), rt.Button.Focused.GetForeground())
	require.Equal(t, lipgloss.Color("#445566"), rt.Button.Blurred.GetBackground())
	require.Equal(t, lipgloss.Color("#778899"), rt.Button.Blurred.GetForeground())
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
		c := resolveColors(th, true)
		want, ok := golden[name]
		require.True(t, ok, name)
		require.Equal(t, want, [6]string{c.Accent, c.Selection, c.Border, c.Dim, c.Success, c.Danger}, name)
	}
}

func TestResolveFillsTheChromeSlots(t *testing.T) {
	r := Resolve(Mint(), true)
	require.Equal(t, "#3EB489", r.Colors.Accent)
	require.NotEqual(t, r.Panel.Border, r.PanelFocused.Border)
	require.NotEqual(t, r.PanelFocused.Border.GetForeground(), r.PanelFocused.Thumb.GetForeground(), "focused, the thumb stands out from its track")
	require.NotEmpty(t, r.Chrome.TabActive.Render("x"))
	require.NotEmpty(t, r.Legend.Key.Render("x"))
	require.NotEmpty(t, r.Modal.Render("x"))
}

func TestResolveWalksTheWholeBaseChain(t *testing.T) {
	root := Theme{Colors: Colors{Border: "#111111"}}
	mid := Theme{Base: &root, Colors: Colors{Accent: "#222222"}}
	rt := Resolve(Theme{Base: &mid, Colors: Colors{Dim: "#333333"}}, true)

	require.Equal(t, "#111111", rt.Colors.Border, "the base's base was skipped")
	require.Equal(t, "#222222", rt.Colors.Accent)
	require.Equal(t, "#333333", rt.Colors.Dim)
}

func TestResolveSurvivesABaseCycle(t *testing.T) {
	a := Theme{Colors: Colors{Accent: "#AAAAAA"}}
	b := Theme{Base: &a, Colors: Colors{Border: "#BBBBBB"}}
	a.Base = &b
	rt := Resolve(a, true)
	require.Equal(t, "#AAAAAA", rt.Colors.Accent)
	require.Equal(t, "#BBBBBB", rt.Colors.Border)
}

func TestSeriesDefaultsToSixDistinctColours(t *testing.T) {
	for _, dark := range []bool{true, false} {
		seen := map[string]bool{}
		for _, s := range resolveColors(Theme{}, dark).Series {
			require.NotEmpty(t, s)
			seen[s] = true
		}
		require.Len(t, seen, 6, "dark %v", dark)
	}
}

func TestAThemeOverridesPartOfItsSeries(t *testing.T) {
	c := resolveColors(Theme{Base: ptr(series), Colors: Colors{Series: [6]string{2: "#FF0000"}}}, true)
	require.Equal(t, "#FF0000", c.Series[2])
	require.Equal(t, series.Colors.Series[0], c.Series[0], "the rest is the base's")
}

// A theme that names no ramp climbs through its own success, warning and danger.
func TestRampFollowsTheThemesStates(t *testing.T) {
	c := resolveColors(Mint(), true)
	require.Equal(t, [3]string{c.Success, c.Warning, c.Danger}, c.Ramp)
}

func TestResolveStylesTheSeriesAndTheRamp(t *testing.T) {
	r := Resolve(series, true)
	require.Equal(t, lipgloss.Color(series.Colors.Series[3]), r.Series[3].GetForeground())
	require.Equal(t, lipgloss.Color(series.Colors.Ramp[2]), r.Ramp[2].GetForeground())
}

func TestPlainKeepsItsSeriesInTheTerminalsColours(t *testing.T) {
	for _, s := range resolveColors(Plain(), true).Series {
		require.NotContains(t, s, "#", "Plain uses ANSI codes only")
	}
}

func TestResolveDerivesTheMatchAndSelectionStyles(t *testing.T) {
	r := Resolve(Mint(), true)
	c := r.Colors
	require.Equal(t, lipgloss.NewStyle().Foreground(lipgloss.Color(c.OnAccent)).Background(lipgloss.Color(c.Info)), r.Match)
	require.Equal(t, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(c.OnAccent)).Background(lipgloss.Color(c.Warning)), r.MatchCurrent)
	require.Equal(t, lipgloss.NewStyle().Foreground(lipgloss.Color(c.Text)).Background(lipgloss.Color(c.Border)), r.Selected)
}

// A theme with its own background picks light or dark from it, whatever the
// terminal answered: light text on a light background would vanish.
func TestABackgroundDecidesLightOrDark(t *testing.T) {
	rt := Resolve(Theme{Colors: Colors{Background: "#FBFAFF"}}, true)
	require.Equal(t, defaultColors(false).Text, rt.Colors.Text)
	require.Equal(t, lipgloss.Color("#FBFAFF"), rt.Background)
	require.Equal(t, lipgloss.Color(rt.Colors.Text), rt.Foreground)

	require.Equal(t, defaultColors(true).Text, Resolve(Theme{Colors: Colors{Background: "#101010"}}, false).Colors.Text)
}

func TestNoBackgroundKeepsTheTerminals(t *testing.T) {
	rt := Resolve(Theme{}, true)
	require.Nil(t, rt.Background)
	require.Nil(t, rt.Foreground)
}

// Default, Plain and Terminal follow the terminal; every other preset brings
// a background, so what the terminal answers changes nothing.
func TestPresetsBringTheirOwnBackground(t *testing.T) {
	for name, th := range All() {
		c := resolveColors(th, false)
		if name == "default" || name == "plain" || name == "terminal" {
			require.Empty(t, c.Background, name)
			continue
		}
		require.NotEmpty(t, c.Background, name)
		require.Equal(t, c, resolveColors(th, true), name)
	}
}

// series is a theme that names its whole series and ramp.
var series = Theme{Colors: Colors{
	Series: [6]string{"#110000", "#220000", "#330000", "#440000", "#550000", "#660000"},
	Ramp:   [3]string{"#001100", "#002200", "#003300"},
}}

func ptr(t Theme) *Theme { return &t }
