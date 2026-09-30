# palette: the : command line: completion, arguments, and the errors it gives.
#
# Record from the repository root: vhs examples/palette/palette.tape

Output examples/palette/palette.gif

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
Type "go build -o palette.exe ./examples/palette"
Enter
Wait
Type "clear"
Enter
Show

Type ".\palette.exe"
Sleep 400ms
Enter
Wait+Screen /apple/
Sleep 1.5s

# -- Complete a name, pass an argument -----------------------------------------------

Type ":"
Sleep 1s
Type "go"
Sleep 1s
Tab
Sleep 800ms
Type "mango"
Sleep 800ms
Enter
Sleep 2s

Type ":theme strawberry"
Sleep 800ms
Enter
Sleep 2s

# -- What it cannot run ------------------------------------------------------------------

Type ":xyz"
Sleep 600ms
Enter
Wait+Screen /unknown command/
Sleep 2s

Type ":delete"
Sleep 600ms
Enter
Wait+Screen /not available/
Sleep 2s

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item palette.exe"
Enter
Wait
