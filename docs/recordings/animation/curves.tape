# animation, curves: the curves tab: every easing curve, raced out, in and in-out
#
# Record from the repository root: vhs docs/recordings/animation/curves.tape

Output docs/recordings/animation/curves.gif

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
Tab@600ms 2
Wait+Screen /steps 6/
Sleep 600ms
Show
Sleep 1s

Type "r"
Sleep 2s
Type "i"
Sleep 2s
Type "i"
Sleep 2s
Type "i"
Sleep 2s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item animation.exe"
Enter
Wait
