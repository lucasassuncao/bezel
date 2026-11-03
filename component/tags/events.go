package tags

// Added is the event of the typed text becoming a tag.
type Added struct{ Tag string }

// Removed is the event of the last tag taken off.
type Removed struct{ Tag string }
