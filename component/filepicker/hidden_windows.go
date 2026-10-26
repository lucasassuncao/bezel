//go:build windows

package filepicker

import "golang.org/x/sys/windows"

// hiddenAttr reports the hidden attribute of path, asked for by its full
// path, not by a name the working directory would resolve.
func hiddenAttr(path string) bool {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	attrs, err := windows.GetFileAttributes(p)
	return err == nil && attrs&windows.FILE_ATTRIBUTE_HIDDEN != 0
}
