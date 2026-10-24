# animation, basics: the basics tab: easings, springs, fades, typed text and the rest
#
# Record from the repository root: vhs docs/recordings/animation/basics.tape

Output docs/recordings/animation/basics.gif

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
Type "go build -o animation.exe ./examples/animation"
Enter
Wait
Type "clear"
Enter
Type ".\animation.exe"
Enter
Wait+Screen /every animation/
Show
Sleep 2.5s

Right@400ms 2
Sleep 1.5s
Left
Sleep 1.5s

Type "f"
Sleep 1.8s
Type "p"
Sleep 2.5s
Type "p"
Sleep 800ms

Type "h"
Sleep 1.2s
Type "h"
Sleep 1s

Type "r"
Sleep 2.5s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item animation.exe"
Enter
Wait
