package clip

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestPasteAnswersWithTheClipboardsText(t *testing.T) {
	msg := Paste(func() (string, error) { return "two\nlines", nil })()
	require.Equal(t, tea.PasteMsg{Content: "two\nlines"}, msg)
}

// Nothing to paste is no message at all: a field gets no empty paste.
func TestPasteOfNothingIsNoMessage(t *testing.T) {
	require.Nil(t, Paste(func() (string, error) { return "", nil })())
	require.Nil(t, Paste(func() (string, error) { return "", errors.New("no xclip") })())
}

// A helper that hangs costs a paste, not the program: the command gives up
// in time, and the read's goroutine ends whenever the helper does.
func TestPasteGivesUpOnAReadThatHangs(t *testing.T) {
	old := timeout
	timeout = 10 * time.Millisecond
	t.Cleanup(func() { timeout = old })

	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	cmd := Paste(func() (string, error) {
		<-release
		return "late", nil
	})
	start := time.Now()
	require.Nil(t, cmd())
	require.Less(t, time.Since(start), time.Second)
}
