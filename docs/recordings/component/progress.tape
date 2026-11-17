# progress: a count that fills on its own, and a stepper walked forward and back.
#
# Record from the repository root: vhs docs/recordings/component/progress.tape

Output docs/recordings/component/progress.gif

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
Type "go build -o progress.exe ./examples/component/progress"
Enter
Wait
Type "clear"
Enter
Type ".\progress.exe"
Enter
Wait+Screen /LINE/
Show
Wait+Screen /40\/40/
Sleep 1.5s

# -- The stepper, forward to the end and back ---------------------------------------------

Type@900ms "nnnn"
Wait+Screen /All 4 steps done/
Sleep 1.5s
Type@700ms "pp"
Sleep 1.5s

# -- The count again --------------------------------------------------------------------------

Type "r"
Wait+Screen /40\/40/
Sleep 1.5s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item progress.exe"
Enter
Wait
