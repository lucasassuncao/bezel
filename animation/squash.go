package animation

import "math"

// contact is how near the floor, in cells, a falling Drop is drawn squashed.
const contact = 0.35

// Shape is how d should be drawn now, squash and stretch, as factors of its
// resting width and height: taller and narrower as it falls fast, wide and
// flat as it hits the floor, round at rest. Their product stays near 1.
func (d Drop) Shape() (w, h float64) {
	if !d.Active() || d.from == d.to {
		return 1, 1
	}
	const dt = 1e-3
	before, _ := d.at(max(d.t-dt, 0))
	after, _ := d.at(d.t + dt)
	speed := math.Abs(after-before) / (2 * dt)
	top := math.Sqrt(2 * d.gravity * math.Abs(d.to-d.from)) // speed at the first hit
	share := min(speed/top, 1)
	if math.Abs(d.Cur-d.to) < contact && share > 0.2 {
		h = 1 - 0.45*share
		return 1 / h, h
	}
	h = 1 + 0.5*share
	return 1 / h, h
}
