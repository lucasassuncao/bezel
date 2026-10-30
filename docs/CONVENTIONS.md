# Conventions

The rules every bezel package follows, so a new one reads like the rest. Each rule is a snippet first, then why. A change that breaks one is a bug in the change, not a new convention.

## Models are values

```go
next := l.SetHeight(10)   // l is untouched; next has the new height
m.list = m.list.SetHeight(10)
```

Every method that changes a model returns the changed copy. Keep the result, or the change is lost. The shell works the same way: a `Tab` that changes goes back through `SetTabs`, and the status row's input goes in with `SetInput` and comes back out with `Input`.

## Building a component

```go
l := list.New(rows).        // NewX(required...)
	WithWrap(cfg.Wrap).       // WithX: optional, set before use, chains
	WithKeys(list.VimKeys())

m.list = m.list.SetRows(reloaded) // SetX: a live change, after something happened
```

- **`NewX(required..., th)`** takes only what the component cannot work without, the theme last when it takes one. It returns only the value; a command to start it comes from `Init()`.
- **`WithX`** is everything optional. A switch takes its state, `WithWrap(on bool)`, so an app can pass a setting it read. An argument most callers would pass as `nil`, `""`, `0` or `false` is a `With`, and so is each of two same-typed parameters in a row: `text.NewTicker(v, dur).WithUp(c).WithDown(c)` cannot be swapped by mistake.
- **`SetX`** changes a live model: `SetRows`, `SetTheme`, `SetStatus`. What can change live has only the `Set`, never a `With` doing the same.
- **A variant is named, not built**: a preset of a shared type is a bare name, `spinner.Dot()`, `effect.Wipe(...)`, `theme.Mint()`.
- **`Config` struct** only where an app sets many fields at once, as `shell.Config`.
- **A declared tree is not a model**: `layout`'s builders read as the screen they describe, `layout.Fixed("help").Info()` and `Columns(...).Collapse(72, "detail")`, so they keep their short names instead of `WithX`.
- **A range stays in order**: a from/to or low/high pair is positional, `animation.NewTween(from, to, dur)`, `NewRubber(lo, hi)`, because the order is the meaning; the same-typed rule above is for settings that could be swapped.

### Sizes

```go
case tea.WindowSizeMsg: // the shell laid out the panes and handed the size on
	m.list = m.list.SetHeight(draw.InnerRect(m.sh.Rect("items")).H)
```

A size is never a constructor parameter nor an argument of `View`: it comes through `SetHeight`/`SetWidth` once the terminal says how big it is. A model not sized yet draws nothing. A layout that changes without a resize (`SetLayout`) sizes again after the change.

Sized this way: `list`, `tree`, `table.Model`, `viewport`, `filepicker`, `accordion`, `form`, `textbox`, `progress`; `browser` has `SetListHeight` and `SetPreviewHeight`.

### Getters and errors

```go
row, ok := m.list.Selected() // a copy; false when nothing is under the cursor
if err := in.Err(); err != nil { ... }
if m.acc.IsOpen(i) { ... }
```

- Getters are the bare noun: `Cursor()`, `Value()`, never `GetCursor()`.
- A bool is an adjective or participle: `Running()`, `Focused()`, `Active()`. `Is`/`Has` leads only when the bare word is already an action on the type: `IsOpen` beside `Toggle`, `IsBusy` beside `Busy`, `HasOverlay` beside `Push`.
- Every model with a cursor over items has `Selected() (T, bool)`: `list`, `tree`, `table.Model`, `browser`. A form control's `Selected() int` is an index instead.
- What went wrong is an `error`: checks are `func(string) error` (`input.WithValidate`, `Prompt.WithValidate`), and the getter is `Err()`.

## The theme

```go
overlay.NewConfirm("Delete", item, yes, th) // bezel calls View: theme in the constructor
m.list.View(th)                             // the app calls View: theme in View
progress.Line("tasks", done, total, w, th)  // a plain function: theme after the data
shell.New(shell.Config{Theme: th})          // the shell: in Config, SetTheme to change
```

Where the theme goes depends on who calls `View`. If bezel does, the component keeps the theme from its constructor. If the app does, the app passes it each time, so one model draws under any theme.

