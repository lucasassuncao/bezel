# shell: two tabs, two panes, a status that expires and a busy spinner.
#
# Record from the repository root: vhs docs/recordings/frame/shell.tape

Output docs/recordings/frame/shell.gif

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
Type "go build -o shell.exe ./examples/frame/shell"
Enter
Wait
Type "clear"
Enter
Type ".\shell.exe"
Enter
Wait+Screen /main.go/
Show
Sleep 1.5s

Down@300ms 2
Sleep 800ms

# -- A status that clears itself ----------------------------------------------------

Type "s"
Wait+Screen /saved/
Sleep 2.5s

# -- Busy, then idle -------------------------------------------------------------------

Type "b"
Wait+Screen /indexing/
Sleep 2.5s

# -- Tabs and panes --------------------------------------------------------------------

Tab
Wait+Screen /ideas/
Sleep 1.5s
Shift+Tab
Sleep 1.2s

# -- Every key, one keystroke away ---------------------------------------------------

Type "?"
Sleep 2.5s
Escape
Sleep 1s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item shell.exe"
Enter
Wait
