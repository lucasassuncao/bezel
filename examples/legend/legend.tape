# legend: more keys than one row holds, counted; w lets the write keys in; ? lists them all.
#
# Record from the repository root: vhs examples/legend/legend.tape

Output examples/legend/legend.gif

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
Type "go build -o legend.exe ./examples/legend"
Enter
Wait
Type "clear"
Enter
Show

Type ".\legend.exe"
Sleep 400ms
Enter
Wait+Screen /write mode/
Sleep 2.5s

Type "?"
Sleep 2.5s
Escape
Sleep 800ms

Type "w"
Wait+Screen /every key is shown/
Sleep 2s

Type "?"
Sleep 3s
Escape
Sleep 800ms

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item legend.exe"
Enter
Wait
