# textbox: a multi-line value with numbered lines; ctrl+s counts them.
#
# Record from the repository root: vhs examples/textbox/textbox.tape

Output examples/textbox/textbox.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      14
Set Width         1200
Set Height        700
Set Theme         "Catppuccin Mocha"
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
Type "go build -o textbox.exe ./examples/textbox"
Enter
Wait
Type "clear"
Enter
Show

Type ".\textbox.exe"
Sleep 400ms
Enter
Wait+Screen /doubled/
Sleep 1.5s

Type "Every key is text here,"
Enter
Type "q included."
Sleep 1s

Ctrl+S
Wait+Screen /lines,/
Sleep 2.5s

Escape
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item textbox.exe"
Enter
Wait
