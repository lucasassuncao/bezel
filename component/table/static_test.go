package table

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestStaticPrintsTheHeaderAndEveryRow(t *testing.T) {
	cols := []Column{{Title: "NAME"}, {Title: "STATUS"}}
	rows := [][]string{{"api", "running"}, {"worker", "stopped"}}
	got := ansi.Strip(Static(cols, rows, 0, th))
	require.Equal(t, "NAME    STATUS\napi     running\nworker  stopped", got)
}

func TestStaticShrinksTheFlexColumnToTheWidth(t *testing.T) {
	cols := []Column{{Title: "NAME", Flex: true}, {Title: "STATUS"}}
	rows := [][]string{{"a-very-long-service-name", "running"}}
	for _, line := range strings.Split(ansi.Strip(Static(cols, rows, 24, th)), "\n") {
		require.LessOrEqual(t, ansi.StringWidth(line), 24)
	}
}

func TestStaticWithNoRowsIsTheHeader(t *testing.T) {
	require.Equal(t, "NAME", ansi.Strip(Static([]Column{{Title: "NAME"}}, nil, 0, th)))
}

// Static is printed straight to the terminal: a cell from data must not be
// able to set the clipboard or clear the screen.
func TestStaticPrintsNoCommandsFromItsCells(t *testing.T) {
	out := Static([]Column{{Title: "name\x1b]0;t\x07"}}, [][]string{{"pod \x1b]52;c;aGk=\x07 \x1b[2J\r"}}, 80, th)
	require.NotContains(t, out, "\x1b]")
	require.NotContains(t, out, "\x1b[2J")
	require.NotContains(t, out, "\r")
}
