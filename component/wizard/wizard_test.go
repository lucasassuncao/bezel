package wizard

import (
	"errors"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel"
	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/component/form"
	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/key"
	"github.com/lucasassuncao/bezel/theme"
)

func th() theme.Resolved { return theme.Resolve(theme.Plain(), true) }

// sample is three pages; the first will not let an empty name through.
func sample() Model {
	named := func(f form.Form) error {
		if in, _ := form.Get[form.Input](f, 0); in.Value() == "" {
			return errors.New("a name is required")
		}
		return nil
	}
	return New(th(),
		NewPage("Project", form.New(th(), form.NewInput("Name").WithWidth(20))).WithCheck(named),
		NewPage("Image", form.New(th(), form.NewSelect("Image", []string{"ubuntu", "alpine"}))),
		NewPage("Review", form.New(th(), form.NewCheckbox("Start now"))),
	).WithIcons(icon.Unicode()).SetWidth(60)
}

func press(m Model, keys ...string) (Model, bezel.Result) {
	var r bezel.Result
	for _, k := range keys {
		m, r = m.Update(bezeltest.Key(k))
	}
	return m, r
}

func click(m Model, x, y int) (Model, bezel.Result) {
	return m.Update(bezel.ClickMsg{MouseClickMsg: tea.MouseClickMsg{Button: tea.MouseLeft, X: x, Y: y}})
}

func view(m Model) string { return ansi.Strip(m.View()) }

func TestTheStepsShowWhereTheWizardIs(t *testing.T) {
	m := sample()
	require.Contains(t, view(m), "1 Project ── 2 Image ── 3 Review")
	m, _ = press(m, "a", "pgdown")
	require.Contains(t, view(m), "✓ Project ── 2 Image ── 3 Review")
}

func TestACheckKeepsThePage(t *testing.T) {
	m, r := press(sample(), "pgdown")
	require.True(t, r.Handled)
	require.Nil(t, r.Event)
	require.Equal(t, 0, m.Step())
	require.Equal(t, "a name is required", m.Err())
	require.Contains(t, view(m), "✗ a name is required")

	m, _ = press(m, "a")
	require.Empty(t, m.Err(), "typing clears the error")
}

func TestNextAndBackTurnThePage(t *testing.T) {
	m, r := press(sample(), "x", "pgdown")
	require.Equal(t, Stepped{Step: 1}, r.Event)
	m, r = press(m, "pgdown")
	require.Equal(t, Stepped{Step: 2}, r.Event)
	require.Contains(t, view(m), "Finish")

	m, r = press(m, "pgup")
	require.Equal(t, Stepped{Step: 1}, r.Event)
	m, _ = press(m, "pgup")
	in, ok := form.Get[form.Input](m.Page(0).Form(), 0)
	require.True(t, ok)
	require.Equal(t, "x", in.Value(), "going back keeps what was typed")

	m, r = press(m, "pgup")
	require.True(t, r.Handled)
	require.Nil(t, r.Event, "nothing before the first page")
	require.NotContains(t, view(m), "Back")
}

func TestNextOnTheLastPageIsDone(t *testing.T) {
	m, r := press(sample(), "x", "pgdown", "pgdown", "space", "pgdown")
	require.Equal(t, Done{}, r.Event)
	require.Equal(t, 2, m.Step())
	cb, _ := form.Get[form.Checkbox](m.Page(2).Form(), 0)
	require.True(t, cb.Checked())
}

func TestAnOpenSelectKeepsThePageTurns(t *testing.T) {
	m, _ := press(sample(), "x", "pgdown", "enter")
	require.Len(t, m.Legend(), 3, "only the open select's keys")
	m, r := press(m, "pgdown")
	require.Nil(t, r.Event)
	require.Equal(t, 1, m.Step())
}

