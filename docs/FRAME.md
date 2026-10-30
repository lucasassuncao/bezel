# The frame

The frame is what bezel draws around an app: the header and its tabs, the panes' borders, the status row, the key legend, and the dialogs and command line that open over the panes. The app draws only what goes inside the panes. Five packages make it, all at the module's root:

| Package | What it does |
|---|---|
| [`shell`](../shell) | the struct an app embeds: resize, tabs, focus, status, the overlay stack, the legend and the keys |
| [`layout`](../layout) | the screen as a tree of named panes, resolved to rects every frame |
| [`overlay`](../overlay) | the stack of dialogs, and the ready-made ones |
| [`palette`](../palette) | the `:` command line |
| [`legend`](../legend), [`key`](../key) | the key hints at the foot of the screen, and the bindings they are made from |

Each package's example and recording is in [COMPONENTS.md](COMPONENTS.md#the-frame); this page is how they fit together.

## An app in four parts

```go
type model struct {
	sh shell.Shell
}

func newModel() model {
	return model{sh: shell.New(shell.Config{
		Layout: layout.Columns(
			layout.Fixed("list", layout.Ratio(1, 3), layout.Min(24)),
			layout.Fill("detail"),
		),
		Theme:   theme.Resolve(theme.Mint, true),
		Title:   "my app",
		Actions: []shell.Action{shell.Help(), shell.ChangePane(), shell.Move(), shell.Quit()},
	})}
}

func (m model) Init() tea.Cmd { return tea.RequestBackgroundColor }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var handled bool
	var cmd tea.Cmd
	if m.sh, handled, cmd = m.sh.Update(msg); handled {
		return m, cmd
	}
	// the messages the app's actions send, and the keys no action claims
	return m, nil
}

func (m model) View() tea.View {
	return tea.NewView(m.sh.View(map[string]shell.Pane{
		"list":   {Title: "Items", Body: m.list},
		"detail": {Title: "Detail", Body: m.detail},
	}))
}
```

