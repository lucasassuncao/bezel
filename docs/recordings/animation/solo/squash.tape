# animation, squash: one item on its own, from ./examples/animation --animation squash
#
# Record from the repository root: vhs docs/recordings/animation/solo/squash.tape

Output docs/recordings/animation/solo/squash.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         500
Set Height        229
Set Theme         {"name":"Campbell","black":"#0C0C0C","red":"#C50F1F","green":"#13A10E","yellow":"#C19C00","blue":"#0037DA","magenta":"#881798","cyan":"#3A96DD","white":"#CCCCCC","brightBlack":"#767676","brightRed":"#E74856","brightGreen":"#16C60C","brightYellow":"#F9F1A5","brightBlue":"#3B78FF","brightMagenta":"#B4009E","brightCyan":"#61D6D6","brightWhite":"#F2F2F2","background":"#0C0C0C","foreground":"#CCCCCC","selection":"#767676","cursor":"#FFFFFF"}
Set Framerate     20
Set PlaybackSpeed 1.0
Set TypingSpeed   45ms
Set Padding       20

# -- Build and start, off camera -------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o anim-solo-squash.exe ./examples/animation"
Enter
Wait
Type "clear"
Enter
Type ".\anim-solo-squash.exe --animation squash"
Enter
Wait+Screen /squash/
Sleep 600ms
Show

Sleep 1s
Type "r"
Sleep 3s
Type "r"
Sleep 3s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item anim-solo-squash.exe"
Enter
Wait
