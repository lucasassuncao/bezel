# animation: h slides a hints pane open and shut.
#
# Record from the repository root: vhs examples/animation/animation.tape

Output examples/animation/animation.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      14
Set Width         1200
Set Height        700
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
Type "go build -o animation.exe ./examples/animation"
Enter
Wait
Type "clear"
Enter
Show

Type ".\animation.exe"
Sleep 400ms
Enter
Wait+Screen /press h/
Sleep 1.5s

Type "h"
Sleep 1.5s
Type "h"
Sleep 1.5s
Type "h"
Sleep 1.5s
Type "h"
Sleep 1.5s

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item animation.exe"
Enter
Wait
