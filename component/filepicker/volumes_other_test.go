//go:build !windows

package filepicker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListVolumesIsEmptyOffWindows(t *testing.T) {
	require.Nil(t, listVolumes(), `"/" already holds every mount`)
}
