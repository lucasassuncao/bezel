package breadcrumb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChosenCarriesThePath(t *testing.T) {
	require.Equal(t, []string{"kv", "app"}, Chosen{Path: []string{"kv", "app"}}.Path)
}
