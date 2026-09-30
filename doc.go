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
//     preset, for a --list-themes flag.
//   - [github.com/lucasassuncao/bezel/overlay]: the stack of modals. Ready-made
//     Alert, Confirm, Help and Prompt.
//   - [github.com/lucasassuncao/bezel/legend]: key entries filtered by capability
//     and packed into a few rows.
//   - [github.com/lucasassuncao/bezel/palette]: the ":" command line, with prefix
//     matching, completion and arguments.
//   - [github.com/lucasassuncao/bezel/list]: a cursor list with section headings
//     and a / filter.
//   - [github.com/lucasassuncao/bezel/tree]: a flat depth-first tree with
//     expand/collapse, generic over the node type.
//   - [github.com/lucasassuncao/bezel/table]: sizes columns to their contents
//     within a width.
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
package bezel
