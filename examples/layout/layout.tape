# layout: a header over two columns; w narrows the window and the side column collapses.
#
# Record from the repository root: vhs examples/layout/layout.tape

Output examples/layout/layout.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      14
Set Width         1200
Set Height        700
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
Type "go build -o layout.exe ./examples/layout"
Enter
Wait
Type "clear"
Enter
Show

Type ".\layout.exe"
Sleep 400ms
Enter
Wait+Screen /side/
Sleep 2s

Type "w"
Sleep 2.5s
Type "w"
Sleep 2s
Type "w"
Sleep 2s
Type "w"
Sleep 1.5s

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item layout.exe"
Enter
Wait
