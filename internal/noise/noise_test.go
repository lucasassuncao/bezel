package noise

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashIsTheSameForTheSameSeedAndIndex(t *testing.T) {
	assert.Equal(t, Hash(7, 3), Hash(7, 3))
	assert.NotEqual(t, Hash(7, 3), Hash(8, 3))
	assert.NotEqual(t, Hash(7, 3), Hash(7, 4))
}

func TestUnitStaysBelowOne(t *testing.T) {
	assert.Equal(t, 0.0, Unit(0))
	assert.Less(t, Unit(^uint64(0)), 1.0)
}
