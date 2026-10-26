//go:build windows

package filepicker

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// Each volume is a drive root such as `C:\`, in letter order, and the drive
// Windows runs from is always one of them.
func TestListVolumesHasTheSystemDrive(t *testing.T) {
	got := listVolumes()
	for _, v := range got {
		require.Regexp(t, `^[A-Z]:\\$`, v)
	}
	require.IsIncreasing(t, got)
	require.Contains(t, got, os.Getenv("SystemDrive")+`\`)
}
