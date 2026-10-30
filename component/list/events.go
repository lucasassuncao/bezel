package list

// Chosen is the event of a row being chosen: enter, or a click on the row
// already under the cursor. Row is a copy of that row.
type Chosen[T any] struct{ Row Row[T] }
