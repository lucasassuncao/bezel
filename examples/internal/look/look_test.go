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

// The terminal keeps its own background, so the recording shows its black.
func TestPlainSetsNoBackground(t *testing.T) {
	require.Empty(t, Plain().Colors.Background)
	require.Nil(t, theme.Resolve(Plain(), true).Background)
}