1. **`shell.Config`** says everything up front: the layout, the theme, the title, the tabs and the keys that hold on every screen.
2. **`Update`** hands every message to the shell first. `handled` true means it was the shell's (a key an action took, a dialog's) and the app returns at once; false means the app's turn. A resize is both: the shell lays the panes out and hands it on, so the app sizes its models from `Rect` on the same message.
3. **`View`** hands the shell a `Pane` per leaf of the layout: a title and a `Body` that draws into the inner rect the shell computed. A leaf missing from the map gives its space back that frame.
4. **`Init`** asks the terminal for its background, so the theme can be resolved [light or dark](THEME.md#light-and-dark).

The shell is a value, like every bezel model: each method that changes it returns the changed copy, and the app keeps it in its own model.

## Layout

A layout is a tree: `layout.Columns` and `layout.Rows` split their space among their children, and a leaf is a named pane.

```go
layout.Rows(
	layout.Fixed("search", layout.Lines(3)),
	layout.Columns(
		layout.Fixed("list", layout.Ratio(1, 3), layout.Min(26), layout.Max(52)),
		layout.Fill("detail"),
		layout.Fixed("hints", layout.Lines(30)).Info(),
	).Collapse(90, "detail"),
)
```

- **`Fixed`** sizes a leaf by its constraints: `Lines(n)` exactly, `Ratio(num, den)` as a share of its parent, clamped by `Min` and `Max`. **`Fill`** takes what its fixed siblings leave.
- **`Collapse(limit, keep)`** keeps only the child holding `keep` while the box is narrower than `limit`: a two-column screen that becomes one on a small terminal.
- **`Info()`** marks a pane that only informs: `ChangePane` skips it.
- A leaf the app leaves out of `View`'s map is gone for that frame, and its space goes to the others. A pane that only shows sometimes is a leaf the app sometimes leaves out.

`SetLayout` swaps the whole tree while the app runs, and `Rect(name)` reports where a pane landed, for a click or a scroll that needs its size.

## Tabs

An app with more than one screen hands the shell its tabs. Each answers for its own screen:

```go
type filesTab struct{ count int }

func (t filesTab) Name() string                { return "files" }
func (t filesTab) Status(shell.Context) string { return fmt.Sprintf("%d files", t.count) }
func (t filesTab) Actions(shell.Context) []shell.Action {
	return []shell.Action{shell.Custom("d", "delete", shell.Send(deleteMsg{}), shell.Needs("write"))}
}
```

`Status` is the text on the status row while the tab shows; `Actions` are its keys, added to `Config.Actions`. Both run several times a frame, so they stay cheap and change nothing. A tab is read as given: when it changes, the app hands it back with `SetTabs`.

`ChangeTab` moves between tabs on `tab` and `shift+tab`; `SetTab(i)` moves from code, and `ActiveTab()` says where the user is. When the tabs are wider than the screen the strip scrolls to keep the active one in sight, with `‹` and `›` where more are hidden. `Config.Transition` plays an [effect](EFFECTS.md#between-tabs) over the panes as the tab changes.

## Actions

An `Action` is one key, defined once: the shell dispatches it, prints it in the legend, lists it in the help and offers it in the command palette.

| Action | Key | What it does |
|---|---|---|
| `Help()` | `?` | opens every key the screen has, in a dialog |
| `ChangeTab()` | `tab` | the next tab, `shift+tab` the one before |
| `ChangePane()` | `tab` | the next pane, sending a `FocusMsg` |
| `Move()`, `Scroll()` | `↑/↓` | listed only: the focused component owns the arrows |
| `Commands()` | `:` | opens the command palette |
| `Quit()` | `q` | quits |
| `Custom(keys, desc, run)` | `keys` | the app's own |
| `Command(name, desc, run)` | none | palette only, as `:name` |

What an action runs is a function of an `ActionContext` that returns a command. `shell.Send(msg)` is the usual one: the action emits `msg`, and the app handles it in its own `Update`, as it would any message. An action never changes the app's copy of the shell; whatever it would change, it says with a message.

Options shape an action without a new constructor:

| Option | What it does |
|---|---|
| `WithKey(keys...)` | binds more keys under one label, as `←/→` |
| `Named(name)`, `Arg(placeholder)` | puts it in the palette as `:name`, taking what follows as `Arg` |
| `Needs(capability)` | drops it where the session refuses the capability |
| `When(func(Context) bool)` | keeps it only while the function says so: a key that only means something in one pane |
| `Group(n)` | starts a new legend row for group `n` |
| `DisplayOnly()` | lists it but leaves its key to the focused component |
| `RunWith(run)` | keeps the key and text and replaces what it does: a `Quit` that asks first |
| `KeyHint(k)` | prints a key beside the command in the palette without binding it |

`shell.Check(actions...)` reports two actions on one key. Call it in a test over each screen's set: at run time the first match wins silently.

### Capabilities

`Config.Can` decides what the session allows. An action that `Needs("write")` is gone from the keys, the legend, the help and the palette while `Can("write")` is false: a read-only mode without a branch in every handler.

```go
shell.Config{
	// ...
	Can: func(c shell.Capability) bool { return c != "write" || m.writable },
}
```

## Focus

```go
case shell.FocusMsg: // sent on every move, by key or click
	if msg.To == "detail" {
		m = m.loadDetail() // the app's own: what entering the pane means
	}
case tea.KeyPressMsg: // keys no action took go to the focused pane
	if m.sh.Focused() == "list" {
		m.list, _ = m.list.Update(msg)
	}
```

The shell knows which pane has focus and draws its border in the accent. `Focused()` names it, `SetFocus(name)` moves it, and `ChangePane` steps through the layout's leaves, skipping the `Info` ones. Each move sends a `shell.FocusMsg{From, To}`, so the app can run what entering and leaving a pane means for it. With `Config.FocusFade` set, the accent colour travels from the pane left to the pane gained over that long instead of jumping. Keys the actions leave alone fall through to the app, which hands them to the component in the focused pane.

## The mouse

The mouse is off until the app asks for it. Two lines turn it on:

```go
shell.New(shell.Config{ /* ... */ Mouse: true})

func (m model) View() tea.View {
	return m.sh.Screen(panes) // the full-screen view, the mouse reporting on
}
```

From then on the shell answers a left click by itself, in `Update`, like a key an action takes:

- on a **tab**, it shows that tab, with `Config.Transition` if there is one;
- on a **dialog**, the dialog takes it: `Confirm`, `Alert` and `Choice` press the button clicked, and a click off the dialog does nothing;
- on a **pane**, it focuses the pane, with the `FocusMsg` and the fade a key gives (an `Info` pane keeps the focus where it was), and sends the app a `shell.ClickMsg`.

`ClickMsg` names the pane and carries the cell inside its body, the rect its `Body` draws in, so the app hands it on the way it hands on a key:

```go
case shell.ClickMsg:
	switch msg.Pane {
	case "list":
		m.list, act = m.list.Click(msg.Y) // act is list.Chosen on a second click
	case "form":
		m.form, cmd = m.form.Click(msg.Y) // the width SetWidth gave it says how tall an open select is
	case "notes":
		m.notes, cmd = m.notes.Update(msg.MouseClickMsg) // a textbox takes the press itself
	}
```

The components' `Click` moves the cursor to the row clicked, and on the row already under it does what enter would: `list` and `table` return `Chosen`, `tree` opens or closes the node. `form.Click` focuses the field and presses a button, a checkbox or a toggle.

A click is all `ClickMsg` carries. The rest of the mouse reaches the app as it came, and two methods move it into a pane:

- **`Mouse(msg)`** finds the pane under the pointer and moves `msg` into its body: the wheel over a [`viewport`](../component/viewport), which scrolls on `tea.MouseWheelMsg`.
- **`MouseIn(pane, msg)`** moves `msg` into a pane wherever the pointer is: a drag that began in a [`textbox`](../component/textbox) and left it still moves its caret.

A dialog of the app's own takes clicks by implementing `overlay.Clicker`: `Click(x, y, body)`, with `x` and `y` counted from the top left of what its `View` drew.

## The status row

```go
case saveMsg:
	m.sh, cmd = m.sh.SetStatus("saved", shell.OK, 2*time.Second)
case indexMsg:
	m.sh, cmd = m.sh.Busy("indexing") // a spinner until Idle
	return m, tea.Batch(cmd, index())
case indexedMsg:
	m.sh = m.sh.Idle()
```

The row under the panes says what is happening:

- `SetStatus(text, level, ttl)` shows a message, `shell.Info`, `shell.OK` or `shell.Error`, and clears it after `ttl`, unless a newer one replaced it.
- `Busy(text)` shows a spinner beside the text until `Idle()`: work in progress.
- With nothing set, it shows the active tab's `Status`.
- `SetInput(m)` puts an [`input`](../component/input) on the row for a one-line answer; `Input()` reads it back after `Update`.

`Copy(text, done, failed)` writes to the clipboard and reports the result on the row.

`StatusStyle` in the `Config` says where these messages show. `shell.StatusBar`, the default, is the row above. With `shell.Toast` each one slides into the top right corner as a [`toast`](../component/toast), and the row gives its line back to the panes: it shows only while a line is typed on it, the palette's or `SetInput`'s, while `Busy` runs, or while the tab's `Status` has text. `Status()` still reads the newest message.

## Overlays

Dialogs float over the panes in a stack. The top one takes every key, `esc` closes it, and the legend shows its keys while it is open.

```go
shell.Custom("d", "delete", func(shell.ActionContext) tea.Cmd {
	yes := func() tea.Msg { return deleteMsg{} }
	return overlay.Open(overlay.NewConfirm("Delete", "Delete main.go?", yes, th))
})
```

`overlay.Open(o)` is the command that pushes a dialog, from an action or anywhere a command goes; `Push(o)` and `Pop()` do it on the shell directly. A dialog closes itself by returning `overlay.Close()`, and hands its answer back the way an action does, as a command the app receives. The ready-made ones are `Alert`, `Confirm`, `Prompt`, `Choice`, `Pick`, `Text`, `Help`, `Pager` and `FilePicker`.

A dialog of the app's own is any type with `Update`, `View(body)` and `Legend()`. Small interfaces opt it into more: `Resizer` hears the body's size before it draws, `EscOwner` keeps `esc` for itself, `StatusLiner` draws the status row while on top, `BottomAnchored` sits at the foot of the screen instead of the middle, and `Clicker` takes [the mouse](#the-mouse).

## The palette and the legend

```go
shell.Command("goto", "open a path", func(c shell.ActionContext) tea.Cmd {
	return func() tea.Msg { return gotoMsg{path: c.Arg} } // ":goto kv/app" gives Arg "kv/app"
}, shell.Arg("path"))
```

`:` opens the command palette at the foot of the screen. It lists every action with a name, completes it with `tab`, and passes what follows the name to the action as `ActionContext.Arg`: `:goto kv/app` runs the action `Named("goto")` with `Arg` `kv/app`. An unknown command, or one the session refuses, says so on the status row.

The legend is built from the same actions, filtered by `Needs` and `When`, and packed into `Config.LegendLines` rows, two by default. Keys that do not fit are counted, and `?` lists them all. A component lists its own keys through its `Legend()`, from the same bindings it answers to, so the hint never drifts from the key.

## Environment

bezel reads the environment in two places, when asked:

| Variable | Effect |
|---|---|
| `BEZEL_ASCII` | icons and marks in ASCII instead of Unicode |
| `BEZEL_NO_MOTION` | every animation born at its end, no frames asked |
| `BEZEL_MOTION` | the speed of motion: `2` twice as fast, `0` off |
| `TERM=dumb` | both of the above: ASCII and no motion |

They are for a user who reads better without motion or symbols, a slow link, or a CI log. An app has nothing to wire: every component asks when it is built.

## Testing an app

Every model is a value and every input a message, so a test drives the app as the terminal would:

```go
m := newModel()
next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
next, _ = next.Update(bezeltest.Key("tab"))
next, _ = next.Update(bezeltest.Key("d"))
```

- [`bezeltest.Key`](../bezeltest) builds the key press a terminal sends for `"tab"`, `"ctrl+s"` or `"q"`.
- Commands are functions: call one to get the message it would send, and feed that back to `Update`.
- An animation asks for frames with `Tick`; a test hands it its own with `FrameAt(time)`, so no test waits on a clock. `animation.Record` plays one to its end as text. See [ANIMATIONS.md](ANIMATIONS.md).
- Set `BEZEL_NO_MOTION=1` in a test that only cares where things land, and clear it in `TestMain` when a CI sets `TERM=dumb`.
- `shell.Check` over each screen's actions catches two keys bound twice.
- A click on a pane comes back from `Update` as a command: call it for the `shell.ClickMsg`, and feed that to the app.
