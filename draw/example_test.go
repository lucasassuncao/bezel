package draw_test

import (
	"fmt"

	"github.com/lucasassuncao/bezel/draw"
)

func ExampleFit() {
	fmt.Printf("[%s]\n", draw.Fit("a long title", 8))
	fmt.Printf("[%s]\n", draw.Fit("short", 8))
	// Output:
	// [a long …]
	// [short   ]
}

func ExampleWrap() {
	for _, line := range draw.Wrap("the quick brown fox jumps over the lazy dog", 15) {
		fmt.Printf("[%s]\n", line)
	}
	// Output:
	// [the quick brown]
	// [fox jumps over]
	// [the lazy dog]
}

func ExampleComposite() {
	bg := "..........\n..........\n.........."
	fmt.Println(draw.Composite("[ok]", bg, 3, 1))
	// Output:
	// ..........
	// ...[ok]...
	// ..........
}

// Control sequences from untrusted text are shown, never executed.
func ExampleSanitize() {
	fmt.Println(draw.Sanitize("name\x1b]52;c;aGk=\x07\n"))
	// Output:
	// name\e]52;c;aGk=\x07\n
}
