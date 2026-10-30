package inline

import (
	"context"
	"io"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.Plain(), true)

func press(c confirm, keys ...string) confirm {
	for _, k := range keys {
		var msg tea.KeyPressMsg
		switch k {
		case "enter":
			msg = tea.KeyPressMsg{Code: tea.KeyEnter}
		case "esc":
			msg = tea.KeyPressMsg{Code: tea.KeyEscape}
		case "right":
			msg = tea.KeyPressMsg{Code: tea.KeyRight}
		default:
			msg = tea.KeyPressMsg{Code: rune(k[0]), Text: k}
		}
		m, _ := c.Update(msg)
		c = m.(confirm)
	}
	return c
}

func TestConfirmAnswers(t *testing.T) {
	cases := []struct {
		name         string
		keys         []string
		yes, aborted bool
	}{
		{"enter keeps the default No", []string{"enter"}, false, false},
		{"y answers at once", []string{"y"}, true, false},
		{"n answers at once", []string{"n"}, false, false},
		{"arrow toggles then enter", []string{"right", "enter"}, true, false},
		{"esc aborts", []string{"esc"}, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := press(newConfirm("Proceed?", th), tc.keys...)
			assert.True(t, c.done)
			assert.Equal(t, tc.yes, c.yes)
			assert.Equal(t, tc.aborted, c.aborted)
		})
	}
}

func TestConfirmIgnoresVimKeys(t *testing.T) {
	for _, k := range []string{"h", "l"} {
		c := press(newConfirm("Proceed?", th), k, "enter")
		assert.False(t, c.yes, "%s does not toggle; the default No stands", k)
	}
}

func TestConfirmHintNamesItsKeys(t *testing.T) {
	assert.Contains(t, ansi.Strip(newConfirm("Proceed?", th).View().Content), "←/→ choose · y/n answer · enter confirm · esc cancel")
}

func TestConfirmViewShowsAnswerWhenDone(t *testing.T) {
	c := press(newConfirm("Proceed?", th), "y")
	assert.Equal(t, "Proceed? Yes\n", ansi.Strip(c.View().Content))
}

// run answers Confirm through opts instead of a terminal.
func run(t *testing.T, opts ...tea.ProgramOption) (bool, error) {
	t.Helper()
	base := []tea.ProgramOption{tea.WithOutput(io.Discard), tea.WithoutSignals()}
	return Confirm("Proceed?", th, append(base, opts...)...)
}

func TestConfirmReadsTheAnswerFromTheGivenInput(t *testing.T) {
	yes, err := run(t, tea.WithInput(strings.NewReader("y")))
	require.NoError(t, err)
	assert.True(t, yes)
}

func TestConfirmReturnsErrAbortedOnEsc(t *testing.T) {
	_, err := run(t, tea.WithInput(strings.NewReader("\x1b")))
	assert.ErrorIs(t, err, ErrAborted)
}

func TestConfirmStopsWhenTheContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	pr, pw := io.Pipe()
	t.Cleanup(func() { _ = pw.Close() })
	_, err := run(t, tea.WithInput(pr), tea.WithContext(ctx))
	assert.ErrorIs(t, err, tea.ErrProgramKilled)
}
