# textbox: a multi-line value with numbered lines; ctrl+s counts them.
#
# Record from the repository root: vhs docs/recordings/component/textbox.tape

Output docs/recordings/component/textbox.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        640
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
Type "go build -o textbox.exe ./examples/component/textbox"
Enter
Wait
Type "clear"
Enter
Type ".\textbox.exe"
Enter
Wait+Screen /doubled/
Show
Sleep 1.5s

Type "Every key is text here,"
Enter
Type "q included."
Sleep 1s

Ctrl+S
Wait+Screen /lines,/
Sleep 2.5s

Hide
Escape
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item textbox.exe"
Enter
Wait
