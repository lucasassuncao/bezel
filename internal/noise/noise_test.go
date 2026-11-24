package noise

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashIsTheSameForTheSameSeedAndIndex(t *testing.T) {
	require.Equal(t, Hash(7, 3), Hash(7, 3))
	require.NotEqual(t, Hash(7, 3), Hash(8, 3))
	require.NotEqual(t, Hash(7, 3), Hash(7, 4))
}

func TestUnitStaysBelowOne(t *testing.T) {
	require.Equal(t, 0.0, Unit(0))
	require.Less(t, Unit(^uint64(0)), 1.0)
}
