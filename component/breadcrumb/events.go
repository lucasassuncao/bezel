package breadcrumb

// Chosen is the event of a path chosen: enter, or a click on the part
// already under the cursor. Path runs from the first part to that one.
type Chosen struct{ Path []string }
