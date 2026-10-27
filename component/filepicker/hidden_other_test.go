//go:build !windows

package filepicker

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHiddenAttrIsUnsetOffWindows(t *testing.T) {
	info, err := os.Stat("/etc/passwd")
	require.NoError(t, err)
	require.False(t, hiddenAttr(info), "a leading dot is the only mark")
}
