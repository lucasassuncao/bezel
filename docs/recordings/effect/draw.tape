# effect, draw: the draw tab: effects over a card that is already drawn
#
# Record from the repository root: vhs docs/recordings/effect/draw.tape

Output docs/recordings/effect/draw.gif

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
Type "go build -o effect.exe ./examples/effect"
Enter
Wait
Type "clear"
Enter
Type ".\effect.exe"
Enter
Wait+Screen /every effect/
Wait+Screen /pattern/
Sleep 1.5s
Show
Sleep 1s

Type "r"
Sleep 1.5s
Type "e"
Sleep 2.6s
Type "e"
Sleep 2.6s
Type "e"
Sleep 1.5s
Type "e"
Sleep 1.5s
Type "e"
Sleep 1.5s
Type "e"
Sleep 2.6s
Type "e"
Sleep 1.5s
Type "e"
Sleep 2.6s
Type "e"
Sleep 1.5s
Type "e"
Sleep 1.5s
Type "e"
Sleep 2.6s
Type "e"
Sleep 2.6s

Type "e"
Sleep 1.5s
Type "p"
Sleep 1.5s
Type "p"
Sleep 1.5s
Type "p"
Sleep 1.5s
Type "p"
Sleep 1.5s
Type "p"
Sleep 1.5s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item effect.exe"
Enter
Wait
