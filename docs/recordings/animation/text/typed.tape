# text, typed: the typed tab: a headline typed out, a marquee and a shake
#
# Record from the repository root: vhs docs/recordings/animation/text/typed.tape

Output docs/recordings/animation/text/typed.gif

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
Type "go build -o text-typed.exe ./examples/animation/text"
Enter
Wait
Type "clear"
Enter
Type ".\text-typed.exe"
Enter
Wait+Screen /every text/
Wait+Screen /marquee/
Sleep 600ms
Show

Sleep 1.5s
Type "t"
Sleep 2.2s
Type "s"
Sleep 1.2s
Type "s"
Sleep 1.5s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item text-typed.exe"
Enter
Wait
