package effect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Each preset plays and lands where its intent says: the block whole, or gone.
func TestPresetsLandWhereTheyMean(t *testing.T) {
	for name, e := range map[string]Effect{
		"enter": Enter(black), "attention": Attention(black), "success": Success(black), "failure": Failure(black),
	} {
		require.True(t, e.Active(), name)
		require.Equal(t, white, fgAt(landed(t, e), whiteText), name)
	}
	require.Equal(t, black, fgAt(landed(t, Exit(black)), whiteText), "exit sinks into bg")
}
