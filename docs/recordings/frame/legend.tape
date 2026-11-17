# legend: more keys than one row holds, counted; w lets the write keys in; ? lists them all.
#
# Record from the repository root: vhs docs/recordings/frame/legend.tape

Output docs/recordings/frame/legend.gif

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
Type "go build -o legend.exe ./examples/frame/legend"
Enter
Wait
Type "clear"
Enter
Type ".\legend.exe"
Enter
Wait+Screen /write mode/
Show
Sleep 2.5s

Type "?"
Sleep 2.5s
Escape
Sleep 800ms

Type "w"
Wait+Screen /every key is shown/
Sleep 2s

Type "?"
Sleep 3s
Escape
Sleep 800ms

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item legend.exe"
Enter
Wait
