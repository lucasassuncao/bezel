# animation, curves: the curves tab: every easing curve, raced out, in and in-out
#
# Record from the repository root: vhs docs/recordings/animation/curves.tape

Output docs/recordings/animation/curves.gif

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

# -- Build and start, off camera -------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o animation.exe ./examples/animation"
Enter
Wait
Type "clear"
Enter
Type ".\animation.exe"
Enter
Wait+Screen /every animation/
Tab@600ms 1
Wait+Screen /steps 6/
Sleep 600ms
Show
Sleep 1s

Type "r"
Sleep 2s
Type "i"
Sleep 2s
Type "i"
Sleep 2s
Type "i"
Sleep 2s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item animation.exe"
Enter
Wait
