# chart: a load fed a sample per tick, paused and started over, beside bars, heat and a donut.
#
# Record from the repository root: vhs docs/recordings/component/chart.tape

Output docs/recordings/component/chart.gif

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
Type "go build -o chart.exe ./examples/component/chart"
Enter
Wait
Type "clear"
Enter
Type ".\chart.exe"
Enter
Wait+Screen /AREA/
Show
Sleep 4s

# -- Pause and start over -----------------------------------------------------------

Type " "
Wait+Screen /paused/
Sleep 1.5s
Type " "
Sleep 1s
Type "r"
Sleep 3s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item chart.exe"
Enter
Wait
