# Architecture

How the repository is laid out, what each folder holds, and why it sits where it does. The rules every package follows inside its own code are in [CONVENTIONS.md](CONVENTIONS.md); this page is about where the code lives and what may import what.

```
bezel/
├── shell, overlay              the frame: what an app embeds and what it opens over the panes
├── theme, draw, legend,        the foundation: colours, drawing, key hints, bindings and rects,
│   key, layout                 under everything else
├── bezeltest                   key presses for tests
├── component/                  what goes inside the panes
├── effect/                     changes to a block that is already drawn
├── animation/                  values that move, and text/ and scene/
├── internal/                   plumbing shared by the folders above, private to the module
├── examples/                   one program per package, by group, and a first app
├── docs/                       the guides, every recording under recordings/ and the site under pages/
└── scripts/                    the Makefile's includes and the scripts it runs
```

## The root: foundation and frame

The packages at the root are the ones every app touches: `shell` to embed, `theme` to resolve, `layout` to declare the screen. They stay at the root so their imports are the shortest, `bezel/shell` rather than a longer path, and so the module's root reads as its main API, as Go modules usually do. They are two layers, not one:

- **The foundation**: `layout`, `key`, `draw`, `legend` and `theme`. They import nothing from the rest of bezel but each other, and everything else is built on them. `layout` and `key` stand alone; `draw` sits on `layout`, `legend` on `draw` and `key`, and `theme` on `draw` and `legend`, whose styles it fills in.
- **The frame**: `overlay` and `shell`. They are the top of the stack: they compose components (the prompt's `input`, the pager's `viewport`, the busy `spinner`), and nothing in the module imports them except one another and the examples. `legend`, `overlay` and the `:` palette are not components, though they draw: the shell owns them, builds them from its `Action`s and draws them around the panes, and an app never places one in a pane of its own.

`bezeltest` is apart from both: it builds the key presses a terminal sends, for tests, and imports nothing from bezel.

## component/

What an app creates and draws inside a pane: `list`, `paginator`, `tree`, `table`, `viewport`, `browser`, `filepicker`, `accordion`, `form`, `input`, `textbox`, `cursor`, `progress`, `spinner`, `timer`, `stopwatch`, `chart`, `icon` and `toast`, and the three that print into a terminal instead of taking it over, `inline`, `logger` and `themebrowser`.

A component imports the foundation, the motion packages and, where it is made of others, other components: `form` holds `input`s, `inline` draws a `progress` bar, `themebrowser` is a `table`, `progress` moves on an `animation.Spring`. A motion type stays inside: no stable component shows one in its API, since `animation` and `effect` are experimental. A component never imports the frame, so any of them works in a pane, in a dialog or in a program without a shell. That is also why the shell cannot click a list for the app: it finds the pane under the pointer and sends a `ClickMsg`, and the app, which holds the list, hands it on. `overlay.FilePicker` is the one dialog built on a component, `filepicker`. `component/` is a folder, not a package: each component is its own package, `component/list` imported as `list`.

## effect/

Changes to a block that is already drawn: fades, dissolves, slides, lights, the transitions from one block to the next, and the presets that name an intent. An effect reads the block's escape codes back into cells and draws them again, so it works over anything a `View` returns.

It is a folder of its own, not part of `animation`, because it is a different kind of thing: an animation moves a value the app draws, an effect changes what was drawn. It runs on the animation clock, so `effect` imports `animation`, never the other way.

## animation/

Values that move on frames that arrive as messages: tweens, springs, fades, points and physics, the easing curves, and the `Clock` that drives several at once. The core imports nothing from bezel but `internal/frames`, so an app can animate its own drawing without anything else.

Two subpackages animate something more specific, and live under it because they are animations themselves:

- `animation/text`: a line of text that types, scrolls, shimmers, rolls or waves.
- `animation/scene`: whole scenes drawn from the time alone, rain or stars, alone or behind a block.

Every animation, effect and scene takes the same `animation.FrameMsg`, so one screen mixes them freely.

## internal/

Plumbing more than one folder needs and no app should depend on. Go lets only the packages under `internal/`'s parent import it, so it sits at the root to reach every folder:

| Package | What it holds | Who uses it |
|---|---|---|
| `frames` | the frame message, the run a timed motion keeps, the motion switches read from the environment | `animation`, `effect`, and the components that move: `list`, `accordion`, `toast` |
| `cells` | a drawn block read back into styled cells, and the canvas they are put into | `effect`, `animation/scene` |
| `noise` | the hash randomness is drawn from, the same for the same frame | `effect`, `animation/text`, `animation/scene` |
| `hue` | colours turned round the HSL wheel | `effect`, `animation/text` |
| `edit` | a line of graphemes with a caret, the editing `input` and `textbox` share, and the caret's blink | `input`, `textbox`, `cursor`, `viewport`, `animation/text` |
| `clip` | the system clipboard, read as the paste a terminal sends | `input`, `textbox` |

A public type can still come from here: `animation.FrameMsg` and `animation.Ease` are aliases of `frames` types, so every package names one type while the run behind it stays private.

## What may import what

From the bottom up, each layer imports only the ones below it:

0. the root `bezel`: `Result`, `Model` and `ClickMsg`, importing only bubbletea
1. `internal/`
2. motion: `animation/` (with `text` and `scene`), then `effect/` on it; they import nothing from bezel but `internal/` and the root
3. the foundation: `layout`, `key`, `draw`, `legend`, `theme`
4. `component/`, which may move with `animation` and `effect` inside
5. the frame: `overlay`, `shell` (with its `:` palette in `shell/internal/palette`)

A change that needs an arrow pointing up, an animation that imports `theme` or a component that imports `overlay`, is a sign the code belongs a layer higher.

## examples/, docs/ and scripts/

- **`examples/`**: one small program per package, in the same groups as the code: `frame/` for the packages at the root, `component/` for the components, and a gallery for each package that moves, a tab per family: `animation/` for the values, `animation/text/` and `animation/scene/` beside it, and `effect/` for the effects. `app/` is a small app built from several of them, which [EXAMPLE.md](EXAMPLE.md) explains step by step. Only code lives here.
- **`docs/recordings/`**: what the examples look like running. Each GIF sits beside the `.tape` that records it, named after its example and grouped as `examples/` is: `component/list.tape` records `examples/component/list`. [Recording](COMPONENTS.md#recording) is how the tapes are run.
- **`docs/`**: the guides. [FRAME.md](FRAME.md) is how the frame fits together, [THEME.md](THEME.md) the colours, [COMPONENTS.md](COMPONENTS.md) walks every package's example, [EFFECTS.md](EFFECTS.md) and [ANIMATIONS.md](ANIMATIONS.md) their packages, [CONVENTIONS.md](CONVENTIONS.md) the rules inside a package, and this page the layout. Each package's own `README.md`, next to its code, is generated from its doc comments by `make docs`. `pages/` holds the website, a landing page and the guides served by docsify: `make stage-docs` assembles it with the guides and recordings into `site/`, which CI publishes to GitHub Pages, and `make serve-docs` serves it locally.
- **`scripts/`**: `make/` holds the Makefile's includes, one per concern (tools, docs, tests, security, release), and `bash/` and `pwsh/` the scripts they call, the same script twice so `make` works on Windows and on Unix alike. `doccheck/` is a Go test that type-checks every Go snippet in the README and the guides against the API, so a guide that names something bezel no longer has fails `make test`.
