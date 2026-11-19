# effect, static: one item on its own, from ./examples/effect --effect static
#
# Record from the repository root: vhs docs/recordings/effect/solo/static.tape

Output docs/recordings/effect/solo/static.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         480
Set Height        355
Set Theme         {"name":"Campbell","black":"#0C0C0C","red":"#C50F1F","green":"#13A10E","yellow":"#C19C00","blue":"#0037DA","magenta":"#881798","cyan":"#3A96DD","white":"#CCCCCC","brightBlack":"#767676","brightRed":"#E74856","brightGreen":"#16C60C","brightYellow":"#F9F1A5","brightBlue":"#3B78FF","brightMagenta":"#B4009E","brightCyan":"#61D6D6","brightWhite":"#F2F2F2","background":"#0C0C0C","foreground":"#CCCCCC","selection":"#767676","cursor":"#FFFFFF"}
Set Framerate     15
Set PlaybackSpeed 1.0
Set TypingSpeed   45ms
Set Padding       20

# -- Build and start, off camera -------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o effect-solo-static.exe ./examples/effect"
Enter
Wait
# Slow motion: the effect runs at 0.45 of its speed, so each phase can be seen.
Type "$env:BEZEL_MOTION='0.45'"
Enter
Type "clear"
Enter
Type ".\effect-solo-static.exe --effect static"
Enter
Wait+Screen /static/
Sleep 600ms
Show

Sleep 1s
Type "r"
Sleep 2.8s
Type "r"
Sleep 2.8s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item effect-solo-static.exe"
Enter
Wait
