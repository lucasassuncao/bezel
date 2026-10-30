package progress

import (
	"github.com/lucasassuncao/bezel/component/chart"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/theme"
)

func th() theme.Resolved { return theme.Resolve(theme.Plain(), true) }

func TestBarFillsInProportion(t *testing.T) {
	require.Equal(t, "█████░░░░░", ansi.Strip(Bar(5, 10, 10, th())))
	require.Equal(t, "░░░░░░░░░░", ansi.Strip(Bar(3, 0, 10, th())), "zero total is nothing done")
	require.Equal(t, "██████████", ansi.Strip(Bar(20, 10, 10, th())), "overshoot stays in the bar")
}

func TestLineIsExactlyTheWidth(t *testing.T) {
	for _, width := range []int{12, 30, 80} {
		got := Line("installing", 3, 8, width, th())
		require.Equal(t, width, ansi.StringWidth(got), "width %d", width)
		require.Contains(t, ansi.Strip(got), "3/8", "the count survives at %d", width)
	}
	require.Contains(t, ansi.Strip(Line("installing", 3, 8, 40, th())), "█")
	require.NotContains(t, ansi.Strip(Line("installing", 3, 8, 20, th())), "█", "too narrow for a bar")
}

func TestStepperMarksEachStage(t *testing.T) {
	steps := []string{"Source", "Options", "Review"}
	got := ansi.Strip(Stepper(steps, 1, 80, th(), icon.Unicode()))
	require.Contains(t, got, "✓ Source ─ ● Options ─ ○ Review")
	require.Equal(t, 80, ansi.StringWidth(Stepper(steps, 1, 80, th(), icon.Unicode())))

	done := ansi.Strip(Stepper(steps, len(steps), 80, th(), icon.ASCII()))
	require.Contains(t, done, "+ Source - + Options - + Review", "the ASCII set joins with ASCII too")
}

func TestStepperCompactsWhenItDoesNotFit(t *testing.T) {
	steps := []string{"Choose a template", "Pick the features", "Review and write"}
	got := ansi.Strip(Stepper(steps, 1, 40, th(), icon.Unicode()))
	require.Equal(t, 40, ansi.StringWidth(got))
	require.Contains(t, got, "Step 2 of 3 · Pick the features")

	require.Contains(t, ansi.Strip(Stepper(steps, 3, 30, th(), icon.Unicode())), "All 3 steps done")
	require.Empty(t, Stepper(nil, 0, 30, th(), icon.Unicode()))
}

// rich is a theme with a whole series and ramp of distinct colours.
var rich = theme.Theme{Colors: theme.Colors{
	Accent: "#FE8019", Selection: "#FABD2F", Border: "#7C6F64", Dim: "#A89984", Text: "#EBDBB2", Background: "#282828",
	Success: "#B8BB26", Warning: "#FABD2F", Danger: "#FB4934", Info: "#83A598", OnAccent: "#282828",
	Series: [6]string{"#B8BB26", "#8EC07C", "#FE8019", "#83A598", "#FABD2F", "#D3869B"},
	Ramp:   [3]string{"#B8BB26", "#FABD2F", "#FE8019"},
}}

func richTheme() theme.Resolved { return theme.Resolve(rich, true) }

func TestMeterIsSquaresFilledAlongTheRamp(t *testing.T) {
	require.Equal(t, "■■■■■□□□□□", ansi.Strip(Meter(5, 10, 10, th())))
	require.Equal(t, "□□□□", ansi.Strip(Meter(1, 0, 4, th())), "zero total is nothing done")
	full := Meter(9, 9, 9, richTheme())
	for _, st := range richTheme().Ramp {
		require.Contains(t, full, st.Render("■■■"))
	}
}

func TestSmoothSplitsACellInEighths(t *testing.T) {
	require.Equal(t, "█▌░░", ansi.Strip(Smooth(3, 8, 4, th())))
	require.Equal(t, "▏░", ansi.Strip(Smooth(1, 16, 2, th())))
	require.Equal(t, "░░░░", ansi.Strip(Smooth(0, 0, 4, th())), "zero total is nothing done")
	require.Equal(t, "████", ansi.Strip(Smooth(9, 8, 4, th())), "overshoot stays in the bar")
}

func TestSmoothInASeriesColour(t *testing.T) {
	require.Contains(t, Smooth(8, 8, 2, richTheme(), chart.InSeries(4)), richTheme().Series[4].Render("██"))
}

func TestANilOptionIsSkipped(t *testing.T) {
	require.Equal(t, ansi.Strip(Smooth(1, 2, 8, th())), ansi.Strip(Smooth(1, 2, 8, th(), nil)))
}

// Every bar keeps to its width, whatever done and total say.
func FuzzBarsStayInTheirWidth(f *testing.F) {
	f.Add("installing", 3, 8, 20)
	f.Add("全角のラベル", -4, 0, 5)
	f.Add("", 9, 3, 0)
	f.Fuzz(func(t *testing.T, label string, done, total, w int) {
		if !appText(label) {
			t.Skip()
		}
		w %= 120
		for name, out := range map[string]string{
			"bar":    Bar(done, total, w, th()),
			"line":   Line(label, done, total, w, th()),
			"meter":  Meter(done, total, w, th()),
			"smooth": Smooth(done, total, w, th()),
			"model":  New().SetWidth(w).ViewAs(float64(done)/float64(max(total, 1)), th()),
		} {
			require.LessOrEqual(t, ansi.StringWidth(out), max(w, 0), "%s: %q", name, out)
		}
	})
}

// appText reports text an app hands a component as it is: valid UTF-8 with
// no control characters but line breaks. Anything else goes through
// draw.Sanitize first, the app's job.
func appText(ss ...string) bool {
	for _, s := range ss {
		if !utf8.ValidString(s) || strings.ContainsFunc(s, func(r rune) bool { return unicode.IsControl(r) && r != '\n' }) {
			return false
		}
	}
	return true
}
