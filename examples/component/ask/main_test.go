package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

// The bytes a terminal sends for the keys the scenes use.
const (
	down  = "\x1b[B"
	up    = "\x1b[A"
	right = "\x1b[C"
	enter = "\r"
	esc   = "\x1b"
)

// answered runs scene name with keys typed into its question and returns the
// CLI's own lines.
func answered(t *testing.T, name, keys string) string {
	t.Helper()
	var w bytes.Buffer
	th := theme.Resolve(theme.Mint(), true)
	err := run(&w, name, th, tea.WithInput(strings.NewReader(keys)), tea.WithOutput(io.Discard), tea.WithoutSignals())
	require.NoError(t, err)
	return w.String()
}

func TestEverySceneAnswers(t *testing.T) {
	cases := []struct{ name, keys, want string }{
		{"input", "api" + enter, "created ./api"},
		{"select", down + enter, "deploying to us-east"},
		{"multiselect", " " + down + " " + enter, "installing git docker"},
		{"radio", right + enter, "shell set to zsh"},
		{"checkbox", " " + enter, "started"},
		{"toggle", right + enter, "telemetry on"},
		{"number", up + enter, "listening on :8082"},
		{"slider", right + right + enter, "starting 6 workers"},
		{"date", right + enter, "scheduled for 2026-11-10"},
		{"tags", "bug" + enter + "ui" + enter + enter, "labelled bug, ui"},
	}
	require.Len(t, cases, len(scenes()), "a case per scene")
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := answered(t, tc.name, tc.keys)
			require.Equal(t, scenes()[tc.name].before+"\n"+tc.want+"\n", out)
		})
	}
}

func TestEscChangesNothing(t *testing.T) {
	require.Contains(t, answered(t, "input", esc), "nothing changed")
}

func TestAnUnknownSceneNamesTheOthers(t *testing.T) {
	err := run(io.Discard, "button", theme.Resolve(theme.Mint(), true))
	require.ErrorContains(t, err, "usage: ask checkbox|date|input|multiselect|number|radio|select|slider|tags|toggle")
}
