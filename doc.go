// Package bezel is the frame around a bubbletea app: layout, panels, theme,
// overlays and a shell that owns resize, tabs, focus, status and the legend.
// The app only draws what goes inside the panes.
//
// This package holds no code. The work is in its subpackages:
//
//   - [github.com/lucasassuncao/bezel/shell]: the struct an app embeds. Owns resize,
//     tabs, focus, status, the overlay stack, the legend and the key actions.
//   - [github.com/lucasassuncao/bezel/layout]: declares the screen as a tree of
//     named leaves and resolves it to rects.
//   - [github.com/lucasassuncao/bezel/draw]: ANSI-safe text fitting, panels,
//     header, tabs, status line and overlay compositing.
//   - [github.com/lucasassuncao/bezel/theme]: the semantic palette, presets and
//     the lipgloss styles everything draws with.
//   - [github.com/lucasassuncao/bezel/themebrowser]: an inline table of every
//     preset and its palette, for a --list-themes flag.
//   - [github.com/lucasassuncao/bezel/overlay]: the stack of modals. Ready-made
//     Alert, Confirm, Text, Choice, Pick, Help, Prompt, Pager and FilePicker.
//   - [github.com/lucasassuncao/bezel/form]: Input, Select, Checkbox, Radio,
//     Toggle and Button, stacked in a Form that moves focus with tab.
//   - [github.com/lucasassuncao/bezel/input]: a one-line text field with
//     readline keys, a placeholder and a mask for secrets.
//   - [github.com/lucasassuncao/bezel/progress]: Bar, Line and Stepper, how far
//     along something is in one row.
//   - [github.com/lucasassuncao/bezel/icon]: one set of state marks, Unicode or
//     ASCII, in the theme's colours.
//   - [github.com/lucasassuncao/bezel/legend]: key entries filtered by capability
//     and packed into a few rows.
//   - [github.com/lucasassuncao/bezel/key]: key bindings, with the help the
//     legend prints and a switch to turn them off.
//   - [github.com/lucasassuncao/bezel/palette]: the ":" command line, with prefix
//     matching, completion and arguments.
//   - [github.com/lucasassuncao/bezel/list]: a cursor list with section headings
//     and a / filter.
//   - [github.com/lucasassuncao/bezel/tree]: a flat depth-first tree with
//     expand/collapse, generic over the node type.
//   - [github.com/lucasassuncao/bezel/table]: sizes columns to their contents
//     within a width, and a cursor over rows under a fixed header.
//   - [github.com/lucasassuncao/bezel/textbox]: the bubbles text area set up for
//     a multi-line value.
//   - [github.com/lucasassuncao/bezel/browser]: a pick-from-a-list screen, labels
//     left and the selected item's detail right.
//   - [github.com/lucasassuncao/bezel/animation]: an eased integer tween for
//     panels that slide open.
//   - [github.com/lucasassuncao/bezel/inline]: yes/no questions and progress bars
//     for CLIs that print logs, without taking the screen.
//   - [github.com/lucasassuncao/bezel/bezeltest]: builds the key presses a
//     terminal sends, for tests that drive a model.
//
// CONVENTIONS.md, at the module's root, is how the packages share one shape:
// where the theme goes, the three forms of Update, constructor or With.
package bezel
