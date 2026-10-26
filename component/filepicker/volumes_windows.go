//go:build windows

package filepicker

import "golang.org/x/sys/windows"

// listVolumes is every drive letter in use, as roots such as `C:\`. It asks
// for the bitmask only, so an empty card reader or a lost share costs nothing.
func listVolumes() []string {
	mask, err := windows.GetLogicalDrives()
	if err != nil {
		return nil
	}
	var out []string
	for i := range 26 {
		if mask&(1<<i) != 0 {
			out = append(out, string(rune('A'+i))+`:\`)
		}
	}
	return out
}
