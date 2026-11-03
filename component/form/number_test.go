package form

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/component/icon"
	"github.com/lucasassuncao/bezel/theme"
)

func numberForm(fields ...Field) Form { return New(theme.Resolve(theme.Plain(), true), fields...) }

func TestNumberStepsWithinItsRange(t *testing.T) {
	f := numberForm(NewNumber("Port", 8080).WithRange(1, 8081).WithStep(1))
	f, _ = press(f, "up", "up")
	require.Equal(t, 8081.0, f.Field(0).(Number).Value(), "held at the top")
	f, _ = press(f, "down")
	require.Contains(t, ansi.Strip(f.SetWidth(40).View()), "‹ 8080 ›")
}

func TestNumberTakesWhatIsTypedAndSettlesOnBlur(t *testing.T) {
	f := numberForm(NewNumber("Ratio", 1).WithRange(0, 2).WithStep(0.25), NewCheckbox("x"))
	require.Contains(t, ansi.Strip(f.SetWidth(40).View()), "‹ 1.00 ›", "step's decimals")
	f, _ = press(f, "backspace", "backspace", "5", "a")
	require.Equal(t, 1.5, f.Field(0).(Number).Value())
	require.Contains(t, ansi.Strip(f.SetWidth(40).View()), "‹ 1.5 ›", "a letter is not the number's")

	f, _ = press(f, "backspace", "backspace", "backspace", "9")
	require.Equal(t, 2.0, f.Field(0).(Number).Value(), "9 held at the top")
	f, _ = press(f, "tab")
	require.Contains(t, ansi.Strip(f.SetWidth(40).View()), "‹ 2.00 ›", "settled on leaving")
}

func TestSliderMovesAlongItsBar(t *testing.T) {
	s := NewSlider("Volume", 50).WithWidth(11).WithIcons(icon.ASCII())
	f := numberForm(s)
	f, _ = press(f, "right", "right")
	require.Equal(t, 52.0, f.Field(0).(Slider).Value())
	line := strings.TrimSpace(ansi.Strip(f.SetWidth(40).View()))
	require.Equal(t, "› Volume  -----*----- 52", line)

	f = numberForm(NewSlider("v", 0).WithRange(0, 1).WithStep(0.5).WithIcons(icon.ASCII()).WithWidth(3))
	f, _ = press(f, "left", "right")
	require.Contains(t, ansi.Strip(f.SetWidth(40).View()), "-*- 0.5")
}

func TestTheLegendFollowsTheField(t *testing.T) {
	f := numberForm(NewNumber("n", 0), NewSlider("s", 0))
	require.Equal(t, "↑/↓", f.Legend()[2].Help.Key)
	f, _ = press(f, "tab")
	require.Equal(t, "←/→", f.Legend()[2].Help.Key)
}
