# A first app

This page builds a small task list from a few bezel packages: the shell around it, a list, a progress line, a prompt and a confirm. The whole program is [`examples/app/main.go`](../examples/app/main.go), about 200 lines; here it is in the order you would write it.

![demo](recordings/app.gif)

To run it:

```sh
go run ./examples/app
go run ./examples/app --read-only   # the keys that change the list are gone
```

`↑/↓` move, `space` marks a task done, `n` adds one, `d` deletes one after asking, `?` lists every key, `q` quits.

## 1. The data and the messages

The app keeps its tasks; bezel only draws them. Every key the user presses becomes a message the app handles in its own `Update`, so the messages come first:

```go
type task struct {
	title string
	done  bool
}

type (
	addMsg     struct{ title string }
	toggleMsg  struct{}
	askDropMsg struct{}
	dropMsg    struct{}
	askAddMsg  struct{}
)
```

`askAddMsg` and `askDropMsg` open a dialog; `addMsg` and `dropMsg` are what the dialog answers with.

## 2. The model

```go
type model struct {
	sh    shell.Shell
	th    theme.Resolved
	tasks []task
	list  list.Model[int] // each row carries its task's index
}
```

- `sh` is the frame: header, borders, status row, legend and dialogs.
- `th` is the resolved theme everything draws with.
- `list` is a [`list`](../component/list) whose rows carry an `int`, the index of their task, so a row leads back to the task it shows.

Everything is a value: each method that changes the shell or the list returns a new one, and the model keeps it.

## 3. The shell

```go
func newModel(readOnly bool) model {
	th := theme.Resolve(theme.Mint(), true)
	sh := shell.New(shell.Config{
		Layout: layout.Columns(
			layout.Fixed("tasks", layout.Ratio(1, 2), layout.Min(30)),
			layout.Fill("detail").Info(),
		).Collapse(70, "tasks"),
		Theme:    th,
		Title:    "tasks",
		Subtitle: "a bezel demo",
		Can:      func(c shell.Capability) bool { return c != "write" || !readOnly },
		Actions: []shell.Action{
			shell.Move(),
			shell.Custom("n", "new", shell.Send(askAddMsg{})),
			shell.Custom("space", "done", shell.Send(toggleMsg{})),
			shell.Custom("d", "delete", shell.Send(askDropMsg{}), shell.Needs("write")),
			shell.Help(),
			shell.Quit(),
		},
	})
	// ...
}
```

- **The layout** is two panes side by side, `tasks` taking half and at least 30 columns, `detail` the rest. `Info()` says the detail only informs, so focus never goes there. Below 70 columns `Collapse` keeps only the tasks.
- **The theme** is the `Mint` preset, resolved for a dark terminal until the terminal says otherwise (step 5).
- **The actions** are the keys. Each one is defined once, and the shell prints it in the legend, lists it under `?` and runs it. `shell.Send(msg)` makes a key send a message, and `Move()` only says the arrows move the cursor; the list handles them.
- **`Needs("write")` and `Can`** are how `--read-only` works: `Can` refuses `write`, and every action that needs it disappears from the keys, the legend and the help. No handler checks a flag.

The rest of `newModel` puts the first tasks in and builds the list from them.

## 4. The list

```go
func (m model) refresh() model {
	rows := make([]list.Row[int], len(m.tasks))
	for i, t := range m.tasks {
		rows[i] = list.Row[int]{Label: t.title, Mark: "○", Value: i}
		if t.done {
			rows[i].Mark = "✓"
		}
	}
	if m.list.Len() == 0 {
		m.list = list.New(rows)
	} else {
		m.list = m.list.SetRows(rows)
	}
	return m
}

func (m model) current() (int, bool) {
	if row, ok := m.list.Selected(); ok {
		return row.Value, true
	}
	return 0, false
}
```

`refresh` rebuilds the rows whenever the tasks change. `SetRows` keeps the cursor on the row it was on, by its label. `current` reads the task under the cursor back from the row's `Value`.

## 5. Update: the shell first

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if bg, ok := msg.(tea.BackgroundColorMsg); ok {
		m.th = theme.Resolve(theme.Mint(), bg.IsDark())
		m.sh = m.sh.SetTheme(m.th)
		return m, nil
	}
	var r bezel.Result
	if m.sh, r = m.sh.Update(msg); r.Handled {
		return m, r.Cmd
	}
	switch msg := msg.(type) {
	// ...
	}
	return m, nil
}
```

Every message goes to the shell first. It takes the keys its actions own and every key while a dialog is open; `Handled` true means there is nothing left for the app. A resize it lays out and hands on (`Handled` false), so the app can size its own models from `Rect`. Then the app handles what is its own.

The background colour comes before it all: `Init` asks the terminal for it, and the answer resolves the theme again, light or dark. [THEME.md](THEME.md#light-and-dark) has the details.

## 6. A prompt and a confirm

`n` sent `askAddMsg`. The app answers it by opening a [prompt](../overlay):

```go
case askAddMsg:
	ask := overlay.NewPrompt("New task", func(title string) tea.Cmd {
		return func() tea.Msg { return addMsg{title: title} }
	}, m.th).WithValidate(func(typed string) error {
		if strings.TrimSpace(typed) == "" {
			return errors.New("a task needs a title")
		}
		return nil
	})
	return m, overlay.Open(ask)
