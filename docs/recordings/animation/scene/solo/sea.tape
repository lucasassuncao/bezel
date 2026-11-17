# animation/scene, sea: one item on its own, from ./examples/animation/scene --scene sea
#
# Record from the repository root: vhs docs/recordings/animation/scene/solo/sea.tape

Output docs/recordings/animation/scene/solo/sea.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         520
Set Height        376
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
Type "go build -o anim-scene-solo-sea.exe ./examples/animation/scene"
Enter
Wait
Type "clear"
Enter
Type ".\anim-scene-solo-sea.exe --scene sea"
Enter
Wait+Screen /sea/
Sleep 600ms
Show

Sleep 4s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item anim-scene-solo-sea.exe"
Enter
Wait
