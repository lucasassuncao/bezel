# stopwatch: time counted up; space starts and stops it, r puts it back to zero.
#
# Record from the repository root: vhs docs/recordings/component/stopwatch.tape

Output docs/recordings/component/stopwatch.gif

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
Type "go build -o stopwatch.exe ./examples/component/stopwatch"
Enter
Wait
Type "clear"
Enter
Type ".\stopwatch.exe"
Enter
Wait+Screen /ELAPSED/
Show
Sleep 1s

# -- Running, held, back to zero and running again --------------------------------

Space
Sleep 3s
Space
Sleep 1.5s
Type "r"
Sleep 1s
Space
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item stopwatch.exe"
Enter
Wait
