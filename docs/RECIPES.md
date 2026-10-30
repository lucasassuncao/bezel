# Recipes

Short answers to "how do I...", each a task from start to end. The guides explain each part on its own; a recipe puts together the parts one task needs. Every snippet here is checked against the API on each `make test`.

## Test an app

Drive the model with the keys a terminal sends and read what it draws. No terminal, no clock:

```go
func TestDownMovesTheCursor(t *testing.T) {
	var m tea.Model = newModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(bezeltest.Key("down"))
	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "▶ second")
}
```

`bezeltest.Key` builds the press bubbletea would, from the name a binding uses: `"down"`, `"ctrl+s"`, `"?"`. Run the command an action returns and hand its message back to `Update`, as the program would.

Motion needs no waiting either. A frame is a message, and `FrameAt` builds the one for any moment:

```go
start := time.Unix(0, 0)
slide := animation.NewTween(0, 10, time.Second)
slide, _ = slide.Update(slide.FrameAt(start))
slide, _ = slide.Update(slide.FrameAt(start.Add(500 * time.Millisecond)))
// slide.Value() is where it is halfway through
```

`animation.Record` plays one to its end and keeps every frame as text, for a test that checks the whole run:

```go
frames := animation.Record(slide, func(t animation.Tween) string {
	return strings.Repeat("█", t.Value())
}, 2*time.Second)
```

Compare what a frame **says**, not its every cell: [Stability](STABILITY.md) does not promise the exact cells across versions.

## A dialog of your own

A dialog is anything with `Update`, `View` and `Legend`. The shell's overlay stack opens it over the panes, hands it the keys first, and closes it on `esc`:

```go
type rename struct {
	name string
	keys struct{ Save key.Binding }
	th   theme.Resolved
}

func newRename(name string, th theme.Resolved) rename {
	d := rename{name: name, th: th}
	d.keys.Save = key.New("enter").WithHelp("enter", "save")
	return d
}

func (d rename) Update(msg tea.Msg) (overlay.Overlay, bezel.Result) {
	if k, ok := msg.(tea.KeyPressMsg); ok && key.Matches(k, d.keys.Save) {
		return d, bezel.Result{Handled: true, Cmd: tea.Batch(overlay.Close(), save(d.name))}
	}
	return d, bezel.Result{}
}

func (d rename) View(body layout.Rect) string {
	return overlay.Box([]string{"Rename", "", d.name}, body, d.th.Modal)
}

func (d rename) Legend() []legend.Entry { return legend.FromBindings([]key.Binding{d.keys.Save}) }
```

Open it from `Update` with `overlay.Open(newRename(item, m.th))`. `overlay.Box` sizes the box to the body and `overlay.Inner(body)` says how wide its text may be. More, each by one method: `SetTheme` (`overlay.Themer`) to repaint when the theme changes, `Resize` (`overlay.Resizer`) to learn the body's size in `Update`, `Click` (`overlay.Clicker`) to take the mouse.

## A CLI without the shell

`inline`, `progress` and `logger` print into the terminal instead of taking it over, for a command that runs and ends:

```go
th := theme.Resolve(theme.Default(), theme.DarkTerminal())
log := slog.New(logger.New(os.Stderr, th))

ok, err := inline.Confirm("Move 12 files?", th)
if err != nil || !ok {
	return err
}
for i, f := range files {
	move(f)
	fmt.Print("\r" + progress.Line("moving", i+1, len(files), 60, th))
}
log.Info("moved", "files", len(files), "to", dest)
```

`theme.DarkTerminal()` asks the terminal for its background, since no bubbletea program is there to send `tea.BackgroundColorMsg`. The logger hangs every attribute under its message as a tree and wraps long values at the terminal's width.

## A read-only mode

Tag what changes things with a capability and let `Config.Can` refuse it. The key, its legend entry, its line in the help and its `:` command all go together, with no branch in each handler:

```go
sh := shell.New(shell.Config{
	// ...
	Can: func(c shell.Capability) bool { return !readOnly || c != "write" },
	Actions: []shell.Action{
		shell.Custom("d", "delete", shell.Send(deleteMsg{}), shell.Needs("write")),
		shell.Custom("r", "rename", shell.Send(renameMsg{}), shell.Needs("write")),
		shell.Help(), shell.Quit(),
	},
})
```

The [first app](EXAMPLE.md) runs this way with `--read-only`.

## `--theme` and `--list-themes`

```go
if listThemes {
	return themebrowser.BrowseInTerminal(theme.Default()) // a table to browse, or plain lines in a pipe
}
t, err := theme.Lookup(themeName) // any case; "" is the adaptive default
if err != nil {
	return fmt.Errorf("%w: --list-themes shows them", err)
}
```

Resolve `t` for the terminal's background once it answers, as [Theme](THEME.md#light-and-dark) shows, and again whenever it changes.

## Animate your own drawing

A motion is a value that moves on frames it asks for. Keep it in the model, start it, hand it its frames, and draw its `Value`:

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case openMsg:
		m.panel = animation.NewTween(m.panel.Value(), 12, 300*time.Millisecond)
		return m, m.panel.Tick()
	case animation.FrameMsg:
		var r bezel.Result
		m.panel, r = m.panel.Update(msg) // takes only its own frames
		return m, r.Cmd
	}
	return m, nil
}

func (m model) panelView() string { return strings.Repeat("█\n", m.panel.Value()) }
```

Starting from `m.panel.Value()` turns a motion around mid-flight instead of jumping. A `Spring` (`To(target)`) bends toward a new target with its speed kept. `BEZEL_NO_MOTION` makes every motion land at once, and [Animations](ANIMATIONS.md) has the rest.
