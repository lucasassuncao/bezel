# theme: every colour role and a list, stepped through a few presets.
#
# Record from the repository root: vhs examples/theme/theme.tape

Output examples/theme/theme.gif

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
Type "go build -o theme.exe ./examples/theme"
Enter
Wait
Type "clear"
Enter
Show

Type ".\theme.exe"
Sleep 400ms
Enter
Wait+Screen /Success/
Sleep 2s

Type "n"
Sleep 1.5s
Type "n"
Sleep 1.5s
Type "n"
Sleep 1.5s
Type "n"
Sleep 1.5s
Type "n"
Sleep 1.5s
Type "p"
Sleep 1.5s

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item theme.exe"
Enter
Wait
