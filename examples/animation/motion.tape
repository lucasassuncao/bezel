# animation, motion: the motion tab: a drop, a pendulum, an orbit, a flock and a rubber band
#
# Record from the repository root: vhs examples/animation/motion.tape

Output examples/animation/motion.gif

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
Type "go build -o anim-motion.exe ./examples/animation"
Enter
Wait
Type "clear"
Enter
Type ".\anim-motion.exe"
Enter
Wait+Screen /every animation/
Tab@600ms 8
Wait+Screen /rubber/
Sleep 1s
Show

Sleep 1s
Type "d"
Sleep 3s
Type "f"
Sleep 1.2s
Type "f"
Sleep 1.2s
Type "f"
Sleep 1.5s
Type "b"
Sleep 1.2s
Type "b"
Sleep 1.5s
Type "d"
Sleep 2.5s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item anim-motion.exe"
Enter
Wait
