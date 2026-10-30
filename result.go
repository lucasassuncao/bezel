package bezel

import tea "charm.land/bubbletea/v2"

// Result is what every Update reports besides the new model. The zero value
// means the message was not the model's.
type Result struct {
	// Handled is true when the message was this model's to take.
	Handled bool
	// Event says what happened, with its data; nil when nothing did.
	Event any
	// Cmd is the effect to run.
	Cmd tea.Cmd
}

// Model is the one shape of Update: the changed copy and a [Result].
type Model[T any] interface {
	Update(tea.Msg) (T, Result)
}

// ClickMsg is a left click the shell sends to the app, already in the
// coordinates of the pane's body. Pane names the pane that was hit.
type ClickMsg struct {
	Pane string
	tea.MouseClickMsg
}
