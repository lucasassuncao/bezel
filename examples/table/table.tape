# table: rows scroll under the header with a scrollbar; w narrows and the description gives way first.
#
# Record from the repository root: vhs examples/table/table.tape

Output examples/table/table.gif

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
Type "go build -o table.exe ./examples/table"
Enter
Wait
Type "clear"
Enter
Type ".\table.exe"
Enter
Wait+Screen /DESCRIPTION/
Show
Sleep 2s

Down@250ms 6
Sleep 800ms
PageDown@600ms 2
Sleep 1.5s
Enter
Sleep 1.5s
PageUp@600ms 3
Sleep 1s

Type "w"
Sleep 2.5s
Type "w"
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item table.exe"
Enter
Wait
