# list: rows under headings, space checking rows for a batch, a spinner on a busy row (b toggles it), a / filter, and enter showing the row's payload.
#
# Record from the repository root: vhs docs/recordings/component/list.tape

Output docs/recordings/component/list.gif

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
Type "go build -o list.exe ./examples/component/list"
Enter
Wait
Type "clear"
Enter
Type ".\list.exe"
Enter
Wait+Screen /Fruit/
Show
Sleep 1.5s

# -- Across a section heading --------------------------------------------------------

Down@300ms 4
Sleep 800ms
Enter
Wait+Screen /kcal/
Sleep 2s

# -- Checking rows for a batch -------------------------------------------------------

Space
Sleep 600ms
Down
Sleep 400ms
Space
Sleep 600ms
Down
Sleep 400ms
Space
Sleep 2s

# -- A busy row ---------------------------------------------------------------------

Type "b"
Sleep 2.5s
Type "b"
Sleep 1s

# -- The filter ------------------------------------------------------------------------

Type "/an"
Sleep 2s
Escape
Sleep 1.5s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item list.exe"
Enter
Wait
