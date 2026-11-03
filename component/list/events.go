package list

// Chosen is the event of a row being chosen: enter, or a click on the row
// already under the cursor. Row is a copy of that row.
type Chosen[T any] struct{ Row Row[T] }

// Toggled is the event of rows checked or unchecked while the list is
// WithMulti: one row for Check, every one that changed for CheckAll.
type Toggled[T any] struct{ Rows []Row[T] }
