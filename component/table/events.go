package table

// Chosen is the event of a row being chosen: enter, or a click on the row
// already under the cursor. Item is a copy of that row.
type Chosen[T any] struct{ Item Item[T] }
