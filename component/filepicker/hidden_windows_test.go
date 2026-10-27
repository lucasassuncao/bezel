//go:build windows

package filepicker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// stat is the info of the one entry ReadDir finds in dir.
func stat(t *testing.T, dir string) os.FileInfo {
	t.Helper()
	des, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, des, 1)
	info, err := des[0].Info()
	require.NoError(t, err)
	return info
}

func TestHiddenAttrReadsWhatReadDirFetched(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.txt")
	require.NoError(t, os.WriteFile(path, nil, 0o600))
	require.False(t, hiddenAttr(stat(t, dir)))

	p, err := windows.UTF16PtrFromString(path)
	require.NoError(t, err)
	require.NoError(t, windows.SetFileAttributes(p, windows.FILE_ATTRIBUTE_HIDDEN))
	require.True(t, hiddenAttr(stat(t, dir)))
}
