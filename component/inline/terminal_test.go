package inline

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/form"
)

func noTTY() (*os.File, *os.File, error) { return nil, nil, errors.New("no tty") }

func TestATerminalOnBothEndsIsLeftAlone(t *testing.T) {
	opened := false
	base, done := terminal(true, true, func() (*os.File, *os.File, error) { opened = true; return noTTY() })
	done()
	require.Nil(t, base)
	require.False(t, opened, "no TTY is opened when stdin and stdout are one")
}

func TestARedirectedEndAsksOnTheTTY(t *testing.T) {
	dir := t.TempDir()
	in, err := os.Create(filepath.Join(dir, "in"))
	require.NoError(t, err)
	out, err := os.Create(filepath.Join(dir, "out"))
	require.NoError(t, err)
	base, done := terminal(true, false, func() (*os.File, *os.File, error) { return in, out, nil })
	require.Len(t, base, 2)
	done()
	_, err = out.WriteString("x")
	require.Error(t, err, "the TTY is closed once the question is over")
}

func TestWithNoTerminalTheQuestionFailsAtOnce(t *testing.T) {
	base, done := terminal(false, false, noTTY)
	defer done()
	_, err := tea.NewProgram(newConfirm("Proceed?", th), append(base, tea.WithoutSignals())...).Run()
	require.ErrorIs(t, err, ErrNoTerminal)
}

func TestConfirmAndAskReturnErrNoTerminalItself(t *testing.T) {
	opts := []tea.ProgramOption{tea.WithInput(noTerminal{}), tea.WithOutput(io.Discard), tea.WithoutSignals()}
	_, err := Confirm("Proceed?", th, opts...)
	require.Equal(t, ErrNoTerminal, err, "not wrapped in the program's kill")
	in, err := Ask(form.NewInput("Name").WithValue("api"), th, opts...)
	require.Equal(t, ErrNoTerminal, err)
	require.Equal(t, "api", in.Value(), "the field comes back as it was given")
}
