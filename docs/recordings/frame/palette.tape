# palette: the : command line: completion, letters matched in order, arguments, and the errors it gives.
#
# Record from the repository root: vhs docs/recordings/frame/palette.tape

Output docs/recordings/frame/palette.gif

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
Type "go build -o palette.exe ./examples/frame/palette"
Enter
Wait
Type "clear"
Enter
Type ".\palette.exe"
Enter
Wait+Screen /apple/
Show
Sleep 1.5s

# -- Complete a name, pass an argument -----------------------------------------------

Type ":"
Sleep 1s
Type "go"
Sleep 1s
Tab
Sleep 800ms
Type "mango"
Sleep 800ms
Enter
Sleep 2s

Type ":theme strawberry"
Sleep 800ms
Enter
Sleep 2s

# -- Letters in order, not only a prefix -----------------------------------------------

Type ":"
Sleep 600ms
Type@200ms "tme"
Sleep 2s
Tab
Sleep 600ms
Type "mint"
Sleep 800ms
Enter
Sleep 2s

# -- What it cannot run ------------------------------------------------------------------

Type ":xyz"
Sleep 600ms
Enter
Wait+Screen /unknown command/
Sleep 2s

Type ":delete"
Sleep 600ms
Enter
Wait+Screen /not available/
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item palette.exe"
Enter
Wait