func TestClicksPressTheButtonsAndReachTheForm(t *testing.T) {
	m, _ := press(sample(), "x")
	lines := strings.Split(view(m), "\n")
	footer := len(lines) - 1
	require.Contains(t, lines[footer], "Next")

	m, r := click(m, 4, footer)
	require.Equal(t, Stepped{Step: 1}, r.Event)

	lines = strings.Split(view(m), "\n")
	footer = len(lines) - 1
	back := strings.Index(lines[footer], "Back")
	m, r = click(m, back, footer)
	require.Equal(t, Stepped{Step: 0}, r.Event)

	_, r = click(m, 0, 0)
	require.False(t, r.Handled, "the steps line takes no click")
	_, r = click(m, 50, footer)
	require.False(t, r.Handled, "past the buttons")

	m, _ = press(m, "pgdown")
	m, r = click(m, 4, formTop)
	require.True(t, r.Handled)
	sel, _ := form.Get[form.Select](m.Page(1).Form(), 0)
	require.Equal(t, "ubuntu", sel.Value())
}

func TestAFailedCheckMovesTheButtonsDown(t *testing.T) {
	m, _ := press(sample(), "pgdown")
	lines := strings.Split(view(m), "\n")
	_, r := click(m, 4, len(lines)-1)
	require.True(t, r.Handled)
	require.Equal(t, "a name is required", m.Err())
}

func TestLegendIsTheFormsThenThePageTurns(t *testing.T) {
	var got []string
	for _, e := range sample().Legend() {
		got = append(got, e.Help.Desc)
	}
	require.Equal(t, "next page", got[len(got)-2])
	require.Equal(t, "previous page", got[len(got)-1])
}

func TestWithKeysLabelsAndStyles(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m := sample().
		WithKeys(Keys{Next: key.New("ctrl+n"), Back: key.New("ctrl+p")}).
		WithLabels("Voltar", "Avançar", "Concluir").
		WithStyles(func(s Styles) Styles { s.Current = red; return s })
	require.Contains(t, m.View(), red.Render("1 Project"))
	m, _ = press(m, "x", "pgdown")
	require.Equal(t, 0, m.Step(), "pgdown is no longer bound")
	m, _ = press(m, "ctrl+n")
	require.Equal(t, 1, m.Step())
	require.Contains(t, view(m), "Voltar")
	require.Contains(t, view(m), "Avançar")
	m = m.SetTheme(theme.Resolve(theme.Mint(), true))
	require.Contains(t, m.View(), red.Render("2 Image"), "the edit outlives SetTheme")
}

func TestNarrowWidthShowsOnlyTheCurrentStep(t *testing.T) {
	m := sample().SetWidth(24)
	require.Contains(t, view(m), "1/3 Project")
	for _, w := range []int{1, 5, 24} {
		for i, l := range strings.Split(sample().SetWidth(w).View(), "\n") {
			require.LessOrEqual(t, ansi.StringWidth(l), w, "width %d row %d", w, i)
		}
	}
}

func TestEmptyAndUnsizedWizards(t *testing.T) {
	m := New(th())
	require.Empty(t, m.SetWidth(40).View())
	require.Nil(t, m.Init())
	require.Nil(t, m.Legend())
	_, r := m.Update(bezeltest.Key("pgdown"))
	require.False(t, r.Handled)
	require.Empty(t, sample().SetWidth(0).View())
	require.Equal(t, Page{}, sample().Page(9))
	require.Equal(t, "Project", sample().Page(0).Title())
}

func TestInitStartsTheShownForm(t *testing.T) {
	require.NotNil(t, sample().Init(), "the name input's caret")
}

// Whatever the titles and the width, no row of the wizard outgrows it.
func FuzzWizardStaysInItsWidth(f *testing.F) {
	f.Add("Project", "全角", 40)
	f.Add("", "", 1)
	f.Fuzz(func(t *testing.T, a, b string, w int) {
		if !utf8.ValidString(a+b) || strings.ContainsFunc(a+b, unicode.IsControl) {
			t.Skip()
		}
		w = (w%100 + 100) % 100
		m := New(th(), NewPage(a, form.New(th(), form.NewInput(b))), NewPage(b, form.New(th()))).SetWidth(w)
		for i, l := range strings.Split(m.View(), "\n") {
			require.LessOrEqual(t, ansi.StringWidth(l), w, "row %d: %q", i, l)
		}
	})
}
