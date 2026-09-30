package main

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/inline"
	"github.com/lucasassuncao/bezel/theme"
)

func answer(ok bool, err error) func(string) (bool, error) {
	return func(string) (bool, error) { return ok, err }
}

func TestInlineYesDrawsTheBarToTheEnd(t *testing.T) {
	var buf bytes.Buffer
	th := theme.Resolve(theme.ThemeMint, true)
	require.NoError(t, run(&buf, th, answer(true, nil), 0))
	out := ansi.Strip(buf.String())
	t.Log("\n" + out)
	require.Contains(t, out, "20/20")
	require.Contains(t, out, "done")
}

func TestInlineNoAndEscSkip(t *testing.T) {
	th := theme.Resolve(theme.ThemeMint, true)
	for _, a := range []func(string) (bool, error){answer(false, nil), answer(false, inline.ErrAborted)} {
		var buf bytes.Buffer
		require.NoError(t, run(&buf, th, a, 0))
		require.Contains(t, buf.String(), "skipped")
		require.NotContains(t, buf.String(), "done")
	}
}
