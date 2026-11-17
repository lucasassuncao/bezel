# logger: a run logged as a tree under each message, wrapped at the width.
#
# Record from the repository root: vhs docs/recordings/component/logger.tape

Output docs/recordings/component/logger.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        900
Set Theme         {"name":"Campbell","black":"#0C0C0C","red":"#C50F1F","green":"#13A10E","yellow":"#C19C00","blue":"#0037DA","magenta":"#881798","cyan":"#3A96DD","white":"#CCCCCC","brightBlack":"#767676","brightRed":"#E74856","brightGreen":"#16C60C","brightYellow":"#F9F1A5","brightBlue":"#3B78FF","brightMagenta":"#B4009E","brightCyan":"#61D6D6","brightWhite":"#F2F2F2","background":"#0C0C0C","foreground":"#CCCCCC","selection":"#767676","cursor":"#FFFFFF"}
Set Framerate     30
Set PlaybackSpeed 1.0
Set TypingSpeed   45ms
Set Padding       20

# -- Build and start, off camera -------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o logger.exe ./examples/component/logger"
Enter
Wait
# clear runs first, so the command line is gone before the camera rolls, and
# the sleep after holds the last frame without a prompt under it.
Type "clear; .\logger.exe; Start-Sleep 60"
Enter
Wait+Screen /No .jpg/
Show

Wait+Screen /batch_id/
Sleep 3s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Ctrl+C
Sleep 500ms
Type "Remove-Item logger.exe"
Enter
Wait
