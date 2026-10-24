// Package bezel is the frame around a bubbletea app: layout, panels, theme,
// overlays and a shell that owns resize, tabs, focus, status and the legend.
// The app only draws what goes inside the panes.
//
// This package holds no code. The work is in its subpackages.
//
// # The frame
//
// What every screen is built in, at the module's root:
//
//   - [github.com/lucasassuncao/bezel/shell]: the struct an app embeds. Owns
//     resize, tabs, focus, status, the overlay stack, the legend and the key
//     actions.
//   - [github.com/lucasassuncao/bezel/layout]: declares the screen as a tree
//     of named leaves and resolves it to rects.
//   - [github.com/lucasassuncao/bezel/draw]: ANSI-safe text fitting, panels,
//     header, tabs, status line and overlay compositing.
//   - [github.com/lucasassuncao/bezel/theme]: the semantic palette, presets
//     and the lipgloss styles everything draws with.
//   - [github.com/lucasassuncao/bezel/overlay]: the stack of modals.
//     Ready-made Alert, Confirm, Text, Choice, Pick, Help, Prompt, Pager and
//     FilePicker.
//   - [github.com/lucasassuncao/bezel/legend]: key entries filtered by
//     capability and packed into a few rows.
//   - [github.com/lucasassuncao/bezel/key]: key bindings, with the help the
//     legend prints and a switch to turn them off.
//   - [github.com/lucasassuncao/bezel/palette]: the ":" command line, with
//     prefix matching, completion and arguments.
//   - [github.com/lucasassuncao/bezel/bezeltest]: builds the key presses a
//     terminal sends, for tests that drive a model.
//
// # Components
//
// What goes inside the panes, under component:
//
//   - [github.com/lucasassuncao/bezel/component/list]: a cursor list with
//     section headings and a / filter.
//   - [github.com/lucasassuncao/bezel/component/paginator]: pages over a run of
//     items, drawn as dots or as "2/5".
//   - [github.com/lucasassuncao/bezel/component/tree]: a flat depth-first tree
//     with expand/collapse, generic over the node type.
//   - [github.com/lucasassuncao/bezel/component/table]: sizes columns to their
//     contents within a width, and a cursor over rows under a fixed header.
//   - [github.com/lucasassuncao/bezel/component/viewport]: scrollable text,
//     wrapped or panned sideways, with a gutter and search highlights.
//   - [github.com/lucasassuncao/bezel/component/browser]: a pick-from-a-list
//     screen, labels left and the selected item's detail right.
//   - [github.com/lucasassuncao/bezel/component/form]: Input, Select, Checkbox,
//     Radio, Toggle and Button, stacked in a Form that moves focus with tab.
//   - [github.com/lucasassuncao/bezel/component/input]: a one-line text field
//     with readline keys, a placeholder and a mask for secrets.
//   - [github.com/lucasassuncao/bezel/component/textbox]: a multi-line text
//     field with numbered lines that wrap at its width.
//   - [github.com/lucasassuncao/bezel/component/cursor]: a text field's caret
//     on its own, blinking, steady or hidden.
//   - [github.com/lucasassuncao/bezel/component/progress]: Bar, Line and
//     Stepper, how far along something is in one row.
//   - [github.com/lucasassuncao/bezel/component/spinner]: twelve spinners or
//     frames of your own, each ticking on its own.
//   - [github.com/lucasassuncao/bezel/component/timer]: a countdown that says
//     when time is up.
//   - [github.com/lucasassuncao/bezel/component/stopwatch]: time counted up,
//     paused and reset.
//   - [github.com/lucasassuncao/bezel/component/chart]: area, spark, bar and
//     heat charts of a run of numbers.
//   - [github.com/lucasassuncao/bezel/component/icon]: one set of state marks,
//     Unicode or ASCII, in the theme's colours.
//   - [github.com/lucasassuncao/bezel/component/inline]: yes/no questions and
//     progress bars for CLIs that print logs, without taking the screen.
//   - [github.com/lucasassuncao/bezel/component/themebrowser]: an inline table
//     of every preset and its palette, for a --list-themes flag.
//
// # Effects
//
//   - [github.com/lucasassuncao/bezel/effect]: effects, lights and
//     transitions over any drawn block.
//
// # Animation
//
//   - [github.com/lucasassuncao/bezel/animation]: tweens, springs, fades, points
//     and physics on easing curves, and one clock to drive them.
//   - [github.com/lucasassuncao/bezel/animation/text]: text that types, scrolls,
//     shimmers, rolls and waves.
//   - [github.com/lucasassuncao/bezel/animation/scene]: rain, snow, stars and
//     fireworks, alone or behind a block.
//
// docs/ARCHITECTURE.md is why the folders are laid out as they are, and
// docs/CONVENTIONS.md how the packages share one shape: where the theme goes,
// the three forms of Update, constructor or With.
package bezel
