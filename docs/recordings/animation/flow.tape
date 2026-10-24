# animation, flow: the flow tab: a timeline, keyframes and a path
#
# Record from the repository root: vhs docs/recordings/animation/flow.tape

Output docs/recordings/animation/flow.gif

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
Type "go build -o anim-flow.exe ./examples/animation"
Enter
Wait
Type "clear"
Enter
Type ".\anim-flow.exe"
Enter
Wait+Screen /every animation/
Tab@600ms 5
Wait+Screen /keyframes/
Sleep 1s
Show

Sleep 2.5s
Type "r"
Sleep 2.8s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item anim-flow.exe"
Enter
Wait
