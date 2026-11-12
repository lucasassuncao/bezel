# Packages

Every package in bezel, grouped by what it is for. Each one also has a `README.md` generated from its doc comments, and its API on [pkg.go.dev](https://pkg.go.dev/github.com/lucasassuncao/bezel).

**The frame**, at the root: what every screen is built in.

- [`shell`](../shell): the struct you embed. Tabs, focus, status, a busy spinner, overlays and the legend; keys are `Action`s that drive the key, the legend and the palette at once.
- [`layout`](../layout): the screen as a tree of `Columns`, `Rows`, `Fixed` and `Fill`, resolved to rects. `Collapse` drops a pane below a width.
- [`theme`](../theme): semantic colours, sixteen presets, and the styles every package draws with.
- [`overlay`](../overlay): a stack of modals: `Alert`, `Confirm`, `Prompt`, `Pick`, `Choice`, `Help`, `Pager`, `FilePicker`.
- [`draw`](../draw): ANSI-safe text fitting, panels (text and a scrollbar in their edges), cards, header, tabs, dividers and buttons.
- [`legend`](../legend): key hints, filtered by what the session allows and packed into a few rows.
- [`key`](../key): key bindings, with the help the legend prints; a `Binding` is a value, switched off by its `Disabled` field.
- [`bezeltest`](../bezeltest): key presses for tests that drive a model, and `Golden` to check a whole screen against a file.

**Components**, in [`component`](../component): what goes inside the panes.

- [`list`](../component/list): a cursor list with sections, a `/` filter and rows checked for a batch.
- [`calendar`](../component/calendar): a month with a cursor on a day, and the calendar a form's date field opens.
- [`breadcrumb`](../component/breadcrumb): a path in one row, `kv › apps › prod`, with a cursor on a part and its middle giving way when narrow.
- [`paginator`](../component/paginator): pages as dots or "2/5", and the slice of items each page shows.
- [`tree`](../component/tree): an expandable tree, generic over your node type; `Static` prints one built from nested branches.
- [`table`](../component/table): columns sized to their contents within a width, and a cursor over rows that scroll under a fixed header, sorted by a column and checked for a batch; `Static` prints the whole table.
- [`form`](../component/form): `Input`, `Select`, `MultiSelect`, `Date`, `Number`, `Slider`, `Tags`, `Checkbox`, `Radio`, `Toggle` and `Button`, lined up in one column.
- [`wizard`](../component/wizard): forms split into pages, the steps drawn on top and a check before each page is left.
- [`tags`](../component/tags): a field that turns what is typed into chips, for filters and labels.
- [`input`](../component/input): a one-line text field with readline keys, a placeholder, a mask for secrets and a format such as `9999-99-99`.
- [`textbox`](../component/textbox): a multi-line text field with line numbers.
- [`viewport`](../component/viewport): scrollable text, wrapped or panned, with line numbers and search highlights.
- [`cursor`](../component/cursor): a caret that blinks, stays or hides, for fields of your own.
- [`browser`](../component/browser): a list on the left, the selected item's detail on the right.
- [`accordion`](../component/accordion): sections that open and close onto their text, sliding.
- [`progress`](../component/progress): `Bar`, `Line`, `Smooth`, `Meter` and `Stepper`, each one row.
- [`spinner`](../component/spinner): twelve spinners and your own frames, each ticking on its own.
- [`timer`](../component/timer): a countdown that says when time is up.
- [`stopwatch`](../component/stopwatch): time counted up, paused and reset.
- [`chart`](../component/chart): `Area` and `Spark` charts of a run of numbers, braille `Lines` of several runs over a y axis, labelled `Bars` side by side, a `Heat` grid and a `Donut`.
- [`canvas`](../component/canvas): free drawing in braille: lines, boxes, circles, points and labels in your own units.
- [`icon`](../component/icon): state marks (✓ ✗ ● ○) in Unicode or ASCII.
- [`toast`](../component/toast): notices that slide into a corner, stack and leave on their own.
- [`inline`](../component/inline): a yes/no question, any form field asked on its own, and a progress bar for CLIs that print logs.
- [`logger`](../component/logger): a `slog.Handler` that hangs every attribute under its message as a tree and wraps long values without cutting them, and `Pane`, a log panel that follows the end as lines arrive and filters them with `/`.
- [`themebrowser`](../component/themebrowser): the preset table behind a `--list-themes` flag.

**Effects**, in [`effect`](../effect): what plays over a block already drawn.

- [`effect`](../effect): effects and lights over any drawn block (fades, dissolves, slides, a melt, a tint, a spotlight, a hover), kept to rows or a part of it, transitions from one block to the next, and intent presets.

**Animation**, in [`animation`](../animation): what moves, and the one clock that drives it.

- [`animation`](../animation): tweens on a dozen easing curves that delay, repeat and yoyo, springs, points, physics (a drop, a pendulum, an orbit, a flock, a rubber band) and animated chart values, with one clock to drive them; `BEZEL_NO_MOTION` turns motion off and `BEZEL_MOTION` sets its speed.
- [`animation/text`](../animation/text): text that types, scrolls, shimmers, scrambles, rolls, waves, decrypts, blinks and rises.
- [`animation/scene`](../animation/scene): rain, snow, confetti, fireworks, stars, static and plasma, alone or behind a block.