| Who calls `View` | Where the theme goes | Packages |
| --- | --- | --- |
| bezel (a fixed interface, or a field that needs the theme before drawing) | constructor, last parameter | `overlay`, `textbox`, `input`, `cursor`, `logger` |
| bezel, for a form's controls | `form.New(th, fields...)`, first because the fields are variadic | `form` |
| the app | `View(th, ...)` | `list`, `tree`, `browser`, `table.Model`, `spinner`, `paginator`, `timer`, `stopwatch`, `viewport`, `accordion` |
| nobody: a plain function | after the data | `progress`, `chart`, `inline`, `icon.Set.Render` |
| the shell | `Config.Theme`; `SetTheme` also repaints open dialogs | `shell` |

- A dialog that reports something is tinted with `WithKind(draw.Danger)`; `Alert` takes its `draw.Kind` in the constructor. A Kind's colour is fixed (blue, green, yellow, red), the same under every theme and whatever `WithStyles` does; a dialog with no Kind keeps the theme's look.
- `draw`, `legend` and `overlay.Box` take raw lipgloss styles: `theme` imports them, so they cannot import it.
- The caret's blink pace travels with the theme (`Resolved.CaretBlink`), so tests shrink it there, not in a package variable.

## Styles

```go
l := list.New(rows).WithStyles(func(s list.Styles) list.Styles {
	s.Cursor = s.Cursor.Bold(true)
	return s
})
```

The theme sets how everything looks; `WithStyles` recolours one component. The edit is applied on top of `DefaultStyles(th)` every time the theme changes, so it survives a light/dark switch. A second `WithStyles` replaces the first.

- `Styles` has one lipgloss style per part painted (`XStyles` when a package has several components).
- `DefaultStyles(th)` is the only place styles are derived from the theme.
- A colour no role in `theme.Resolved` fits gets a new role in `buildDerivedStyles`, never a literal in the component.

Each package is in one class, and a non-themable one says so in its doc comment:

| Class | The app can | Packages |
| --- | --- | --- |
| Themable | switch the theme and use `WithStyles` | `list`, `tree`, `table`, `browser`, `accordion`, `input`, `textbox`, `form`, `cursor`, `filepicker`, `logger`, `paginator`, `progress.Model`, `spinner`, `stopwatch`, `timer`, `viewport`, every `overlay` dialog but `Alert` |
| Theme only | switch the theme (a plain function keeps no edit) | `chart`, `progress.Bar` and siblings, `inline`, `icon.Set.Render` |
| Non-themable | nothing: the colours mean something, or are arguments | `overlay.Alert` and any `Kind` tint, `toast` (its `draw.Kind` colours are fixed), `animation` and subpackages, `effect`, `themebrowser` |

## One Update

```go
var r bezel.Result
if m.sh, r = m.sh.Update(msg); r.Handled { // the frame's first
	return m, r.Cmd
}
m.list, r = m.list.Update(msg)
switch e := r.Event.(type) {
case list.Chosen[string]:
	return m.open(e.Row.Value)
}
return m, r.Cmd
```

Every model answers the same signature, `Update(tea.Msg) (T, bezel.Result)`, and `bezel.Model[T]` names it. The `Result` has three fields: `Handled` says the message was this model's, `Event` says what happened (nil when nothing did), `Cmd` is the effect to run. A model that has nothing to report leaves `Event` nil; a model with nothing to run leaves `Cmd` nil.

`Handled` is true exactly when:

1. the key matches one of the model's bindings, even if nothing moved (`↑` on the first row is the list's);
2. the model is focused and takes typed text or a `tea.PasteMsg`, or an open filter takes text;
3. the message is a frame tagged with the model's id, or the own tick of `timer`, `stopwatch` or a caret;
4. a `bezel.ClickMsg` lands on a row the model drew;
5. in `shell.Shell` and `overlay.Stack`, wherever they took the message. A resize stays false.

Anything else is `Result{}`, and the app keeps the message.

The events are structs in the component's package. Read them with a type switch on `Result.Event`, not with `==`: some hold a func and are not comparable. They are:

| Package | Event |
| --- | --- |
| `list` | `Chosen[T]{Row}` |
| `table` | `Chosen[T]{Item}` |
| `tree` | `Expanded[N]{Node}`, `Collapsed[N]{Node}` |
| `browser` | `Chosen{Item}`, `Dismissed{}` |
| `paginator` | `Paged{Page}` |

An event reports; the app decides what it means. A cursor that starts to move puts its own `Tick()` in `Result.Cmd`, so the app never calls it after `Update`; `Tick()` stays public for after `SetRows`. A new component answers this signature, never a new shape.

## Keys

