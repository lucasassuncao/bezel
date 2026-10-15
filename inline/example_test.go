package inline_test

import (
	"fmt"
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/inline"
	"github.com/lucasassuncao/bezel/theme"
)

// Program options answer the question from a reader, as a test would.
func ExampleConfirm() {
	th := theme.Resolve(theme.Plain, true)
	yes, err := inline.Confirm("Delete 3 files?", th,
		tea.WithInput(strings.NewReader("y")), tea.WithOutput(io.Discard))
	fmt.Println(yes, err)
	// Output:
	// true <nil>
}
