package table

// Chosen is the event of a row being chosen: enter, or a click on the row
// already under the cursor. Item is a copy of that row.
type Chosen[T any] struct{ Item Item[T] }

// Toggled is the event of rows checked or unchecked while the table is
// WithMulti: one row for Check, every one that changed for CheckAll.
type Toggled[T any] struct{ Items []Item[T] }

// Sorted is the event of the rows sorted again: Sort, or a click on a
// column's title, which sorts by it and then the other way.
type Sorted struct {
	Column int
	Desc   bool
}
