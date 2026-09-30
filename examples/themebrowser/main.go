// Themebrowser: the table a --list-themes flag prints, scrollable in a
// terminal and plain text in a pipe.
package main

import (
	"fmt"
	"os"

	"github.com/lucasassuncao/bezel/theme"
	"github.com/lucasassuncao/bezel/themebrowser"
)

func main() {
	if err := themebrowser.BrowseInTerminal(theme.ThemeMint); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
