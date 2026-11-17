# breadcrumb: a store browsed by path; enter goes into a folder, tab moves to the path, and the arrows and enter go back up.
#
# Record from the repository root: vhs docs/recordings/component/breadcrumb.tape

Output docs/recordings/component/breadcrumb.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        520
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
Type "go build -o breadcrumb.exe ./examples/component/breadcrumb"
Enter
Wait
Type "clear"
Enter
Type ".\breadcrumb.exe"
Enter
Wait+Screen /infra/
Show
Sleep 1.5s

# -- Down into the folders -------------------------------------------------------------

Enter
Sleep 900ms
Enter
Sleep 900ms
Enter
Sleep 900ms
Enter
Sleep 1.2s
Enter
Sleep 2s

# -- Back up through the path ----------------------------------------------------------

Tab
Sleep 800ms
Left@600ms 3
Sleep 800ms
Enter
Sleep 1.5s
Down
Sleep 500ms
Enter
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item breadcrumb.exe"
Enter
Wait
