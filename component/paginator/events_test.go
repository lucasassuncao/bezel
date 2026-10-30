package paginator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPagedCarriesThePageFromZero(t *testing.T) {
	require.Equal(t, 2, Paged{Page: 2}.Page)
	require.Equal(t, Paged{}, Paged{})
}
