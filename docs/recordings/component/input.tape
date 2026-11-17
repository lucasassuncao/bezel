# input: a name that must not be empty, a command tab completes, a masked password and a time typed into a format.
#
# Record from the repository root: vhs docs/recordings/component/input.tape

Output docs/recordings/component/input.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        520
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
Type "go build -o input.exe ./examples/component/input"
Enter
Wait
Type "clear"
Enter
Type ".\input.exe"
Enter
Wait+Screen /deploy/
Show
Sleep 1s

# -- The name, left empty first ------------------------------------------------------

Down
Wait+Screen /a name is required/
Sleep 1.5s
Up
Type "ana"
Sleep 800ms

# -- The command, completed with tab ---------------------------------------------------

Down
Type "ro"
Sleep 1s
Tab
Sleep 1s

# -- The password, behind its mask ----------------------------------------------------

Down
Type "s3cret"
Sleep 1s

# -- The time, laid on its format ------------------------------------------------------

Down
Sleep 600ms
Type@220ms "202611021430"
Sleep 1.5s
Enter
Wait+Screen /rollback/
Sleep 2.5s

Hide
Escape
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item input.exe"
Enter
Wait
