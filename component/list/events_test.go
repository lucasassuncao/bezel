package list

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChosenCarriesACopyOfTheRow(t *testing.T) {
	row := Row[string]{Label: "a", Value: "x"}
	require.Equal(t, row, Chosen[string]{Row: row}.Row)
	require.Equal(t, Chosen[int]{}, Chosen[int]{})
}
