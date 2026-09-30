# tree: a directory tree: expand, collapse, and :reveal down to a path.
#
# Record from the repository root: vhs examples/tree/tree.tape

Output examples/tree/tree.gif

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
Type "go build -o tree.exe ./examples/tree"
Enter
Wait
Type "clear"
Enter
Show

Type ".\tree.exe"
Sleep 400ms
Enter
Wait+Screen /internal/
Sleep 1.5s

Right
Sleep 1s
Down
Sleep 600ms
Right
Sleep 1.2s
Left
Sleep 1s
Up
Sleep 600ms
Left
Sleep 1.2s

Type ":reveal internal/tui/keys/keys.go"
Sleep 800ms
Enter
Wait+Screen /keys.go/
Sleep 2.5s

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item tree.exe"
Enter
Wait
