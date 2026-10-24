# draw: a section per helper, scrolled through.
#
# Record from the repository root: vhs docs/recordings/frame/draw.tape

Output docs/recordings/frame/draw.gif

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
Type "go build -o draw.exe ./examples/frame/draw"
Enter
Wait
Type "clear"
Enter
Type ".\draw.exe"
Enter
Wait+Screen /HEADER/
Show
Sleep 2s

Down@250ms 16
Sleep 1.5s
Down@250ms 14
Sleep 1.5s
Down@250ms 12
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item draw.exe"
Enter
Wait
