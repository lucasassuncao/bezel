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

// Window is the part of a list that fits in Height rows, from Offset, kept on
// Cursor. While rows overflow below it, its last row says how many more
// there are instead of showing one, the way list and tree draw "↓ N more".
type Window struct{ Cursor, Offset, Height int }

// Clamp fits w to n rows: the cursor on one of them, no blank rows under a
// list scrolled past its end, and the cursor clear of the "more" row.
func (w Window) Clamp(n int) Window {
	if n <= 0 {
		return Window{Height: w.Height}
	}
	w.Cursor = min(max(w.Cursor, 0), n-1)
	w.Offset = ClampScroll(w.Cursor, min(w.Offset, max(n-w.Height, 0)), w.Height)
	if w.more(n) && w.Cursor >= w.Offset+w.Height-1 {
		w.Offset = w.Cursor - w.Height + 2
	}
	return w
}

// Move steps the cursor delta rows of n, passing over the ones skip reports
// (nil skips none) and stopping at the last row it can land on.
func (w Window) Move(delta, n int, skip func(i int) bool) Window {
	if n <= 0 {
		return w.Clamp(n)
	}
	w.Cursor = min(max(w.Cursor, 0), n-1)
	dir := 1
	if delta < 0 {
		dir, delta = -1, -delta
	}
	for ; delta > 0; delta-- {
		next := w.Cursor + dir
		for next >= 0 && next < n && skip != nil && skip(next) {
			next += dir
		}
		if next < 0 || next >= n {
			break
		}
		w.Cursor = next
	}
	return w.Clamp(n)
}

// Bounds is the rows [start, end) of n that w shows, and the count its "more"
// row reports, zero when it has none. Clamp w first.
func (w Window) Bounds(n int) (start, end, below int) {
	start = min(max(w.Offset, 0), n)
	visible := w.Height
	if w.more(n) {
		visible--
	}
	end = min(start+max(visible, 0), n)
	if w.more(n) {
		below = n - end
	}
	return start, end, below
}

// more reports rows overflowing below w. A one-row window never trades its
// only row for the count.
func (w Window) more(n int) bool { return w.Height > 1 && w.Offset+w.Height < n }

// Scroll is how much of Total rows a window shows, from Offset: what a
// scrollbar draws. The zero value draws none.
type Scroll struct{ Offset, Visible, Total int }

// Scroll is w over n rows, as a scrollbar sees it: the "more" row is not one
// of the rows shown. Clamp w first.
func (w Window) Scroll(n int) Scroll {
	start, end, _ := w.Bounds(n)
	return Scroll{Offset: start, Visible: end - start, Total: n}
}

// Thumb places a scrollbar's thumb on a track rows long: from pos, size rows,
// and size 0 when every row shows. The thumb touches an end of the track only
// while the window does, so it never reads as the top when it is not.
func (s Scroll) Thumb(track int) (pos, size int) {
	if track <= 0 || s.Visible <= 0 || s.Total <= s.Visible {
		return 0, 0
	}
	size = max(1, (track*s.Visible+s.Total/2)/s.Total)
	if track > 1 {
		size = min(size, track-1) // a full track would say nothing is hidden
	}
	room, last := track-size, s.Total-s.Visible
	off := min(max(s.Offset, 0), last)
	pos = (off*room + last/2) / last
	if room >= 2 {
		if off > 0 {
			pos = max(pos, 1)
		}
		if off < last {
			pos = min(pos, room-1)
		}
	}
	return pos, size
}
