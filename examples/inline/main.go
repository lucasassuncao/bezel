// Inline: a CLI that prints logs, asks one question below them and draws a
// progress bar in place; nothing takes over the screen.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/lucasassuncao/bezel/inline"
	"github.com/lucasassuncao/bezel/theme"
)

const files = 20

// run is main with its terminal passed in, so a test can answer for the user.
func run(w io.Writer, th theme.Resolved, confirm func(string) (bool, error), pause time.Duration) error {
	fmt.Fprintln(w, "fetching index... ok")
	fmt.Fprintf(w, "%d files to download (48 MB)\n", files)
	fmt.Fprintln(w, "target: ./cache")

	ok, err := confirm("Download them now?")
	if errors.Is(err, inline.ErrAborted) || (err == nil && !ok) {
		fmt.Fprintln(w, "skipped")
		return nil
	}
	if err != nil {
		return err
	}
	for i := 0; i <= files; i++ {
		// \r redraws the bar on its own line instead of printing a new one.
		fmt.Fprintf(w, "\r%s %2d/%d", inline.Bar(i, files, 30, th), i, files)
		time.Sleep(pause)
	}
	fmt.Fprintln(w, "\ndone")
	return nil
}

func main() {
	th := theme.Resolve(theme.ThemeMint, true)
	confirm := func(q string) (bool, error) { return inline.Confirm(q, th) }
	if err := run(os.Stdout, th, confirm, 120*time.Millisecond); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
