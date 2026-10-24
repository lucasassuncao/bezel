package frames

import (
	"math"
	"os"
	"strconv"
	"time"
)

// Still reports whether motion is switched off: BEZEL_NO_MOTION set to
// anything, TERM=dumb, or BEZEL_MOTION at 0.
func Still() bool {
	return os.Getenv("BEZEL_NO_MOTION") != "" || os.Getenv("TERM") == "dumb" || Speed() <= 0
}

// Speed is how fast motion plays, from BEZEL_MOTION: 2 twice as fast, 0.5 at
// half speed, 1 when it is unset or not a number.
func Speed() float64 {
	s, err := strconv.ParseFloat(os.Getenv("BEZEL_MOTION"), 64)
	if err != nil || math.IsNaN(s) || math.IsInf(s, 0) {
		return 1
	}
	return s
}

// Paced is d of wall time as motion at Speed lives it.
func Paced(d time.Duration) time.Duration {
	if s := Speed(); s > 0 {
		return time.Duration(float64(d) * s)
	}
	return d
}
