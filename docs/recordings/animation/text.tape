# animation, text: the text tab: words that wave, turn, land, blink and rise
#
# Record from the repository root: vhs docs/recordings/animation/text.tape

Output docs/recordings/animation/text.gif

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
Type "go build -o anim-text.exe ./examples/animation"
Enter
Wait
Type "clear"
Enter
Type ".\anim-text.exe"
Enter
Wait+Screen /every animation/
Tab@600ms 4
Wait+Screen /ellipsis/
Sleep 1s
Show

Sleep 2.5s
Type "d"
Sleep 2s
Type "t"
Sleep 1.3s
Type "t"
Sleep 1.3s
Type "l"
Sleep 1.3s
Type "k"
Sleep 1.2s
Type "c"
Sleep 1.3s
Type "t"
Sleep 1.5s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item anim-text.exe"
Enter
Wait
