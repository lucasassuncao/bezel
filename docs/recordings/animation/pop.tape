# animation, pop: the pop tab: a burst, a donut, a race and a ball that squashes
#
# Record from the repository root: vhs docs/recordings/animation/pop.tape

Output docs/recordings/animation/pop.gif

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
Type "go build -o anim-pop.exe ./examples/animation"
Enter
Wait
Type "clear"
Enter
Type ".\anim-pop.exe"
Enter
Wait+Screen /every animation/
Tab@600ms 6
Wait+Screen /squash/
Sleep 1s
Show

Sleep 1s
Type "x"
Sleep 1.5s
Type "o"
Sleep 1.3s
Type "v"
Sleep 1.3s
Type "d"
Sleep 1.6s
Type "v"
Sleep 1.2s
Type "x"
Sleep 1.5s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item anim-pop.exe"
Enter
Wait
