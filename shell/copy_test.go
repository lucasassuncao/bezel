package shell

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// run feeds a command's message back into the shell, the way the program would.
func run(t *testing.T, s Shell, cmd func() any) Shell {
	t.Helper()
	require.NotNil(t, cmd)
	s, handled, _ := s.Update(cmd())
	require.True(t, handled, "the copy's answer is the shell's")
	return s
}

func TestCopyWritesTheTextAndSaysDone(t *testing.T) {
	var got string
	s := sized(New(actionConfig())).WithClipboard(func(text string) error { got = text; return nil })

	cmd := s.Copy("kv/app/db", "copied: kv/app/db", "copy path")
	s = run(t, s, func() any { return cmd() })

	require.Equal(t, "kv/app/db", got)
	require.Equal(t, "copied: kv/app/db", s.Status())
	require.Equal(t, OK, s.StatusLevel())
}

func TestCopyReportsTheFailureInTheAppsWords(t *testing.T) {
	s := sized(New(actionConfig())).WithClipboard(func(string) error { return errors.New("xclip: not found") })

	cmd := s.Copy("secret", "copied", "copy policy")
	s = run(t, s, func() any { return cmd() })

	require.Equal(t, "copy policy: xclip: not found", s.Status())
	require.Equal(t, Error, s.StatusLevel())
}

// A helper that never returns must not freeze the program: the copy gives up
// and says what is most likely wrong.
func TestCopyGivesUpOnAHelperThatHangs(t *testing.T) {
	old := clipboardTimeout
	clipboardTimeout = 20 * time.Millisecond
	t.Cleanup(func() { clipboardTimeout = old })

	block := make(chan struct{})
	defer close(block)
	s := sized(New(actionConfig())).WithClipboard(func(string) error { <-block; return nil })

	cmd := s.Copy("x", "copied", "copy")
	s = run(t, s, func() any { return cmd() })

	require.Equal(t, Error, s.StatusLevel())
	require.Contains(t, s.Status(), "did not answer")
}

// The writer is chosen when Copy is called, on the loop, so replacing it
// afterwards cannot race with a copy already under way.
func TestCopyUsesTheWriterSetWhenItWasCalled(t *testing.T) {
	var mu sync.Mutex
	var first, second []string
	s := sized(New(actionConfig())).WithClipboard(func(t string) error { mu.Lock(); first = append(first, t); mu.Unlock(); return nil })
	cmd := s.Copy("a", "copied", "copy")
	s = s.WithClipboard(func(t string) error { mu.Lock(); second = append(second, t); mu.Unlock(); return nil })
	run(t, s, func() any { return cmd() })

	require.Equal(t, []string{"a"}, first)
	require.Empty(t, second)
}

// Error text from a helper process reaches the screen, so it is sanitised.
func TestCopyStatusIsSafeToDraw(t *testing.T) {
	s := sized(New(actionConfig())).WithClipboard(func(string) error { return errors.New("bad\x1b]52;c;aGk=\x07") })
	cmd := s.Copy("x", "copied", "copy")
	s = run(t, s, func() any { return cmd() })
	require.NotContains(t, s.Status(), "\x1b")
}
