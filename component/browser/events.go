package browser

// Chosen is the event of an item being chosen: enter on the list. Item is
// the item under the cursor.
type Chosen struct{ Item Item }

// Dismissed is the event of esc on the list: close without choosing.
type Dismissed struct{}
