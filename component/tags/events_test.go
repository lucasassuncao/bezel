package tags

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTheEventsCarryTheTag(t *testing.T) {
	require.Equal(t, "go", Added{Tag: "go"}.Tag)
	require.Equal(t, "go", Removed{Tag: "go"}.Tag)
}
