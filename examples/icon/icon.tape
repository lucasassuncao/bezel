# icon: every state mark in both sets, and a list swapped between them.
#
# Record from the repository root: vhs examples/icon/icon.tape

Output examples/icon/icon.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        900
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
Type "go build -o icon.exe ./examples/icon"
Enter
Wait
Type "clear"
Enter
Type ".\icon.exe"
Enter
Wait+Screen /STATES/
Show
Sleep 3s

# -- The list in ASCII, then back ------------------------------------------------------------

Type "a"
Wait+Screen /ASCII/
Sleep 2.5s
Type "a"
Wait+Screen /UNICODE/
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item icon.exe"
Enter
Wait
