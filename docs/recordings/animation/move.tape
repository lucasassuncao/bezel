# animation, move: the move tab: points that ease and spring, and a marker on a path
#
# Record from the repository root: vhs docs/recordings/animation/move.tape

Output docs/recordings/animation/move.gif

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

# -- Build and start, off camera -------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o anim-move.exe ./examples/animation"
Enter
Wait
Type "clear"
Enter
Type ".\anim-move.exe"
Enter
Wait+Screen /every animation/
Tab@600ms 2
Wait+Screen /sprung/
Sleep 1s
Show

Sleep 1s
Type "m"
Sleep 1.5s
Type "p"
Sleep 1.2s
Type "m"
Sleep 1.5s
Type "m"
Sleep 2s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item anim-move.exe"
Enter
Wait
