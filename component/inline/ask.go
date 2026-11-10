package inline

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/component/form"
	"github.com/lucasassuncao/bezel/draw"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/legend"
	"github.com/lucasassuncao/bezel/theme"
)

// Ask asks for one form field below the current line and returns it as the
// user left it: a select or multiselect opens at once, enter answers, esc or
// ctrl+c return ErrAborted, no terminal ErrNoTerminal. opts work as in Confirm.
func Ask[F form.Field](f F, th theme.Resolved, opts ...tea.ProgramOption) (F, error) {
	final, err := onTerminal(newAsk(f, th), opts)
	if err != nil {
		return f, err
	}
	a, ok := final.(ask)
	if !ok {
		return f, fmt.Errorf("inline: unexpected final model %T", final)
	}
	got, _ := form.Get[F](a.form, 0)
	if a.aborted {
		return got, ErrAborted
	}
	return got, nil
}

// The width a question draws at until the terminal reports its own.
const askWidth = 80

type ask struct {
	form    form.Form // the one field asked for
	width   int
	th      *theme.Resolved // shared by every copy and never written
	done    bool
	aborted bool
	keys    askKeys
}

// askKeys are what end the question; every other key goes to the field.
type askKeys struct {
	Accept, Cancel, Abort key.Binding
}

func defaultAskKeys() askKeys {
	return askKeys{
		Accept: key.New("enter").WithHelp("enter", "confirm"),
		Cancel: key.New("esc").WithHelp("esc", "cancel"),
		Abort:  key.New("ctrl+c"),
	}
}

// newAsk opens a field that holds the keys, such as a select, with enter.
func newAsk(f form.Field, th theme.Resolved) ask {
	a := ask{form: form.New(th, f), width: askWidth, th: &th, keys: defaultAskKeys()}
	if _, ok := f.(form.Holder); ok {
		a.form, _ = a.form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	}
	return a
}

func (a ask) Init() tea.Cmd { return a.form.Init() }

// holding reports whether the field keeps every key but ctrl+c, as an open select does.
func (a ask) holding() bool {
	h, ok := a.form.Field(0).(form.Holder)
	return ok && h.Holding()
}

// Update ends the question on Accept, or on Press closing an open list or
// calendar; while the field holds the keys, esc is its own.
func (a ask) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		return a, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, a.keys.Abort):
			return a.ended(true)
		case a.holding():
			var r bezel.Result
			a.form, r = a.form.Update(msg)
			if key.Matches(msg, a.keys.Accept) && !a.holding() {
				return a.ended(false)
			}
			return a, r.Cmd
		case key.Matches(msg, a.keys.Accept):
			// A tags field adds what was typed first; an empty enter answers.
			if _, ok := a.form.Field(0).(form.Tags); ok {
				var r bezel.Result
				if a.form, r = a.form.Update(msg); r.Event != nil {
					return a, r.Cmd
				}
			}
			return a.ended(false)
		case key.Matches(msg, a.keys.Cancel):
			return a.ended(true)
		}
	}
	var r bezel.Result
	a.form, r = a.form.Update(msg)
	return a, r.Cmd
}

// ended stops the program with the field as it is.
func (a ask) ended(aborted bool) (tea.Model, tea.Cmd) {
	a.done, a.aborted = true, aborted
	return a, tea.Quit
}

func (a ask) styles() *theme.Resolved {
	if a.th == nil {
		return &noTheme
	}
	return a.th
}

// View is the field and its keys while asked, then one line with the answer.
func (a ask) View() tea.View {
	if a.done {
		return tea.NewView(a.answer() + "\n")
	}
	body := a.form.SetWidth(a.width).View()
	return tea.NewView(body + "\n\n" + draw.Fit(a.styles().Muted.Render(a.hint()), a.width) + "\n")
}

// answer is the field's label and value, or its closed row when it has no
// label, as a checkbox, whose text is in its box. A masked value stays hidden.
func (a ask) answer() string {
	fl := a.form.Field(0)
	value := ""
	switch v := fl.(type) {
	case interface{ Values() []string }:
		value = strings.Join(v.Values(), ", ")
	case interface{ Tags() []string }:
		value = strings.Join(v.Tags(), ", ")
	case interface{ Value() string }:
		value = v.Value()
	case interface{ Value() float64 }:
		value = strconv.FormatFloat(v.Value(), 'f', -1, 64)
	case interface{ On() bool }:
		value = map[bool]string{true: "on", false: "off"}[v.On()]
	}
	if s, ok := fl.(interface{ Secret() bool }); ok && s.Secret() && value != "" {
		value = strings.Repeat("*", 6) // as long as any password, giving away none
	}
	if a.aborted {
		value = "cancelled"
	}
	if fl.Label() == "" {
		return strings.TrimRight(fl.Blur().View(false, 0, a.width), " ")
	}
	return draw.Truncate(a.styles().Bold.Render(draw.OneLine(fl.Label()))+" "+a.styles().Dim.Render(draw.OneLine(value)), a.width)
}

// hint is the field's keys without the form's tab and shift+tab, then enter
// and esc unless the field holds them.
func (a ask) hint() string {
	var parts []string
	for _, e := range a.form.Legend() {
		if e.Help.Key == "tab" || e.Help.Key == "shift+tab" {
			continue
		}
		parts = append(parts, e.Help.Key+" "+e.Help.Desc)
	}
	if !a.holding() {
		for _, e := range legend.FromBindings([]key.Binding{a.keys.Accept, a.keys.Cancel}) {
			parts = append(parts, e.Help.Key+" "+e.Help.Desc)
		}
	}
	return strings.Join(slices.Compact(parts), " · ")
}
