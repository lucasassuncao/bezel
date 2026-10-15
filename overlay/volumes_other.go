//go:build !windows

package overlay

// listVolumes is nil off Windows: "/" already holds every mount.
func listVolumes() []string { return nil }
