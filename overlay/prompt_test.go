package overlay

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/lucasassuncao/bezel/layout"
	"github.com/lucasassuncao/bezel/legend"
)

type acceptedMsg struct{ value string }

func newPrompt() Prompt {
	return NewPrompt("New secret", func(v string) tea.Cmd {
		return func() tea.Msg { return acceptedMsg{v} }
	}, lipgloss.NewStyle(), legend.Style{})
}

func typeInto(p Prompt, s string) Prompt {
	for _, r := range s {
		o, _ := p.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		p = o.(Prompt)
	}
	return p
}

func promptView(p Prompt) string { return ansi.Strip(p.View(layout.Rect{W: 80, H: 30})) }

func TestPromptAcceptsWhatWasTyped(t *testing.T) {
	p := typeInto(newPrompt(), "db")
	require.Equal(t, "db", p.Value())

	_, cmd := p.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	msgs := drain(cmd())
	require.Contains(t, msgs, tea.Msg(acceptedMsg{"db"}))
	require.Contains(t, msgs, tea.Msg(CloseMsg{}), "accepting closes the prompt")
}

// A typed confirmation must be reproduced exactly: enter does nothing until
// it is, and the hint says what unlocks it instead of offering enter.
func TestPromptWithRequireLocksEnterUntilTheTextMatches(t *testing.T) {
	p := newPrompt().WithRequire("prod")
	require.False(t, p.Accepts())
	require.Contains(t, promptView(p), "type the exact text to unlock")

	p = typeInto(p, "pro")
	_, cmd := p.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Nil(t, cmd, "a near miss must not run the action")

	p = typeInto(p, "d")
	require.True(t, p.Accepts())
	require.NotContains(t, promptView(p), "unlock")
	require.Contains(t, promptView(p), "[enter]")
}

func TestPromptWithValidateShowsTheProblemAndBlocks(t *testing.T) {
	p := newPrompt().WithValidate(func(v string) string {
		if strings.Contains(v, " ") {
			return "no spaces in a name"
		}
		return ""
	})
	p = typeInto(p, "a b")
	require.False(t, p.Accepts())
	require.Contains(t, promptView(p), "no spaces in a name")
	_, cmd := p.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Nil(t, cmd)
}

// A labelled prompt leads with its field and the preview under it; a bare
// one reads like a typed confirmation, the warning first and the input last.
func TestPromptLayoutFollowsTheLabel(t *testing.T) {
	labelled := newPrompt().WithLabel("name").WithLines("A slash makes folders.").
		WithPreview("path", func(v string) string { return "kv/" + v })
	labelled = typeInto(labelled, "db")
	v := promptView(labelled)
	require.Less(t, strings.Index(v, "name"), strings.Index(v, "A slash makes folders."))
	require.Contains(t, v, "kv/db", "the preview follows what is typed")

	bare := typeInto(newPrompt().WithLines("This cannot be undone."), "x")
	v = promptView(bare)
	require.Less(t, strings.Index(v, "This cannot be undone."), strings.Index(v, "> x"))
}

func TestPromptTakesAPaste(t *testing.T) {
	o, _ := newPrompt().Update(tea.PasteMsg{Content: "team/api"})
	require.Equal(t, "team/api", o.(Prompt).Value())
}

// Esc belongs to the stack, which closes the prompt without accepting.
func TestPromptEscClosesWithoutAccepting(t *testing.T) {
	s := Stack{}.Push(typeInto(newPrompt(), "db"))
	_, _, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.Equal(t, tea.Msg(CloseMsg{}), cmd())
}

func TestPromptLegendNamesTheAcceptHint(t *testing.T) {
	keys := newPrompt().WithAcceptHint("open the editor").Legend()
	line := ansi.Strip(legend.HintLine(keys, legend.Style{}))
	require.Contains(t, line, "[enter] open the editor")
	require.Contains(t, line, "[esc] cancel")
}
