# calendar: a month with a cursor on a day; the arrows move a day or a week, pgup/pgdown a month, enter chooses.
#
# Record from the repository root: vhs docs/recordings/component/calendar.tape

Output docs/recordings/component/calendar.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        460
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
Type "go build -o calendar.exe ./examples/component/calendar"
Enter
Wait
Type "clear"
Enter
Type ".\calendar.exe"
Enter
Wait+Screen /November 2026/
Show
Sleep 1.5s

# -- A day, a week, a month, then choose ------------------------------------------------

Right@500ms 3
Sleep 600ms
Down@500ms 2
Sleep 800ms
PageDown
Sleep 1.2s
PageDown
Sleep 1.2s
Left@500ms 2
Sleep 800ms
Enter
Sleep 2s
PageUp@700ms 3
Sleep 1.5s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item calendar.exe"
Enter
Wait
