# timer: ten seconds counted down; space stops and resumes, r starts over.
#
# Record from the repository root: vhs examples/timer/timer.tape

Output examples/timer/timer.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        640
Set Theme         "Catppuccin Mocha"
Set Framerate     30
Set PlaybackSpeed 1.0
Set TypingSpeed   45ms
Set Padding       20

# -- Build, off camera ----------------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o timer.exe ./examples/timer"
Enter
Wait
Type "clear"
Enter
Type ".\timer.exe"
Enter
Wait+Screen /COUNTDOWN/
Show
Sleep 3s

# -- Stopped a moment, then on to the end ------------------------------------------

Space
Sleep 1.5s
Space
Wait+Screen@15s /time's up/
Sleep 1.5s

# -- Over again -----------------------------------------------------------------------

Type "r"
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item timer.exe"
Enter
Wait
