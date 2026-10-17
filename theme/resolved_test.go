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
