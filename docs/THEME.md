# Theme

Every colour bezel draws comes from one place, the [`theme`](../theme) package. An app picks a theme, resolves it for the terminal it runs in, and hands the result to the shell and to every component. Nothing else holds a colour of its own.

```go
th := theme.Resolve(theme.Mint, true) // a preset, for a dark terminal
sh := shell.New(shell.Config{Theme: th /* ... */})
body := items.View(th)                  // a component draws under the theme it is handed
```

## Roles, not colours

A theme names what a colour is *for*. `theme.Colors` has one field per role:

| Role | What it colours |
|---|---|
| `Accent` | focused borders, section headings, legend keys |
| `Selection` | the cursor row, the focused panel's title |
| `Border` | unfocused borders, status and hint text |
| `Dim` | secondary text, items not yet present |
| `Text` | body text |
| `Success`, `Warning`, `Danger`, `Info` | what a thing reports: healthy, a draft, an error, a count |
| `OnAccent` | text drawn on a filled accent, success, warning or danger |
| `Series` | six colours that tell things of one kind apart, as a dashboard's panels |
| `Ramp` | three colours that climb from low to high, as a meter fills |

Each value is a lipgloss colour string: a hex value (`"#7C3AED"`) or an ANSI code (`"63"`). An empty one falls back, so a theme sets only what it means to change: one that sets `Accent` alone also gets it for `Selection` and `Info`, and one without a `Ramp` climbs through its own state colours.

## Presets

bezel ships 59 presets, in categories:

| Category | Presets |
|---|---|
| Miscellaneous | `Default`, `Plain`, `Terminal`, `Gruvbox` |
| Fruit | `Mint`, `Banana`, `Strawberry`, `Blueberry`, … 28 in all |
| Horizon, Super Mario, Sonic | one per character |

Three are for when colour itself is the question:

- **`Default`** leaves every role empty, so each follows the terminal's light or dark background.
- **`Plain`** uses only the 16 ANSI colours, for a terminal with few.
- **`Terminal`** takes every colour from the terminal's own palette, so the app matches the scheme the user already runs.

For a `--theme` flag, `theme.Lookup(name)` finds a preset by name in any case, empty being `Default`, and `theme.All()` and `theme.Categories()` list them. For `--list-themes`, [`themebrowser`](../component/themebrowser) draws every preset with its palette:

```go
if listThemes {
	return themebrowser.BrowseInTerminal() // a scrollable table, or plain lines in a pipe
}
t, err := theme.Lookup(flagTheme)
```

## A theme of your own

A `Theme` layers: `Base` points at the theme under it, and every role it sets overrides that one. A preset is a base to build on, never something to change:

```go
brand := theme.Theme{
	Base:   &theme.Mint,
	Colors: theme.Colors{Accent: "#FF5F87", Danger: "#D7005F"},
}
```

Everything not set comes from `Mint`, and what `Mint` leaves empty from the adaptive default. A `Base` may have a `Base` of its own, so a company theme and a product's tweak of it are two layers.

`Theme.Styles` goes one step further, for the few elements a colour does not settle: `Cursor`, `Muted` and `Danger` take a whole lipgloss style that replaces the one derived from the colours.

## Light and dark

`theme.Resolve(t, dark)` turns a theme into the styles everything draws with, for a dark or a light background. The terminal says which: ask in `Init`, and resolve again when it answers.

```go
func (m model) Init() tea.Cmd { return tea.RequestBackgroundColor }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if bg, ok := msg.(tea.BackgroundColorMsg); ok {
		m.th = theme.Resolve(m.theme, bg.IsDark())
		m.sh = m.sh.SetTheme(m.th)
	}
	// ...
}
```

The answer can come after the first frame, so the app starts from a guess, usually dark, and switches once. Code that resolves a theme outside a running program, a `--list-themes` or a plain CLI, asks with `theme.DarkTerminal()`.

## Drawing with it

`theme.Resolved` holds the colours it came from and a lipgloss style for each use:

- text that says what it is: `Success`, `Warning`, `Danger`, `Info`, `Accent`, `Dim`, `Muted`, `Bold`, `Key`;
- lists and tables: `Section`, `TableHeader`, `Cursor`, `Badge`, and `Match` and `MatchCurrent` for search hits;
- the frame's own: `Panel`, `PanelFocused`, `Chrome`, `Legend`, `StatusOK`, `StatusErr`, `Modal`, `Button`;
- `Series` and `Ramp`, the colours of the same name as foreground styles.

An app draws its pane bodies with these, so a theme switch recolours the app's text along with bezel's:

```go
func (m model) detail(r layout.Rect) string {
	return m.th.Dim.Render("modified ") + m.th.Info.Render(m.when)
}
```

`CaretBlink` is how fast every text field built with the theme blinks its caret; a test shrinks it there.

## Where the theme goes

Who calls `View` decides where a component takes its theme, and [CONVENTIONS.md](CONVENTIONS.md#the-theme) has the full table:

- the shell, from `Config.Theme`, replaced with `SetTheme`;
- a component the app draws, in `View(th, ...)`, so one model draws under any theme;
- a dialog or a form field, which bezel draws, as its constructor's last argument;
- a plain function, such as a progress bar or a chart, after the data it draws.

## One component, recoloured

Where one component should look different from the theme, without a new theme for everything, its `WithStyles` edits the styles derived from the theme:

```go
sp := spinner.New(spinner.Dot()).WithStyles(func(s spinner.Styles) spinner.Styles {
	s.Frame = s.Frame.Foreground(lipgloss.Color("#FF5F87"))
	return s
})
```

The edit is kept, not the result: it applies again on every theme, so the spinner stays pink through a switch from light to dark. `DefaultStyles(th)` is what the edit starts from.
