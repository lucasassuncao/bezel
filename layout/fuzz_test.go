package layout

import "testing"

// Whatever the constraints and the screen size, every pane lands inside the
// body with a non-negative size, and the panes of a box tile it exactly.
func FuzzResolveTilesTheBody(f *testing.F) {
	f.Add(100, 30, 1, 1, 4, 20, 0, 3, false)
	f.Add(10, 2, 50, 3, 2, 0, 5, 0, true)
	f.Add(0, 0, -4, -1, 0, -9, -3, 99, false)
	f.Fuzz(func(t *testing.T, w, h, header, num, den, minW, maxW, info int, dropSide bool) {
		w, h = mod(w, 500), mod(h, 200)
		cs := []Constraint{Ratio(num%50, den%50), Min(minW % 300)}
		if maxW != 0 {
			cs = append(cs, Max(maxW%300))
		}
		screen := Rows(
			Fixed("header", Lines(header%60)),
			Columns(Fixed("side", cs...), Fill("main"), Fixed("info", Lines(info%40))),
		)
		present := func(name string) bool { return !dropSide || name != "side" }
		body := Rect{X: 3, Y: 2, W: w, H: h}
		rects, _ := Resolve(screen, body, present)

		for name, r := range rects {
			if r.W < 0 || r.H < 0 || r.X < body.X || r.Y < body.Y ||
				r.X+r.W > body.X+body.W || r.Y+r.H > body.Y+body.H {
				t.Fatalf("%s = %+v escapes body %+v", name, r, body)
			}
		}
		if got := rects["header"].H + rects["main"].H; got != h {
			t.Fatalf("rows cover %d of %d: %v", got, h, rects)
		}
		if got := rects["side"].W + rects["main"].W + rects["info"].W; got != w {
			t.Fatalf("columns cover %d of %d: %v", got, w, rects)
		}
	})
}

// mod is n folded into [0, m), negative n included.
func mod(n, m int) int { return (n%m + m) % m }
