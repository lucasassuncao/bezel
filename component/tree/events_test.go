package tree

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandedAndCollapsedCarryACopyOfTheNode(t *testing.T) {
	n := node{label: "a", expanded: true}
	require.Equal(t, n, Expanded[node]{Node: n}.Node)
	require.Equal(t, n, Collapsed[node]{Node: n}.Node)
}
