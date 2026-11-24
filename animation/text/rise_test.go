package text

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/animation"
)

func TestRiseFloatsUpAndFades(t *testing.T) {
	start := time.Unix(1000, 0)
	r := NewRise("copied!", 2, 100*ms).WithFade(white, black)
	lines := strings.Split(ansi.Strip(r.View()), "\n")
	require.Equal(t, []string{"       ", "       ", "copied!"}, lines, "it starts on its own line, the bottom")

	r.run.Ease = animation.Linear
	r, _ = r.Advance(start)
	r, _ = r.Advance(start.Add(50 * ms))
	require.Equal(t, "copied!", strings.Split(ansi.Strip(r.View()), "\n")[1])

	r, done := r.Advance(start.Add(time.Second))
	require.True(t, done)
	require.Equal(t, "       \n       \n       ", r.View(), "gone, its room kept")
}
