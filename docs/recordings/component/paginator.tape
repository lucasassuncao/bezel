# paginator: 42 items, 8 to a page; the arrows turn the page, t switches to dots.
#
# Record from the repository root: vhs docs/recordings/component/paginator.tape

Output docs/recordings/component/paginator.gif

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
Type "go build -o paginator.exe ./examples/component/paginator"
Enter
Wait
Type "clear"
Enter
Type ".\paginator.exe"
Enter
Wait+Screen /1\/6/
Show
Sleep 1.5s

# -- Forward as numbers, then back as dots ------------------------------------------

Right@900ms 3
Sleep 1s
Type "t"
Sleep 1.5s
Left@900ms 2
Sleep 1s
Type "t"
Sleep 1.5s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item paginator.exe"
Enter
Wait
