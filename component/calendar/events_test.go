package calendar

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestChosenCarriesTheDay(t *testing.T) {
	d := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	require.Equal(t, d, Chosen{Day: d}.Day)
}
