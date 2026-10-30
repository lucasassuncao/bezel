package table

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChosenCarriesACopyOfTheItem(t *testing.T) {
	it := Item[string]{Cells: []string{"a"}, Value: "x"}
	require.Equal(t, it, Chosen[string]{Item: it}.Item)
}
