# Conventions

The rules every bezel package follows, so a new one reads like the rest. A change that breaks one is a bug in the change, not a new convention.

## Values, not pointers

Every model is a value: a method that changes it returns the changed copy and leaves the receiver as it was.

```go
p = p.WithCursor(2)            // p before the call still has its old cursor
sh, handled, cmd := sh.Update(msg)
```

The shell keeps what it is handed by value too: `SetInput` takes the status row's input and `Input` reads it back after `Update`; a `Tab` that changes goes back through `SetTabs`.

## The theme

A component that draws takes a `theme.Resolved`; loose lipgloss styles reach it only through `WithStyles`, below. Where the theme goes depends on who calls `View`:

| Who calls `View` | Where the theme goes | Packages |
| --- | --- | --- |
| bezel, through a fixed interface (`overlay.Overlay`, `slog.Handler`), or a field that needs the theme before it draws (a caret's blink pace) | the constructor's last parameter | `overlay`, `palette`, `textbox`, `input`, `cursor`, `logger` |
| bezel, for the controls of a form | `form.New(th, fields...)`, first because the fields are variadic; the form hands it to every control and `SetTheme` replaces it | `form` |
| the app | `View(th, ...)`, so one model draws under any theme | `list`, `tree`, `browser`, `table.Model`, `spinner`, `paginator`, `timer`, `stopwatch`, `viewport`, `accordion`, `toast` |
| nobody: it is a plain function | after the data it draws | `progress`, `chart`, `inline`, `icon.Set.Render` |
| `shell` | `shell.Config.Theme`, replaced with `SetTheme` | `shell` |

- A dialog that reports something is tinted with `WithKind(overlay.Danger)` and the like. `Alert` takes its `Kind` in the constructor, because it always reports one. The `Kind` colours stay the theme's success, warning and danger: a `WithStyles` edit never changes what a dialog reports.
- `draw`, `legend` and `overlay.Box` sit below the theme and take the lipgloss style they paint with: `theme` imports them, so they cannot import it.
- How an input looks travels with the theme too, the caret's blink included: `Resolved.CaretBlink`, zero for `draw.DefaultBlink`. Tests shrink it there; there is no package variable to set.

## Styles

The theme sets how everything looks; a component's own styles let an app recolour that one component without touching the theme.

- `Styles` holds one lipgloss style per part the component paints; `XStyles` where a package has several components.
- `DefaultStyles(th)` derives them from the theme, and is the only place that does.
- `WithStyles(edit)` stores `edit func(Styles) Styles` and applies it to `DefaultStyles(th)` whenever the component has a theme: when built, on `SetTheme`, or in `View(th, ...)`. The app's change outlives a theme switch. A second `WithStyles` replaces the first edit.
- A part that no role in `theme.Resolved` fits gets a new role there, derived in `buildDerivedStyles`, never a colour written into the component.

Not everything can follow the theme. Each package sits in one of three classes, and one that is not themable says so in its doc comment:

| Class | What the app can do | Packages |
| --- | --- | --- |
| Themable | switch the theme, and recolour this one component with `WithStyles` | `list`, `tree`, `table`, `browser`, `accordion`, `input`, `textbox`, `form`, `palette`, `cursor`, `filepicker`, `logger`, `paginator`, `progress.Model`, `spinner`, `stopwatch`, `timer`, `viewport`, and every `overlay` dialog but `Alert` (one `overlay.Styles` for all of them) |
| Theme only | switch the theme; a plain function keeps no edit, so it has no `Styles` | `chart`, `progress.Bar` and its siblings, `inline`, `icon.Set.Render` |
| Non-themable | nothing: the colours report something, or arrive as arguments | `overlay.Alert` and the `Kind` tint of any dialog, `toast`, `animation` and its subpackages, `effect`, `themebrowser` |

## Three shapes of Update

| Shape | Who | Why |
| --- | --- | --- |
| `Update(tea.Msg) (T, bool, tea.Cmd)` | `shell.Shell`, `overlay.Stack` | They route: `handled` false means the message is still the app's. |
| `Update(tea.KeyPressMsg) (T, Action)` | `list`, `tree`, `browser`, `table.Model`, `paginator` | They only move a cursor. The `Action` says what a key did (`list.Chosen`, `tree.Expanded`), the app decides what it means, and they never produce a command. A list with motion keeps that shape: `Tick` and `Animate` carry its frames apart from `Update`. |
| `Update(tea.Msg) (I, tea.Cmd)` | `overlay.Overlay`, `form.Field`, `palette`; `form.Form`, `input.Model`, `textbox.Model`, `cursor.Model`, `spinner.Model`, `timer.Model`, `stopwatch.Model`, `viewport.Model`, `accordion.Model`, `toast.Model` and the types of `animation`, its subpackages and `effect` return themselves | The leaves bubbletea drives. The interface `I` lets a stack or a form hold any of them. |

A new component picks the row whose "Why" fits it, not a fourth shape.

## Keys

- Bindings live in a struct: `Keys` where a package has one component (`list`, `tree`, `browser`, `palette`, `form`, `paginator`, `viewport`), `XKeys` where it has several (`overlay.ConfirmKeys`, `overlay.PickKeys`, ...). `DefaultKeys()` or `DefaultXKeys()` builds them and `WithKeys` replaces them. `form.Keys` serves the form and every control in it, and `Form.WithKeys` rebinds them all.
- `Update` matches with `key.Matches`, never a key string, and `Legend` comes from the same bindings through `legend.FromBindings`, so a component lists exactly what it answers to.
- Every component with a `Keys` struct has a `Legend()`, which follows its state: a list's esc only while filtering, a form's for the control in focus. A text field lists only the keys worth a hint (paste, complete, copy); its editing keys are the terminal's usual ones and go unlisted.
- A pair that reads as one entry gives its help to the first binding (`Up` as "↑/↓") and none to the second, which `FromBindings` then leaves out.
- Esc belongs to the stack, which closes the top dialog before the dialog sees the key. A dialog still lists it; its own binding for it only reaches a host that drives the dialog without a stack.

## Clicks

The mouse reaches a component the way a key does: the app hands it on, already moved into the pane's cells by the shell's `ClickMsg`, `Mouse` or `MouseIn`.

- A component that moves a cursor takes a click through **`Click(y int) (T, Action)`** beside its `Update`, with the same return: `list`, `table.Model`, `tree`. `y` is a line of what its `View` drew. The first click moves the cursor to the row; a click on the row already under it does what enter does there.
- `form.Form` takes **`Click(y, width int) (Form, tea.Cmd)`**: the width is the one its `View` gets, which sets how tall an open select draws.
- A leaf that tracks a press, a drag and a release, as `textbox` does, takes the mouse messages in its own `Update`, like its keys.
- A dialog takes clicks by implementing `overlay.Clicker`, counted from the top left of what it drew.

`Click` is not a fourth shape of `Update`: it is a second way in, and answers as `Update` does.

## Constructor, With and Set

- **`NewX(required..., th)`**: what the component cannot work without, the theme last. A callback may be documented as accepting nil. A constructor returns only the value: a command to start it comes from `Init()`, as `form.Form.Init` does.
- **Sizes are not required**: the rows and cells a pane has arrive after the first `tea.WindowSizeMsg`, so they come through `SetHeight` and `SetWidth` (`list`, `tree`, `table.Model`, `viewport`, `filepicker`), never as a constructor parameter.
- **`WithX(...) X`**: everything optional, set before the component is used. Each returns a copy and chains with the others.
- **`SetX(...) X`**: changes a live component after something happened: `SetItems` after a reload, `SetTheme` after `tea.BackgroundColorMsg`, `SetTabs`, `SetStatus`.
- **`Config` struct**: only where an app sets many fields at once, most of them every time, as `shell.Config`.
- **Getters** are the bare noun: `Cursor()`, `Value()`, `Selected()`, never `GetCursor()`.

An argument most callers would pass as `nil`, `""`, `0` or `false` is a `With`, not a constructor parameter: `form.NewCheckbox(label).WithChecked(on)`, `browser.New(items).WithCurrent(label)`. Two parameters of one type in a row that read the same way are a `With` each, so they cannot be swapped unseen: `text.NewTicker(v, dur).WithUp(c).WithDown(c)`.

## No hidden state

- No package variable is meant to be changed by an app. The environment is read in two places, both when called: `icon.FromEnv()` (ASCII for `BEZEL_ASCII` or `TERM=dumb`), which the controls that draw marks default to and `WithIcons(set)` overrides, and `animation.Still()` with `animation.Speed()` (no motion for `BEZEL_NO_MOTION`, `TERM=dumb` or `BEZEL_MOTION=0`, and the speed of the rest from `BEZEL_MOTION`), which every animation consults when built and as it plays. Everything else that varies per app arrives through a constructor, a `With`, the theme or `shell.Config`.
- `View` never writes back. An overlay that needs its size in `Update` implements `overlay.Resizer`, and the stack hands it the body when pushed and on every new layout. Whatever scrolls keeps its cursor and offset in a `layout.Window`, and hands a pane's scrollbar its position through `Scroll()`.
- Time arrives as a message, like keys. No animation reads the clock: the first `FrameMsg` its `Tick` delivers starts it, each one takes only its own frames, and a `Clock` hands one frame's time to several at once.
- The presets (`theme.Mint`, `theme.Plain`, ...) are values to build on: point a `Theme`'s `Base` at one, or copy it, and never change the preset itself.

## Patterns by name

A new component reaches for one of these before it invents a shape.

| Pattern | Where, and what it buys |
| --- | --- |
| Model-View-Update | Every model: `Update` returns the next value and a command, `View` only reads, effects run as commands. |
| Composite | `layout.Node`, boxes of boxes and leaves: one tree resolves, prunes absent leaves and collapses as a whole. |
| Chain of responsibility | `Shell.Update`: the overlays, then the status input, then the actions; `handled` false hands the rest to the app. |
| Command | `shell.Action`, `palette.Item`: a key, a name and what it runs, held as data the legend, the palette and the dispatch read. |
| Functional options | `shell.Option` on the prebuilt actions, `With` on every model and on `key.Binding`: optional settings without a growing constructor. |
| Role interfaces | `overlay.EscOwner`, `Resizer`, `StatusLiner`, `BottomAnchored`, `Clicker`, `form.Holder`, `form.Initer`: an overlay or field opts into a behaviour by implementing one small interface. |
| Strategy | `tree.View`'s row, `list.WithMatch`, `Prompt.WithValidate`, `Config.Can`, `When`: the app decides, bezel applies the answer. |
| Cascade | `theme.Theme.Base`: each layer overrides the one under it, so a preset is a base to build on. |
