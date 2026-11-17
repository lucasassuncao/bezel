# tags: chips that filter a list; enter adds a topic, tab completes it, backspace on an empty field takes the last off.
#
# Record from the repository root: vhs docs/recordings/component/tags.tape

Output docs/recordings/component/tags.gif

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
Type "go build -o tags.exe ./examples/component/tags"
Enter
Wait
Type "clear"
Enter
Type ".\tags.exe"
Enter
Wait+Screen /8 of 8/
Show
Sleep 1.5s

# -- Two topics, one completed with tab, then the last one off -------------------------

Type@150ms "tu"
Sleep 800ms
Tab
Sleep 600ms
Enter
Sleep 1.5s
Type@150ms "st"
Sleep 800ms
Tab
Sleep 600ms
Enter
Sleep 2s
Backspace
Sleep 1.5s
Type@150ms "framework"
Enter
Sleep 2s

Hide
Escape
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item tags.exe"
Enter
Wait
