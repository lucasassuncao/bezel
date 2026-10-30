package browser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChosenCarriesTheItemAndDismissedNothing(t *testing.T) {
	require.Equal(t, "a", Chosen{Item: Item{Label: "a"}}.Item.Label)
	require.Equal(t, Dismissed{}, Dismissed{})
}
