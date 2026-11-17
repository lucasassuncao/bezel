# logpane: a deploy logs through slog into two logger.Panes, bezel's Handler above and slog's JSON through Writer below; scrolling up holds a pane still and counts what arrives, paging back down follows again; / filters a pane to the records holding the text.
#
# Record from the repository root: vhs docs/recordings/component/logpane.tape

Output docs/recordings/component/logpane.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        760
Set Theme         {"name":"Campbell","black":"#0C0C0C","red":"#C50F1F","green":"#13A10E","yellow":"#C19C00","blue":"#0037DA","magenta":"#881798","cyan":"#3A96DD","white":"#CCCCCC","brightBlack":"#767676","brightRed":"#E74856","brightGreen":"#16C60C","brightYellow":"#F9F1A5","brightBlue":"#3B78FF","brightMagenta":"#B4009E","brightCyan":"#61D6D6","brightWhite":"#F2F2F2","background":"#0C0C0C","foreground":"#CCCCCC","selection":"#767676","cursor":"#FFFFFF"}
Set Framerate     10
Set PlaybackSpeed 1.0
Set TypingSpeed   45ms
Set Padding       20

# -- Build, off camera ----------------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o logpane.exe ./examples/component/logpane"
Enter
Wait
Type "clear"
Enter
Type ".\logpane.exe"
Enter
Wait+Screen /following/
Show
Sleep 3s

Up@300ms 8
Sleep 3s
PageDown@200ms 10
Sleep 2s

Type "/"
Sleep 500ms
Type@120ms "refused"
Sleep 2s
Enter
Sleep 3s

Tab
Sleep 500ms
Type "/"
Sleep 500ms
Type@120ms "warn"
Enter
Sleep 3s
Escape
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item logpane.exe"
Enter
Wait
