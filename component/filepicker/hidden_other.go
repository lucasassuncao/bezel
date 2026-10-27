//go:build !windows

package filepicker

import "os"

// hiddenAttr is false off Windows: there a leading dot is what hides a file.
func hiddenAttr(os.FileInfo) bool { return false }
