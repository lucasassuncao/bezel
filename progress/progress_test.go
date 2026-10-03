package progress

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/icon"
	"github.com/lucasassuncao/bezel/theme"
)

func th() theme.Resolved { return theme.Resolve(theme.ThemePlain, true) }

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
	got := ansi.Strip(Stepper(steps, 1, 80, th(), icon.Unicode))
	require.Contains(t, got, "✓ Source ─ ● Options ─ ○ Review")
	require.Equal(t, 80, ansi.StringWidth(Stepper(steps, 1, 80, th(), icon.Unicode)))

	done := ansi.Strip(Stepper(steps, len(steps), 80, th(), icon.ASCII))
	require.Contains(t, done, "+ Source - + Options - + Review", "the ASCII set joins with ASCII too")
}

func TestStepperCompactsWhenItDoesNotFit(t *testing.T) {
	steps := []string{"Choose a template", "Pick the features", "Review and write"}
	got := ansi.Strip(Stepper(steps, 1, 40, th(), icon.Unicode))
	require.Equal(t, 40, ansi.StringWidth(got))
	require.Contains(t, got, "Step 2 of 3 · Pick the features")

	require.Contains(t, ansi.Strip(Stepper(steps, 3, 30, th(), icon.Unicode)), "All 3 steps done")
	require.Empty(t, Stepper(nil, 0, 30, th(), icon.Unicode))
}
