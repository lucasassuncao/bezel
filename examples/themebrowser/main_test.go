package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/themebrowser"
)

// A test has no terminal, so this is what BrowseInTerminal falls back to.
func TestListPrintsEveryCategory(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, themebrowser.List(&buf))
	t.Log("\n" + buf.String())
	require.Contains(t, buf.String(), "mint")
}
