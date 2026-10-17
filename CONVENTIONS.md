# Conventions

The rules every bezel package follows, so a new one reads like the rest and an app never has to guess which shape a package chose. A change that breaks one of them is a bug in the change, not a new convention.

## Values, not pointers

Every model is a value. A method that changes it returns the changed copy, and the receiver is left as it was:

```go
p = p.WithCursor(2)            // p before the call still has its old cursor
sh, handled, cmd := sh.Update(msg)
```

The exception is `textbox`, which hands back the bubbles `textarea.Model` itself and follows bubbles' own shape, `SetText(&ta, s)` included.

What the shell is handed it keeps by value too. `SetInput` takes the status row's input and `Input` reads the edits back after `Update`; a `Tab` that is a value goes back through `SetTabs` when the app's state changes, instead of being a pointer the model and the shell both hold.

## The theme

A component that draws takes a `theme.Resolved`, never loose lipgloss styles. Where it goes depends on who calls `View`:

| Who calls `View` | Where the theme goes | Packages |
| --- | --- | --- |
| bezel, through an interface with a fixed `View` (`overlay.Overlay`, `form.Field`) | the constructor, as its last parameter | `overlay`, `form`, `palette`, `textbox` |
| the app | `View(th, ...)`, so one model can be drawn under any theme | `list`, `tree`, `browser` |
| nobody, it is a plain function | after the data it draws | `progress`, `inline`, `icon.Set.Render` |
| `shell` | `shell.Config.Theme`, replaced with `SetTheme` | `shell` |

A dialog that reports something is tinted with `WithKind(overlay.Danger)` and the like, which colours its border from the same theme. `Alert` takes its `Kind` in the constructor, because an alert always reports one.

Two layers stay below the theme and take the lipgloss style they paint with: `draw` (text fitting, panels, buttons) and `legend`, plus `overlay.Box` for a host that draws its own modal. `theme` builds on them, so they cannot import it.

Everything about how an input looks travels with the theme too, the caret's blink included: `theme.Resolved.CaretBlink`, zero for `draw.DefaultBlink`. A test suite shrinks it on the theme it builds; there is no package variable to set.

## Three shapes of Update

| Shape | Who | Why |
| --- | --- | --- |
| `Update(tea.Msg) (T, bool, tea.Cmd)` | `shell.Shell`, `overlay.Stack` | They route messages. `handled` false means the message is still the app's. |
| `Update(tea.KeyPressMsg) (T, Action)` | `list`, `tree` | They only move a cursor. The `Action` says what a key did (`list.Chosen`, `tree.Expanded`) and the app decides what that means; they never produce a command. |
| `Update(tea.Msg) (I, tea.Cmd)` | `overlay.Overlay`, `form.Field`, `palette`, and `form.Form` and `animation.Tween`, which return themselves | The leaves bubbletea drives. An interface `I` fixes the return type, so the stack and the form can hold any of them. |

A new component picks the row whose "Why" fits it, not a fourth shape.

## Keys

A component that answers to keys keeps them as `key.Binding`s in a struct: `Keys` where the package has one component (`list`, `tree`, `browser`, `palette`, `form`), `XKeys` where it has several (`overlay.ConfirmKeys`, `overlay.PickKeys`, ...). `DefaultKeys()` or `DefaultXKeys()` builds them and `WithKeys` replaces them. `form.Keys` is shared by the form and its controls, each reading the bindings that concern it, and `Form.WithKeys` rebinds them all.

`Update` matches with `key.Matches`, never a key string, and `Legend` is built from the same bindings with `legend.FromBindings`, so what a component lists is what it answers to. A pair that reads as one entry gives its help to the first binding (`Up` as "↑/↓") and none to the second, which `FromBindings` leaves out. Esc is the stack's, which closes the top dialog before the dialog sees the key: a dialog lists it, and a binding for it only reaches a host that drives the dialog without a stack.

## Constructor, With and Set

- **`NewX(required..., th)`**: what the component cannot work without, positionally, the theme last. A callback may be documented as accepting nil.
- **`WithX(...) X`**: everything optional, set before the component is used. Each returns a copy and chains with the others.
- **`SetX(...) X`**: changes a live component in answer to something that happened: `SetItems` after a reload, `SetTheme` after `tea.BackgroundColorMsg`, `SetTabs` after the tabs' state changed, `SetStatus`.
- **`Config` struct**: only where an app sets many fields at once and most of them every time, as `shell.Config`.
- **Getters** are the bare noun: `Cursor()`, `Value()`, `Selected()`, never `GetCursor()`.

If a constructor would take an argument most callers pass as `nil` or `""`, that argument is a `With`.

## No hidden state

bezel keeps no package variable an app is meant to change, and reads the environment in one place: `icon.FromEnv()`, which picks ASCII marks for `BEZEL_ASCII` or `TERM=dumb`. The form controls that draw marks call it for their default and take `WithIcons(set)` to override it. Anything else that varies per app arrives through a constructor, a `With`, the theme or `shell.Config`.

`View` never writes back. An overlay that needs the body's size in `Update`, such as how far a page scrolls, implements `overlay.Resizer`: the stack hands it the body when pushed and on every new layout, so nothing is measured in `View` and carried over through a pointer the copies share. Whatever scrolls keeps its cursor and offset in a `layout.Window`: the list and its filter, the tree, a `Pick`, the file picker's drives.

Time reaches a model the way keys do, as a message. `animation.New` reads no clock: the first `FrameMsg` its `Tick` delivers starts it, and each tween takes only its own frames, so one replaced mid-flight cannot run a second clock.

The theme presets (`theme.Mint`, `theme.Plain`, ...) are exported values to build on: point a `Theme`'s `Base` at one, or copy it, and never change the preset itself.

## Patterns by name

The shapes above are patterns that recur across the packages. A new component reaches for one of these before it invents a shape.

| Pattern | Where | What it buys |
| --- | --- | --- |
| Model-View-Update | every model | `Update` returns the next value and a command, `View` only reads, and effects run as commands that answer with messages. |
| Composite | `layout.Node`: a `Box` of boxes and `Leaf`s | One tree resolves to rects, prunes absent leaves and collapses as a whole. |
| Chain of responsibility | `Shell.Update`: the overlays, then the status input, then the actions | Each link takes what is its own; `handled` false hands the rest to the app. |
| Command | `shell.Action`, `palette.Item` | A key, a name and what it runs, held as data the legend, the palette and the dispatch all read. |
| Functional options | `shell.Option` on the prebuilt actions; `With` on every model | Optional settings without a constructor that grows a parameter for each. |
| Role interfaces | `overlay.EscOwner`, `Resizer`, `StatusLiner`, `BottomAnchored`, `form.Holder` | An overlay or field opts into a behaviour by implementing one small interface, as `io.WriterTo` works. |
| Strategy | `tree.View`'s row, `list.WithMatch`, `Prompt.WithValidate`, `Config.Can`, `When` | A function where the app decides and bezel only applies the answer. |
| Adapter | `overlay.FilePicker`, `textbox`, `form.Input` | A bubbles component in bezel's shape: its theme, its keys, its legend. |
| Cascade | `theme.Theme.Base` | Each layer overrides the one under it, so a preset is a base to build on. |
