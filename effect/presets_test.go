package effect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Each preset plays and lands where its intent says: the block whole, or gone.
func TestPresetsLandWhereTheyMean(t *testing.T) {
	for name, e := range map[string]Effect{
		"enter": Enter(black), "attention": Attention(black), "success": Success(black), "failure": Failure(black),
	} {
		assert.True(t, e.Active(), name)
		assert.Equal(t, white, fgAt(landed(t, e), whiteText), name)
	}
	assert.Equal(t, black, fgAt(landed(t, Exit(black)), whiteText), "exit sinks into bg")
}
