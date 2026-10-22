//go:build !windows

package overlay

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHiddenAttrIsUnsetOffWindows(t *testing.T) {
	require.False(t, hiddenAttr("/etc/passwd"), "a leading dot is the only mark")
}
