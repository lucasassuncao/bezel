# draw: a section per helper, scrolled through.
#
# Record from the repository root: vhs examples/draw/draw.tape

Output examples/draw/draw.gif

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
Type "go build -o draw.exe ./examples/draw"
Enter
Wait
Type "clear"
Enter
Show

Type ".\draw.exe"
Sleep 400ms
Enter
Wait+Screen /HEADER/
Sleep 2s

Down@250ms 12
Sleep 1.5s
Down@250ms 14
Sleep 2s
Up@150ms 26
Sleep 1.5s

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item draw.exe"
Enter
Wait
