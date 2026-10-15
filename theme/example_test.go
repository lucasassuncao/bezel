package theme_test

import (
	"fmt"

	"github.com/lucasassuncao/bezel/theme"
)

// Lookup backs a --theme flag: names are case-insensitive, empty is Default.
func ExampleLookup() {
	t, err := theme.Lookup("BANANA")
	fmt.Println(err == nil, t.Colors == theme.Banana.Colors)
	_, err = theme.Lookup("bananna")
	fmt.Println(err)
	// Output:
	// true true
	// unknown theme "bananna"
}
