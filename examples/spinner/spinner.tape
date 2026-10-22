# spinner: the twelve presets turning; space stops them all and starts them again.
#
# Record from the repository root: vhs examples/spinner/spinner.tape

Output examples/spinner/spinner.gif

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
Type "go build -o spinner.exe ./examples/spinner"
Enter
Wait
Type "clear"
Enter
Type ".\spinner.exe"
Enter
Wait+Screen /Hamburger/
Show
Sleep 4s

# -- Stopped, then turning again --------------------------------------------------

Space
Sleep 1.5s
Space
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item spinner.exe"
Enter
Wait
