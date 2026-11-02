package shell

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lucasassuncao/bezel/internal/cells"
	"github.com/lucasassuncao/bezel/internal/frames"
)

// fadeIn is how long the backdrop takes to settle behind a modal that opens.
const fadeIn = 150 * time.Millisecond

// opened is s after an Update that found before modals open: one opening on
// an empty stack starts the backdrop's fade, and the command brings its frames.
func (s Shell) opened(before int) (Shell, tea.Cmd) {
	switch {
	case s.overlays.Len() == 0:
		s.fade, s.fadeP = frames.Run{}, 0
		return s, nil
	case before > 0:
		return s, nil
	}
	s.fade, s.fadeP = frames.NewRun(fadeIn), 0
	return s, s.fade.Tick()
}

// fadeFrame takes the backdrop's own frame: s after it, whether it was the
// backdrop's, and the command for the next one.
func (s Shell) fadeFrame(msg tea.Msg) (Shell, bool, tea.Cmd) {
	run, p, ok, cmd := s.fade.Update(msg)
	if !ok {
		return s, false, nil
	}
	s.fade, s.fadeP = run, p
	return s, true, cmd
}

// backdrop is screen blended toward the theme's background by its Backdrop,
// as far as the fade has come; untouched without a background to blend to.
func (s Shell) backdrop(screen string) string {
	th := &s.cfg.Theme
	f := th.Backdrop
	if s.fade.Active() {
		f *= frames.InOut(s.fadeP)
	}
	if f <= 0 || th.Background == nil {
		return screen
	}
	return cells.Parse(screen).Toward(th.Background, th.Foreground, f).String()
}
