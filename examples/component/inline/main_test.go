package main

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/form"
	"github.com/lucasassuncao/bezel/component/inline"
	"github.com/lucasassuncao/bezel/theme"
)

func answer(ok bool, err error) func(string) (bool, error) {
	return func(string) (bool, error) { return ok, err }
}

// pick answers the mirror question with the i-th option, or with err.
func pick(i int, err error) func(form.Select) (form.Select, error) {
	return func(s form.Select) (form.Select, error) { return s.WithSelected(i), err }
}

func TestInlineYesDrawsTheBarToTheEnd(t *testing.T) {
	var buf bytes.Buffer
	th := theme.Resolve(theme.Mint(), true)
	require.NoError(t, run(&buf, th, pick(1, nil), answer(true, nil), 0))
	out := ansi.Strip(buf.String())
	t.Log("\n" + out)
	require.Contains(t, out, "from us-east")
	require.Contains(t, out, "20/20")
	require.Contains(t, out, "done")
}

func TestInlineNoAndEscSkip(t *testing.T) {
	th := theme.Resolve(theme.Mint(), true)
	for _, a := range []func(string) (bool, error){answer(false, nil), answer(false, inline.ErrAborted)} {
		var buf bytes.Buffer
		require.NoError(t, run(&buf, th, pick(0, nil), a, 0))
		require.Contains(t, buf.String(), "skipped")
		require.NotContains(t, buf.String(), "done")
	}
}

func TestInlineEscOnTheMirrorSkipsTheQuestion(t *testing.T) {
	var buf bytes.Buffer
	th := theme.Resolve(theme.Mint(), true)
	require.NoError(t, run(&buf, th, pick(0, inline.ErrAborted), answer(true, nil), 0))
	require.Contains(t, buf.String(), "skipped")
	require.NotContains(t, buf.String(), "target")
}
