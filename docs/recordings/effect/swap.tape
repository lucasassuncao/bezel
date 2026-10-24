# effect, swap: the swap tab: one card carried into another
#
# Record from the repository root: vhs docs/recordings/effect/swap.tape

Output docs/recordings/effect/swap.gif

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
Type "go build -o effect-swap.exe ./examples/effect"
Enter
Wait
Type "clear"
Enter
Type ".\effect-swap.exe"
Enter
Wait+Screen /every effect/
Tab@600ms 2
Wait+Screen /transition/
Sleep 1s
Show

Sleep 1s
Type "r"
Sleep 1.4s
Type "n"
Sleep 1.4s
Type "n"
Sleep 1.4s
Type "n"
Sleep 1.4s
Type "n"
Sleep 1.4s
Type "n"
Sleep 1.4s
Type "n"
Sleep 1.4s
Type "n"
Sleep 1.4s
Type "r"
Sleep 1.6s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item effect-swap.exe"
Enter
Wait
