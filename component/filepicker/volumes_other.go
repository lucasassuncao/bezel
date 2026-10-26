//go:build !windows

package filepicker

// listVolumes is nil off Windows: "/" already holds every mount.
func listVolumes() []string { return nil }
