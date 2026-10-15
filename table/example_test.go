package table_test

import (
	"fmt"

	"github.com/lucasassuncao/bezel/table"
)

// The Flex column gives way first when the rows do not fit.
func ExampleFit() {
	cols := []table.Column{
		{Title: "Name", Min: 4, Flex: true},
		{Title: "Source"},
		{Title: "Version"},
	}
	rows := [][]string{
		{"Google Chrome", "winget", "126.0.6478.127"},
		{"Visual Studio Code", "scoop", "1.91"},
	}
	widths := table.Fit(cols, rows, 38)
	fmt.Println(table.Titles(cols, widths))
	for _, r := range rows {
		fmt.Println(table.Row(widths, r))
	}
	// Output:
	// Name            Source  Version
	// Google Chrome   winget  126.0.6478.127
	// Visual Studio…  scoop   1.91
}
