// Package noise is the hash the animations draw their randomness from: the
// same seed and index give the same number, so a frame draws the same twice.
package noise

// Hash is a hash of seed and k (splitmix64).
func Hash(seed, k uint64) uint64 {
	z := seed + k*0x9E3779B97F4A7C15
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

// Unit is n as a share in [0,1).
func Unit(n uint64) float64 { return float64(n>>11) / (1 << 53) }
