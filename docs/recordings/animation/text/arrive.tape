# text, arrive: the arrive tab: text that streams, morphs, is typed, flips, glows, burns and falls
#
# Record from the repository root: vhs docs/recordings/animation/text/arrive.tape

Output docs/recordings/animation/text/arrive.gif

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
Type "go build -o text-arrive.exe ./examples/animation/text"
Enter
Wait
Type "clear"
Enter
Type ".\text-arrive.exe"
Enter
Wait+Screen /every text/
Tab@600ms 3
Wait+Screen /stream/
Sleep 1s
Show

Sleep 1.5s
Type "s"
Sleep 2s
Type "m"
Sleep 1.2s
Type "y"
Sleep 2.5s
Type "f"
Sleep 1.2s
Type "f"
Sleep 1.2s
Type "b"
Sleep 2s
Type "a"
Sleep 2s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item text-arrive.exe"
Enter
Wait
