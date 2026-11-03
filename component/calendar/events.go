package calendar

import "time"

// Chosen is the event of a day chosen: enter, or a click on the day already
// under the cursor. Day is at midnight.
type Chosen struct{ Day time.Time }
