package filepicker

// Chosen is the event of an entry picked: enter, or a click on the entry
// already under the cursor. Path is absolute.
type Chosen struct{ Path string }
