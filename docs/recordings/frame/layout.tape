# layout: a header over two columns; w narrows the window and the side column collapses.
#
# Record from the repository root: vhs docs/recordings/frame/layout.tape

Output docs/recordings/frame/layout.gif

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

# -- Build, off camera ----------------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o layout.exe ./examples/frame/layout"
Enter
Wait
Type "clear"
Enter
Type ".\layout.exe"
Enter
Wait+Screen /side/
Show
Sleep 2s

Type "w"
Sleep 2.5s
Type "w"
Sleep 2s
Type "w"
Sleep 2s
Type "w"
Sleep 1.5s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item layout.exe"
Enter
Wait
