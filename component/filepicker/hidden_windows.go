//go:build windows

package filepicker

import (
	"os"
	"syscall"
)

// hiddenAttr reports the hidden attribute ReadDir already fetched with info,
// so a listing costs no extra call per entry.
func hiddenAttr(info os.FileInfo) bool {
	d, ok := info.Sys().(*syscall.Win32FileAttributeData)
	return ok && d.FileAttributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0
}
