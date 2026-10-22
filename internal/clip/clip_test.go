package clip

import (
	"errors"
	"testing"

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