case addMsg:
	m.tasks = append(slices.Clone(m.tasks), task{title: strings.TrimSpace(msg.title)})
	m = m.refresh()
	m.sh, cmd = m.sh.SetStatus("added "+msg.title, shell.OK, 2*time.Second)
	return m, cmd
```

`overlay.Open` is a command: the shell pushes the dialog, which takes every key until `enter` or `esc`. `WithValidate` refuses an empty title. On `enter` the prompt runs the function it was given, which sends `addMsg`, and the app adds the task and says so on the status row for two seconds.

Deleting asks first, the same way, with a confirm drawn in the danger colour:

```go
case askDropMsg:
	if i, ok := m.current(); ok {
		yes := func() tea.Msg { return dropMsg{} }
		ask := overlay.NewConfirm("Delete task", m.tasks[i].title, yes, m.th).WithKind(draw.Danger)
		return m, overlay.Open(ask)
	}
```

`toggleMsg` flips the task under the cursor and refreshes the list, and `dropMsg` removes it. Neither needs a dialog.

## 7. The keys left over

```go
case tea.KeyPressMsg:
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	m.list = m.list.SetHeight(draw.InnerRect(m.sh.Rect("tasks")).H)
	m.list, _ = m.list.Update(msg)
```

A key no action took reaches the app, which hands it to the list: the arrows, `home`, `end`. The list needs to know how many rows it shows to scroll right, so it is sized to the inside of its pane first. `m.sh.Rect("tasks")` is where the shell put the pane, and `draw.InnerRect` takes the border off.

## 8. View: the panes

```go
func (m model) View() tea.View {
	l := m.list.SetHeight(draw.InnerRect(m.sh.Rect("tasks")).H)
	return m.sh.View(map[string]shell.Pane{
		"tasks": {
			Title:  "tasks",
			Bottom: fmt.Sprintf("%d/%d done", m.doneCount(), len(m.tasks)),
			Scroll: l.Scroll(),
			Body:   func(layout.Rect) string { return l.View(m.th) },
		},
		"detail": {Title: "detail", Body: m.detail},
	})
}
```

The app hands the shell one `Pane` per leaf of the layout. A pane has a title, text for the bottom edge, a scrollbar from the list, and a `Body`: a function the shell calls with the rect inside the border, once it knows the size. The shell draws everything else, including the header, the borders, the focus, the status row and the legend.

The detail pane draws the task under the cursor, and a [`progress`](../component/progress) line for the whole list:

```go
func (m model) detail(r layout.Rect) string {
	i, ok := m.current()
	if !ok {
		return draw.EmptyState(r, "nothing to do", "n adds a task", m.th.Chrome)
	}
	t := m.tasks[i]
	state := m.th.Warning.Render("to do")
	if t.done {
		state = m.th.Success.Render("done")
	}
	return strings.Join([]string{
		m.th.Accent.Render(t.title),
		"",
		m.th.Dim.Render("state  ") + state,
		"",
		progress.Line("all tasks", m.doneCount(), len(m.tasks), r.W, m.th),
	}, "\n")
}
```

Every colour comes from the theme: `Accent`, `Dim`, `Warning`, `Success`. Switching the theme recolours the app's own text along with bezel's. With no task left, `draw.EmptyState` centres a message and a hint in the pane.

## 9. main

```go
func main() {
	readOnly := flag.Bool("read-only", false, "hide the keys that change the list")
	flag.Parse()
	if _, err := tea.NewProgram(newModel(*readOnly)).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

## Testing it

The model is a value and every input a message, so a test presses keys the way a terminal would. From [`main_test.go`](../examples/app/main_test.go):

```go
func TestDeleteAsksFirst(t *testing.T) {
	m := press(sized(t, false), "d")
	require.Contains(t, ansi.Strip(m.View().Content), "Delete task")
	m = press(m, "y")
	out := ansi.Strip(m.View().Content)
	require.Contains(t, out, "deleted write the release notes")
	require.Contains(t, out, "0/3 done")
}
```

`press` sends a [`bezeltest.Key`](../bezeltest) and feeds back whatever the commands it starts send, skipping the timers. The test file has it in full.

## Where next

- [FRAME.md](FRAME.md): tabs, focus between panes, the command palette and dialogs of your own.
- [THEME.md](THEME.md): another preset, a theme of your own, recolouring one component.
- [COMPONENTS.md](COMPONENTS.md): every other component, each with its example.
- [EFFECTS.md](EFFECTS.md) and [ANIMATIONS.md](ANIMATIONS.md): motion, when the app wants it.
