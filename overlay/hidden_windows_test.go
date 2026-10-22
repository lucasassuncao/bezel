//go:build windows

package overlay

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// The attribute is read where the file is, whatever the working directory.
func TestHiddenAttrReadsTheFileItIsGiven(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.txt")
	require.NoError(t, os.WriteFile(path, nil, 0o600))
	require.False(t, hiddenAttr(path))

	p, err := windows.UTF16PtrFromString(path)
	require.NoError(t, err)
	require.NoError(t, windows.SetFileAttributes(p, windows.FILE_ATTRIBUTE_HIDDEN))
	require.True(t, hiddenAttr(path))
	require.False(t, hiddenAttr(filepath.Join(dir, "missing")), "no file, nothing hidden")
}
