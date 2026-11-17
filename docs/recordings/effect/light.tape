# effect, light: the light tab: lights and presets on a card, a diff and its undo
#
# Record from the repository root: vhs docs/recordings/effect/light.tape

Output docs/recordings/effect/light.gif

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
Type "go build -o effect-light.exe ./examples/effect"
Enter
Wait
Type "clear"
Enter
Type ".\effect-light.exe"
Enter
Wait+Screen /every effect/
Tab@600ms 1
Wait+Screen /ghost/
Sleep 1s
Show

Sleep 1.5s
Type "f"
Sleep 2s
Type "u"
Sleep 2s
Type "e"
Sleep 2.4s
Type "e"
Sleep 1s
Type "e"
Sleep 1.6s
Type "e"
Sleep 1s
Type "e"
Sleep 1.6s
Type "e"
Sleep 2.6s
Type "e"
Sleep 2.6s
Type "e"
Sleep 1s
Type "e"
Sleep 2.6s
Type "e"
Sleep 1s
Type "e"
Sleep 1s
Type "e"
Sleep 1.2s
Type "e"
Sleep 1s
Type "e"
Sleep 1s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item effect-light.exe"
Enter
Wait
