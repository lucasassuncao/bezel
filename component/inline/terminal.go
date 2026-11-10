package inline

import (
	"errors"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"
)

// ErrNoTerminal is what Confirm and Ask return when no terminal can answer
// them, as in CI or a cron job: the app falls back to a flag or a default.
var ErrNoTerminal = errors.New("inline: no terminal to ask on")

// onTerminal asks m on the user's terminal, even when stdin or stdout is redirected;
// opts come last, so a test's WithInput or WithOutput wins.
func onTerminal(m tea.Model, opts []tea.ProgramOption) (tea.Model, error) {
	base, done := terminal(isTerminal(os.Stdin), isTerminal(os.Stdout), tea.OpenTTY)
	defer done()
	final, err := tea.NewProgram(m, append(base, opts...)...).Run()
	if errors.Is(err, ErrNoTerminal) {
		return final, ErrNoTerminal
	}
	return final, err
}

func isTerminal(f *os.File) bool { return term.IsTerminal(f.Fd()) }

// terminal is where a question is drawn and read when stdin or stdout is not
// a terminal: the TTY when one opens, or else input that fails with
// ErrNoTerminal at once, so the question never waits on keys that cannot come.
func terminal(stdin, stdout bool, open func() (*os.File, *os.File, error)) ([]tea.ProgramOption, func()) {
	if stdin && stdout {
		return nil, func() {}
	}
	in, out, err := open()
	if err != nil {
		return []tea.ProgramOption{tea.WithInput(noTerminal{}), tea.WithOutput(io.Discard)}, func() {}
	}
	return []tea.ProgramOption{tea.WithInput(in), tea.WithOutput(out)}, func() {
		_ = in.Close()
		_ = out.Close()
	}
}

// noTerminal is input with nobody behind it.
type noTerminal struct{}

func (noTerminal) Read([]byte) (int, error) { return 0, ErrNoTerminal }
