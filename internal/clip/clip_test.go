package clip

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestPasteAnswersWithTheClipboardsText(t *testing.T) {
	msg := Paste(func() (string, error) { return "two\nlines", nil }, 0)()
	require.Equal(t, tea.PasteMsg{Content: "two\nlines"}, msg)
}

// Nothing to paste is no message at all: a field gets no empty paste.
func TestPasteOfNothingIsNoMessage(t *testing.T) {
	require.Nil(t, Paste(func() (string, error) { return "", nil }, 0)())
	require.Nil(t, Paste(func() (string, error) { return "", errors.New("no xclip") }, 0)())
}

// A helper that hangs costs a paste, not the program: the command gives up
// in time, and the read's goroutine ends whenever the helper does.
func TestPasteGivesUpOnAReadThatHangs(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	cmd := Paste(func() (string, error) {
		<-release
		return "late", nil
	}, 10*time.Millisecond)
	start := time.Now()
	require.Nil(t, cmd())
	require.Less(t, time.Since(start), time.Second)
}

func TestCopyWritesAndSaysHowItWent(t *testing.T) {
	var got string
	done := func(err error) tea.Msg { return err }
	require.Nil(t, Copy(func(s string) error { got = s; return nil }, "text", 0, done)())
	require.Equal(t, "text", got)
	require.EqualError(t, Copy(func(string) error { return errors.New("no xclip") }, "x", 0, done)().(error), "no xclip")
}

func TestCopyGivesUpOnAWriteThatHangs(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	msg := Copy(func(string) error { <-release; return nil }, "x", 10*time.Millisecond, func(err error) tea.Msg { return err })()
	require.ErrorIs(t, msg.(error), ErrTimeout)
}
