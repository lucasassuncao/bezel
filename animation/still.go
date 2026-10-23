package animation

import "os"

// Still reports whether motion is switched off: BEZEL_NO_MOTION set to
// anything, or TERM=dumb. Then every animation is born where it would end,
// for a user who finds motion hard to follow, a slow link or a CI log.
func Still() bool { return os.Getenv("BEZEL_NO_MOTION") != "" || os.Getenv("TERM") == "dumb" }