```go
type Keys struct{ Up, Down, Enter key.Binding }

func DefaultKeys() Keys { /* arrows and named keys only */ }
func VimKeys() Keys     { /* the same, plus j/k */ }

// in Update:
case key.Matches(k, m.keys.Up):
// for the legend, from the same bindings:
legend.FromBindings([]key.Binding{m.keys.Up, m.keys.Enter})
```

- Bindings live in `Keys` (`XKeys` when a package has several components); `WithKeys` replaces them. `form.Keys` serves the form and all its controls.
- `Update` matches with `key.Matches`, never a key string, and `Legend()` is built from the same bindings, so the legend lists exactly what the component answers to. It follows state: a list's esc shows only while filtering.
- No `DefaultKeys` binds a letter. Vim's letters are the opt-in `VimKeys()` preset; a field you type into has none.
- A pair that reads as one entry gives its help to the first binding (`Up` as "↑/↓") and none to the second.
- Esc belongs to the overlay stack, which closes the top dialog before the dialog sees the key.

## Clicks

```go
case shell.ClickMsg: // already moved into the pane's cells
	if msg.Pane == "items" {
		m.list, r = m.list.Update(msg) // the same Update as a key
	}
```

A click enters through `Update`, like a key. `shell.ClickMsg` is `bezel.ClickMsg`: the pane's name, which the app routes by, and the click in the pane's cells. `list`, `table.Model` and `tree` use `Y`: the first click moves the cursor, a click on the current row does what enter does and reports the same event. `form.Form` focuses the field and presses a button, a checkbox or a toggle. A leaf that tracks press, drag and release (`textbox`) takes the click in the same `Update`, and the drag as the `tea.Mouse*Msg` that `shell.MouseIn` hands it. A dialog of the app's own that wants clicks implements `overlay.Clicker`, a role interface and not a second `Update`.

## Time is a message

```go
func (m model) Init() tea.Cmd { return m.spin.Tick() }

case animation.FrameMsg:
	m.spin, r = m.spin.Update(msg) // takes only its own frames

// in a test, no clock:
m.spin, _ = m.spin.Update(m.spin.FrameAt(start.Add(time.Second)))
```

Nothing reads the clock. `Tick` asks for frames, each frame is tagged so a model takes only its own, and tests feed `FrameAt(t)`. Everything that moves answers the same five: `Active()`, `Tick()`, `FrameAt(at)`, `Update` and `Advance(now) (T, bool)`, true once settled. `accordion` has no `FrameAt`, since each section keeps its own clock.

`BEZEL_NO_MOTION` holds every moving thing still. Only `timer`, `stopwatch` and the blinking caret keep their own tick, because they count real time, not motion.

## No hidden state

- No package variable is meant to be changed by an app. Per-app settings arrive through a constructor, a `With`, the theme or `shell.Config`.
- The environment is read in two places only: `icon.FromEnv()` (`BEZEL_ASCII`, `TERM=dumb`) and `animation.Still()`/`Speed()` (`BEZEL_NO_MOTION`, `BEZEL_MOTION`, `TERM=dumb`). The terminal is asked in one: `theme.DarkTerminal()`, for a program that resolves a theme without a running bubbletea program to send `tea.BackgroundColorMsg`.
- `View` never writes back. An overlay that needs its size in `Update` implements `overlay.Resizer`; whatever scrolls keeps its offset in a `layout.Window` and reports it through `Scroll()`.
- Presets (`theme.Mint()`, ...) are functions that build a fresh value each call, so no app can change one for another; point a `Theme`'s `Base` at one to build on it.

## Patterns by name

A new component reaches for one of these before inventing a shape.

| Pattern | Where |
| --- | --- |
| Model-View-Update | every model: `Update` returns the next value and a `Result`, `View` only reads |
| Composite | `layout.Node`: boxes of boxes and leaves, resolved as one tree |
| Chain of responsibility | `Shell.Update`: overlays, then the status input, then actions, then the app |
| Command | `shell.Action`: a key, a name and what it runs, as data |
| Functional options | `shell.Option`, every `With` |
| Role interfaces | `overlay.EscOwner`, `Resizer`, `StatusLiner`, `BottomAnchored`, `Clicker`, `Themer`, `form.Holder`, `form.Initer` |
| Strategy | `list.WithMatch`, `Prompt.WithValidate`, `Config.Can`, `When`: the app decides, bezel applies |
| Cascade | `theme.Theme.Base`: each layer overrides the one below |

## Compatibility

What v1 promises, and so what a change may do without a `/v2`, is on its own page: [STABILITY.md](STABILITY.md). A change that breaks it is a bug in the change.
