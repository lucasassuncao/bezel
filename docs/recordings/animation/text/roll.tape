# text, roll: the roll tab: shimmer, scramble, an odometer and a gradient
#
# Record from the repository root: vhs docs/recordings/animation/text/roll.tape

Output docs/recordings/animation/text/roll.gif

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
Type "go build -o text-roll.exe ./examples/animation/text"
Enter
Wait
Type "clear"
Enter
Type ".\text-roll.exe"
Enter
Wait+Screen /every text/
Tab
Wait+Screen /shimmer/
Sleep 600ms
Show
Sleep 1.5s

Type "x"
Sleep 1.5s
Type "o"
Sleep 1.5s
Type "o"
Sleep 1.5s
Type "x"
Sleep 1.5s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item text-roll.exe"
Enter
Wait
