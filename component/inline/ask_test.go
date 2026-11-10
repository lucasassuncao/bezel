package inline

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/bezeltest"
	"github.com/lucasassuncao/bezel/component/form"
	"github.com/lucasassuncao/bezel/component/icon"
)

func answer(f form.Field, keys ...string) ask {
	a := newAsk(f, th)
	for _, k := range keys {
		m, _ := a.Update(bezeltest.Key(k))
		a = m.(ask)
	}
	return a
}

func shown(a ask) string { return ansi.Strip(a.View().Content) }

func images() form.Select {
	return form.NewSelect("Image", []string{"ubuntu", "debian", "alpine"}).WithIcons(icon.Unicode())
}

func TestAskOpensASelectAndEnterPicks(t *testing.T) {
	a := answer(images())
	require.True(t, a.holding(), "the list is open before a key")
	require.Contains(t, shown(a), "alpine")

	a = answer(images(), "down", "enter")
	require.True(t, a.done)
	require.False(t, a.aborted)
	got, _ := form.Get[form.Select](a.form, 0)
	require.Equal(t, "debian", got.Value())
	require.Equal(t, "Image debian\n", shown(a))
}

func TestAskEscClosesAnOpenListBeforeItCancels(t *testing.T) {
	a := answer(images(), "down", "esc")
	require.False(t, a.done)
	got, _ := form.Get[form.Select](a.form, 0)
	require.Equal(t, "ubuntu", got.Value(), "esc left the value as it was")

	a = answer(images(), "esc", "esc")
	require.True(t, a.aborted)
	require.Equal(t, "Image cancelled\n", shown(a))
}

func TestAskSpaceInASelectPicksWithoutEnding(t *testing.T) {
	a := answer(images(), "down", "down", "space")
	require.False(t, a.done)
	require.Contains(t, shown(a), "‹ alpine ›")
	a = answer(images(), "down", "down", "space", "enter")
	require.True(t, a.done)
}

func TestAskChecksAMultiSelect(t *testing.T) {
	ms := form.NewMultiSelect("Tools", []string{"git", "docker", "node"})
	a := answer(ms, "space", "down", "space", "enter")
	require.True(t, a.done)
	got, _ := form.Get[form.MultiSelect](a.form, 0)
	require.Equal(t, []string{"git", "docker"}, got.Values())
	require.Equal(t, "Tools git, docker\n", shown(a))
}

func TestAskTypesIntoAnInput(t *testing.T) {
	a := answer(form.NewInput("Name"), "a", "p", "i", "enter")
	got, _ := form.Get[form.Input](a.form, 0)
	require.Equal(t, "api", got.Value())
	require.Equal(t, "Name api\n", shown(a))

	a = answer(form.NewInput("Name"), "q", "esc")
	require.True(t, a.aborted, "q is text; esc cancels")
}

func TestAskNeverPrintsAMaskedAnswer(t *testing.T) {
	a := answer(form.NewInput("Password").WithMask('•'), "s", "3", "c", "r", "enter")
	got, _ := form.Get[form.Input](a.form, 0)
	require.Equal(t, "s3cr", got.Value(), "the app still reads it")
	require.Equal(t, "Password ******\n", shown(a), "six stars, whatever its length")
	require.Equal(t, "Password \n", shown(answer(form.NewInput("Password").WithMask('•'), "enter")))
}

func TestAskFlipsACheckboxAndAToggle(t *testing.T) {
	a := answer(form.NewCheckbox("Start now"), "space", "enter")
	cb, _ := form.Get[form.Checkbox](a.form, 0)
	require.True(t, cb.Checked(), "enter answers instead of flipping it back")
	require.Contains(t, shown(a), "[x] Start now")

	a = answer(form.NewToggle("Dark"), "right", "enter")
	require.Equal(t, "Dark on\n", shown(a))
}

func TestAskAddsTagsUntilAnEmptyEnter(t *testing.T) {
	a := answer(form.NewTags("Labels"), "a", "p", "i", "enter", "w", "e", "b", "enter")
	require.False(t, a.done, "enter added each tag")
	a = answer(form.NewTags("Labels"), "a", "p", "i", "enter", "enter")
	require.True(t, a.done)
	require.Equal(t, "Labels api\n", shown(a))
}

func TestAskPicksADayAndTheCalendarCloses(t *testing.T) {
	day := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	a := answer(form.NewDate("Release", day))
	require.True(t, a.holding(), "the calendar is open before a key")
	a = answer(form.NewDate("Release", day), "right", "enter")
	require.True(t, a.done)
	require.Equal(t, "Release 2026-11-10\n", shown(a))
}

func TestAskReadsANumber(t *testing.T) {
	a := answer(form.NewNumber("Port", 8080), "up", "enter")
	require.Equal(t, "Port 8081\n", shown(a))
}

func TestCtrlCAbortsEvenAnOpenList(t *testing.T) {
	require.True(t, answer(images(), "ctrl+c").aborted)
}

func TestAskHintNamesTheKeysThatWork(t *testing.T) {
	require.Contains(t, shown(answer(images())), "↑/↓ move · space press · esc close")
	closed := shown(answer(images(), "esc"))
	require.Contains(t, closed, "space press · enter confirm · esc cancel")
	require.NotContains(t, closed, "tab")
}

func TestAskDrawsAtTheTerminalsWidth(t *testing.T) {
	m, _ := newAsk(images(), th).Update(tea.WindowSizeMsg{Width: 12, Height: 10})
	for i, l := range strings.Split(shown(m.(ask)), "\n") {
		require.LessOrEqual(t, ansi.StringWidth(l), 12, "row %d", i)
	}
	require.NotNil(t, newAsk(form.NewInput("Name"), th).Init(), "the caret blinks")
	require.Equal(t, &noTheme, ask{}.styles())
}

// askRun answers Ask through opts instead of a terminal.
func askRun[F form.Field](t *testing.T, f F, opts ...tea.ProgramOption) (F, error) {
	t.Helper()
	base := []tea.ProgramOption{tea.WithOutput(io.Discard), tea.WithoutSignals()}
	return Ask(f, th, append(base, opts...)...)
}

func TestAskReadsTheAnswerFromTheGivenInput(t *testing.T) {
	got, err := askRun(t, images(), tea.WithInput(strings.NewReader("\x1b[B\r")))
	require.NoError(t, err)
	assert.Equal(t, "debian", got.Value())
}

func TestAskReturnsErrAbortedOnEsc(t *testing.T) {
	got, err := askRun(t, form.NewInput("Name").WithValue("x"), tea.WithInput(strings.NewReader("\x1b")))
	assert.ErrorIs(t, err, ErrAborted)
	assert.Equal(t, "x", got.Value(), "the field comes back as it was left")
}

func TestAskStopsWhenTheContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	pr, pw := io.Pipe()
	t.Cleanup(func() { _ = pw.Close() })
	_, err := askRun(t, images(), tea.WithInput(pr), tea.WithContext(ctx))
	assert.ErrorIs(t, err, tea.ErrProgramKilled)
}
