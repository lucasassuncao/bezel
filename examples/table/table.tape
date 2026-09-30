# table: columns sized to their contents; w narrows and the description gives way first.
#
# Record from the repository root: vhs examples/table/table.tape

Output examples/table/table.gif

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
Type "go build -o table.exe ./examples/table"
Enter
Wait
Type "clear"
Enter
Show

Type ".\table.exe"
Sleep 400ms
Enter
Wait+Screen /DESCRIPTION/
Sleep 2s

Type "w"
Sleep 2.5s
Type "w"
Sleep 2s

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item table.exe"
Enter
Wait
