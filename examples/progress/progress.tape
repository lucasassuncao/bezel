# progress: a count that fills on its own, and a stepper walked forward and back.
#
# Record from the repository root: vhs examples/progress/progress.tape

Output examples/progress/progress.gif

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
Type "go build -o progress.exe ./examples/progress"
Enter
Wait
Type "clear"
Enter
Show

Type ".\progress.exe"
Sleep 400ms
Enter
Wait+Screen /LINE/
Wait+Screen /40\/40/
Sleep 1.5s

# -- The stepper, forward to the end and back ---------------------------------------------

Type@900ms "nnnn"
Wait+Screen /All 4 steps done/
Sleep 1.5s
Type@700ms "pp"
Sleep 1.5s

# -- The count again --------------------------------------------------------------------------

Type "r"
Wait+Screen /40\/40/
Sleep 1.5s

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item progress.exe"
Enter
Wait
