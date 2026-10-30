package tree

// Expanded is the event of a parent opening: right, or a click on the
// parent under the cursor. Node is a copy of it, open.
type Expanded[N any] struct{ Node N }

// Collapsed is the event of a parent closing: left, or a click on the
// parent under the cursor. Node is a copy of it, closed.
type Collapsed[N any] struct{ Node N }
