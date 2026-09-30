package layout

// ClampScroll moves offset so that cursor stays inside a window of height
// rows: up when the cursor is above it, down when below. Never negative; a
// non-positive height only applies the floor.
func ClampScroll(cursor, offset, height int) int {
	if cursor < offset {
		offset = cursor
	}
	if height > 0 && cursor >= offset+height {
		offset = cursor - height + 1
	}
	return max(offset, 0)
}

// ScrollStart keeps the cursor inside a window of height rows without
// jumping around: centred on it where the list allows.
func ScrollStart(cursor, total, height int) int {
	if total <= height || height <= 0 {
		return 0
	}
	return min(max(cursor-height/2, 0), total-height)
}
