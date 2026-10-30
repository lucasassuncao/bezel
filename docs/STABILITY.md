# Stability

What bezel promises from v1 on: which packages hold still, which may move, and what kind of change needs a new major version. An app that upgrades within v1 reads this page to know what can break.

## What is stable

| Packages | Promise |
|---|---|
| the root `bezel`, `bezeltest` | stable |
| the foundation: `layout`, `key`, `draw`, `legend`, `theme` | stable |
| every package under `component/` | stable |
| the frame: `overlay`, `shell` | stable |
| `animation`, `animation/text`, `animation/scene`, `effect` | **experimental** |

The experimental packages are the motion: tweens, springs, text that moves, scenes and effects. They are where bezel still learns what it wants, so they may change in a minor release, and the release notes say so when they do. A stable package may use them inside, as `progress` moves on a spring, but none of their types appears in a stable package's API. If one ever must, that part of the API is experimental too.

## Changes a minor release may make

- a new package, function, method, `WithX`, `SetX`, preset or event: a type switch on `Result.Event` passes over an event it does not know;
- a new field in a `Config`, `Styles` or `Keys` struct, as long as its zero value keeps the behaviour from before;
- a new role interface, the way `overlay.Resizer` came: behaviour an app opts into by implementing it.

## Changes only a new major version (`/v2`) makes

- removing or renaming anything exported, or changing a signature;
- adding a method to an interface an app implements (`shell.Tab`, the overlay roles, `form.Holder`): it breaks every implementation, so new behaviour comes as a new role interface instead;
- showing an `animation` or `effect` type in a stable package's API.

## What is not promised

- **What `View` draws, cell for cell**: spacing, glyphs, the colours a preset derives, `DefaultStyles` and `DefaultKeys`. A test that compares a whole frame may need its expected text again after an upgrade; a test that looks for what a frame says, as the examples' tests do, rarely does.
- **The experimental packages**, as above.
- **The bubbletea and lipgloss types** in bezel's signatures (`tea.Msg`, `tea.View`, `lipgloss.Style`, ...): bezel follows their v2, and a break there is a break here.

## How the promise is kept

- [CONVENTIONS.md](CONVENTIONS.md) is how a package is written, so a new one fits without a breaking change later: the theme's place, the one `Update`, `With` for options and `Set` for live changes.
- The guides' Go snippets are type-checked against the API on every `make test` (`scripts/doccheck`), so the docs cannot describe an API bezel no longer has.
- Every package that draws is fuzzed to keep to the size it is given, and every frame of motion is benchmarked against a 2ms budget on a 120x40 screen.
