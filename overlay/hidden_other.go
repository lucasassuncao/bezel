//go:build !windows

package overlay

// hiddenAttr is false off Windows: there a leading dot is what hides a file.
func hiddenAttr(string) bool { return false }
