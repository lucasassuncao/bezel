# tree: a directory tree: expand, collapse, a spinner on a busy node (space), and :reveal down to a path.
#
# Record from the repository root: vhs docs/recordings/component/tree.tape

Output docs/recordings/component/tree.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        640
Set Theme         {"name":"Campbell","black":"#0C0C0C","red":"#C50F1F","green":"#13A10E","yellow":"#C19C00","blue":"#0037DA","magenta":"#881798","cyan":"#3A96DD","white":"#CCCCCC","brightBlack":"#767676","brightRed":"#E74856","brightGreen":"#16C60C","brightYellow":"#F9F1A5","brightBlue":"#3B78FF","brightMagenta":"#B4009E","brightCyan":"#61D6D6","brightWhite":"#F2F2F2","background":"#0C0C0C","foreground":"#CCCCCC","selection":"#767676","cursor":"#FFFFFF"}
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
Type "go build -o tree.exe ./examples/component/tree"
Enter
Wait
Type "clear"
Enter
Type ".\tree.exe"
Enter
Wait+Screen /internal/
Show
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

Space
Sleep 2.5s
Space
Sleep 1s

Type ":reveal internal/tui/keys/keys.go"
Sleep 800ms
Enter
Wait+Screen /keys.go/
Sleep 2.5s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item tree.exe"
Enter
Wait
