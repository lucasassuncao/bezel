# table: rows scroll under the header with a scrollbar; ←/→ move between cells, ctrl+s sorts by a column, / filters, space checks rows for a batch, b turns a spinner on a row, w narrows the window.
#
# Record from the repository root: vhs docs/recordings/component/table.tape

Output docs/recordings/component/table.gif

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
Type "go build -o table.exe ./examples/component/table"
Enter
Wait
Type "clear"
Enter
Type ".\table.exe"
Enter
Wait+Screen /DESCRIPTION/
Show
Sleep 2s

Down@250ms 6
Sleep 800ms
PageDown@600ms 2
Sleep 1.5s
Enter
Sleep 1.5s
PageUp@600ms 3
Sleep 1s

Ctrl+S
Sleep 1.5s
Ctrl+S
Sleep 1.5s
Ctrl+S
Sleep 2s

Right@600ms 2
Sleep 1.5s
Left@400ms 2
Sleep 1s

Type "/"
Sleep 600ms
Type@250ms "x/s"
Sleep 1.2s
Down
Sleep 800ms
Enter
Sleep 1.5s
Space
Sleep 600ms
Up
Sleep 400ms
Space
Sleep 2s
Type "b"
Sleep 2.5s
Type "b"
Sleep 1s

Type "w"
Sleep 2.5s
Type "w"
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item table.exe"
Enter
Wait
