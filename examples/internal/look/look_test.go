package look

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

// Plain here has a colour for every role theme.Plain sets, each one as RGB.
func TestPlainIsThemePlainAsRGB(t *testing.T) {
	want, got := theme.Plain().Colors, Plain().Colors
	roles := map[string][2]string{
		"accent":  {want.Accent, got.Accent},
		"border":  {want.Border, got.Border},
		"dim":     {want.Dim, got.Dim},
		"success": {want.Success, got.Success},
		"danger":  {want.Danger, got.Danger},
		"warning": {want.Warning, got.Warning},
		"info":    {want.Info, got.Info},
	}
	for name, c := range roles {
		require.NotEmpty(t, c[0], name+" is one of theme.Plain's")
		require.True(t, strings.HasPrefix(c[1], "#"), name)
	}
	for i, s := range got.Series {
		require.True(t, strings.HasPrefix(s, "#"), "series %d", i)
	}
}

// A style with no colour answers black to RGBA, so Bold's foreground is no
// text colour for a scene or a light; Text gives the theme's own.
func TestTextIsTheThemesTextColourNotBlack(t *testing.T) {
	for name, th := range map[string]theme.Theme{"look": Plain(), "mint": theme.Mint(), "plain": theme.Plain()} {
		res := theme.Resolve(th, true)
		r, g, b, _ := Text(res).RGBA()
		require.NotZero(t, r+g+b, name)
	}
}

// Brightest reads the strongest channel of any colour a styled string sets.
func TestBrightestIsTheStrongestChannelInTheStyledText(t *testing.T) {
	require.Equal(t, 0, Brightest("plain"))
	require.Equal(t, 0, Brightest("\x1b[38;2;0;0;0mdark\x1b[m"))
	require.Equal(t, 204, Brightest("\x1b[38;2;12;204;7mx\x1b[48;2;1;2;3my\x1b[m"))
	require.Equal(t, 250, Brightest("\x1b[48;2;250;0;0mbg\x1b[m"))
}

// The terminal keeps its own background, so the recording shows its black.
func TestPlainSetsNoBackground(t *testing.T) {
	require.Empty(t, Plain().Colors.Background)
	require.Nil(t, theme.Resolve(Plain(), true).Background)
}
