# Components

One small program per package, each with a recording of it running. Every GIF is a real session of its program, recorded from the `.tape` beside it in [`docs/recordings`](recordings). Run any of them with `go run ./examples/<group>/<name>`, as `go run ./examples/component/list`; for a small app that puts several together, `go run ./examples/app`, explained step by step in [EXAMPLE.md](EXAMPLE.md).

The widgets that go inside the panes live under [`component`](../component); the frame around them, and the key presses for tests, sit at the root. The effects have [EFFECTS.md](EFFECTS.md) and the animations [ANIMATIONS.md](ANIMATIONS.md).

## The frame

How these fit together, from the first `shell.Config` to a dialog of your own, is [FRAME.md](FRAME.md).

What every screen is built in: the shell an app embeds, the layout it resolves, the theme it draws with, and the dialogs, legend and command line around the panes.

### Shell

Source: [`examples/frame/shell/main.go`](../examples/frame/shell/main.go)

Two tabs and two panes. `s` sets a status that clears itself, `b` shows the busy spinner until the work is done, `tab` switches tab and `?` lists every key.

![shell](recordings/frame/shell.gif)

### Layout

Source: [`examples/frame/layout/main.go`](../examples/frame/layout/main.go)

A header over two columns, each pane printing the rect it was given. `w` narrows the window below the `Collapse` limit and the side column drops out, then comes back.

![layout](recordings/frame/layout.gif)

### Theme

Source: [`examples/frame/theme/main.go`](../examples/frame/theme/main.go)

Every colour role and a list drawn with it. `n` and `p` step through a few presets, and `SetTheme` swaps them on the running shell.

![theme](recordings/frame/theme.gif)

### Overlay

Source: [`examples/frame/overlay/main.go`](../examples/frame/overlay/main.go)

The ready-made modals, one per row: an alert, a confirm that answers the app, a prompt that refuses an empty name, the help list, a pager over a command's output and a file picker. `enter` opens the one under the cursor and `esc` closes it.

![overlay](recordings/frame/overlay.gif)

### Draw

Source: [`examples/frame/draw/main.go`](../examples/frame/draw/main.go)

One scrolling pane with a section per helper: header, tabs, breadcrumb, dividers, a card with its actions in the footer, buttons, key-value rows, wrapped and truncated text, numbered lines and the empty state.

![draw](recordings/frame/draw.gif)

### Legend

Source: [`examples/frame/legend/main.go`](../examples/frame/legend/main.go)

More keys than one row holds: the ones cut are counted and `?` lists them all. Five keys need a write capability; `w` grants it and they join the legend.

![legend](recordings/frame/legend.gif)

### Palette

Source: [`examples/frame/palette/main.go`](../examples/frame/palette/main.go)

The `:` command line. A name completes with `tab`, a command takes an argument, and one that is unknown or not allowed says so on the status row.

![palette](recordings/frame/palette.gif)

## Lists, tables and text

What a pane shows: rows with a cursor, a tree, a table, pages of items, a long text to scroll, and a list beside its detail.

### List

Source: [`examples/component/list/main.go`](../examples/component/list/main.go)

Rows under section headings the cursor skips, `enter` showing the payload a row carries, and a `/` filter that takes every key, `q` included. `WithMotion` makes the cursor glide from row to row, a row that arrives light up and one that leaves fade where it was; `Update` still returns no command, so the app calls `Tick` after it and hands the frames to `Animate`.

![list](recordings/component/list.gif)

### Paginator

Source: [`examples/component/paginator/main.go`](../examples/component/paginator/main.go)

Forty-two items, eight to a page, with the page under them. `←` and `→` turn the page, as do `pgup`/`pgdown` and `h`/`l`, and `t` switches between "2/6" and a row of dots.

![paginator](recordings/component/paginator.gif)

### Filepicker

Source: [`examples/component/filepicker/main.go`](../examples/component/filepicker/main.go)

The repository browsed inside a pane, directories first and sizes in a column. `enter` opens a directory or picks a file and names it on the status row, `←` climbs, and `p` shows or hides the permissions column.

![filepicker](recordings/component/filepicker.gif)

### Tree

Source: [`examples/component/tree/main.go`](../examples/component/tree/main.go)

A directory tree over the app's own node type. The arrows expand and collapse; `:reveal` opens every parent on the way down to a path.

![tree](recordings/component/tree.gif)

### Table

Source: [`examples/component/table/main.go`](../examples/component/table/main.go)

Columns sized to their contents, and a cursor over rows that scroll under a fixed header, with the position at the foot of the pane, a scrollbar down its edge and the row's detail below. `pgdown` and `pgup` move a page at a time, `enter` reports the chosen row, and `w` narrows the window: the description gives way first, and the version and licence stay whole.

![table](recordings/component/table.gif)

### Viewport

Source: [`examples/component/viewport/main.go`](../examples/component/viewport/main.go)

An access log, numbered in a gutter, with its error lines in red and every `ERROR` highlighted. The arrows, `pgup`/`pgdown` and `home`/`end` scroll, `←` and `→` pan the long lines, `n` and `N` go from error to error, `w` wraps the lines, and the mouse wheel scrolls.

![viewport](recordings/component/viewport.gif)

### Browser

Source: [`examples/component/browser/main.go`](../examples/component/browser/main.go)

