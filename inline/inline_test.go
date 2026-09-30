package inline

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/lucasassuncao/bezel/theme"
)

var th = theme.Resolve(theme.ThemePlain, true)

func press(c confirm, keys ...string) confirm {
	for _, k := range keys {
		var msg tea.KeyPressMsg
		switch k {
		case "enter":
			msg = tea.KeyPressMsg{Code: tea.KeyEnter}
		case "esc":
			msg = tea.KeyPressMsg{Code: tea.KeyEscape}
		case "right":
			msg = tea.KeyPressMsg{Code: tea.KeyRight}
		default:
			msg = tea.KeyPressMsg{Code: rune(k[0]), Text: k}
		}
		m, _ := c.Update(msg)
		c = m.(confirm)
	}
	return c
}

func TestConfirmAnswers(t *testing.T) {
	cases := []struct {
		name         string
		keys         []string
		yes, aborted bool
	}{
		{"enter keeps the default No", []string{"enter"}, false, false},
		{"y answers at once", []string{"y"}, true, false},
		{"n answers at once", []string{"n"}, false, false},
		{"arrow toggles then enter", []string{"right", "enter"}, true, false},
		{"esc aborts", []string{"esc"}, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := press(newConfirm("Proceed?", th), tc.keys...)
			assert.True(t, c.done)
			assert.Equal(t, tc.yes, c.yes)
			assert.Equal(t, tc.aborted, c.aborted)
		})
	}
}

func TestConfirmIgnoresVimKeys(t *testing.T) {
	for _, k := range []string{"h", "l"} {
		c := press(newConfirm("Proceed?", th), k, "enter")
		assert.False(t, c.yes, "%s does not toggle; the default No stands", k)
	}
}

func TestConfirmViewShowsAnswerWhenDone(t *testing.T) {
	c := press(newConfirm("Proceed?", th), "y")
	assert.Equal(t, "Proceed? Yes\n", ansi.Strip(c.View().Content))
}

func TestBarWidths(t *testing.T) {
	for _, tc := range []struct{ done, total, filled int }{
		{0, 10, 0}, {5, 10, 5}, {10, 10, 10}, {20, 10, 10}, {3, 0, 0},
	} {
		got := ansi.Strip(Bar(tc.done, tc.total, 10, th))
		assert.Equal(t, 10, ansi.StringWidth(got))
		assert.Equal(t, tc.filled, strings.Count(got, "█"), "%d/%d", tc.done, tc.total)
	}
}