Recipe names on the left, the selected recipe on the right. `tab` moves the keys to the recipe so it scrolls, and back.

![browser](recordings/component/browser.gif)

### Accordion

Source: [`examples/component/accordion/main.go`](../examples/component/accordion/main.go)

A column of questions, each a heading that opens onto its answer, sliding rather than jumping. The arrows move between headings and `enter` or `space` opens and closes the one under the cursor; a section closing half way closes from where it is.

![accordion](recordings/component/accordion.gif)

## Input

What the user types into: a form of controls, a multi-line field, and the caret any field of your own can draw.

### Form

Source: [`examples/component/form/main.go`](../examples/component/form/main.go)

One of each control: a name typed into an `Input`, an image picked from a `Select` that opens under it, a `Radio`, a `Toggle` and a `Checkbox`, then the `Button` that submits. `tab` and `shift+tab` move between them; Create reports the values on the status row.

![form](recordings/component/form.gif)

### Textbox

Source: [`examples/component/textbox/main.go`](../examples/component/textbox/main.go)

A multi-line value with its lines numbered. Every key is text, so the app quits on `esc`; `ctrl+s` counts the lines. Shift with the arrows selects, and so does dragging the mouse: the shell's `MouseIn` moves the pointer into the pane's cells.

![textbox](recordings/component/textbox.gif)

### Cursor

Source: [`examples/component/cursor/main.go`](../examples/component/cursor/main.go)

Three cursors, one per mode, each on the first letter of a word. `tab` moves the focus down the rows, and `m` turns the focused cursor from blinking to steady to hidden and around again.

![cursor](recordings/component/cursor.gif)

## Progress, time and state

How far along something is, how long it took, and what state it is in.

### Progress

Source: [`examples/component/progress/main.go`](../examples/component/progress/main.go)

A count that fills on its own, as a `Line` and as a bare `Bar`, and a wizard's stages in a `Stepper`. `n` and `p` walk the stages; the copy drawn at 30 columns folds to "Step 2 of 4". `r` restarts the count.

![progress](recordings/component/progress.gif)

### Spinner

Source: [`examples/component/spinner/main.go`](../examples/component/spinner/main.go)

The twelve presets side by side, each turning at its own pace. `space` stops them all and starts them again.

![spinner](recordings/component/spinner.gif)

### Timer

Source: [`examples/component/timer/main.go`](../examples/component/timer/main.go)

Ten seconds counted down a tenth at a time, written to a tenth with `WithFormat`. `space` stops and resumes it, `r` starts it over, and the status row says when time is up.

![timer](recordings/component/timer.gif)

### Stopwatch

Source: [`examples/component/stopwatch/main.go`](../examples/component/stopwatch/main.go)

Time counted up a tenth at a time. `space` starts and stops it, and `r` puts it back to zero, running or not.

![stopwatch](recordings/component/stopwatch.gif)

### Icon

Source: [`examples/component/icon/main.go`](../examples/component/icon/main.go)

Every state mark in both sets, then a list of packages and a source chain drawn as a tree. `a` swaps them between Unicode and ASCII, the set `BEZEL_ASCII` or `TERM=dumb` picks on its own.

![icon](recordings/component/icon.gif)

### Toast

Source: [`examples/component/toast/main.go`](../examples/component/toast/main.go)

Notices that slide into the top right corner, stack with the newest on top, and leave on their own after three seconds, each edged in the colour of its kind. `Over(screen, th)` draws them on top of whatever the shell drew. `s`, `c`, `w` and `e` push one of each kind.

![toast](recordings/component/toast.gif)

## Outside a full screen

For a CLI that prints logs rather than takes the screen, and for the tests that drive a model.

### Inline

Source: [`examples/component/inline/main.go`](../examples/component/inline/main.go)

A CLI that prints logs, asks one question below them, then redraws a progress bar in place. Nothing takes over the screen.

![inline](recordings/component/inline.gif)

### Themebrowser

Source: [`examples/component/themebrowser/main.go`](../examples/component/themebrowser/main.go)

The table a `--list-themes` flag prints: scrollable in a terminal, plain lines in a pipe.

![themebrowser](recordings/component/themebrowser.gif)

### Bezeltest

Source: [`examples/frame/bezeltest/main_test.go`](../examples/frame/bezeltest/main_test.go)

No recording: this one is a test. A counter, and the test that drives it with the key presses a terminal would send.

## Recording

The tapes live in `docs/recordings`, beside the GIFs they make, and are written for Windows: each sets `Set Shell "powershell"`, builds a `.exe` off camera and deletes it when done. Every example also starts and quits off camera, so its GIF opens and loops on the program; the two that print into the terminal rather than take it over, `inline` and `themebrowser`, run behind a `clear` so their command line never shows. The galleries, of animation, text, scenes and effects, have a tape per tab. They draw 1000 px wide in JetBrainsMono Nerd Font at 16px, so GitHub shows them close to full size, and 640 px tall unless the example needs another height. With [VHS](https://github.com/charmbracelet/vhs), `ttyd` and `ffmpeg` on the `PATH` and the font installed, from the repository root:

```sh
vhs docs/recordings/component/list.tape
```

From the root, always: `Output` resolves against the working directory, not against the tape. Without the font, VHS falls back to another monospace and the recording drifts from the others.
