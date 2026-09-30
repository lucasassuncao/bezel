# list: rows under headings, a / filter, and enter showing the row's payload.
#
# Record from the repository root: vhs examples/list/list.tape

Output examples/list/list.gif

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
Type "go build -o list.exe ./examples/list"
Enter
Wait
Type "clear"
Enter
Show

Type ".\list.exe"
Sleep 400ms
Enter
Wait+Screen /Fruit/
Sleep 1.5s

# -- Across a section heading --------------------------------------------------------

Down@300ms 4
Sleep 800ms
Enter
Wait+Screen /kcal/
Sleep 2s

# -- The filter ------------------------------------------------------------------------

Type "/an"
Sleep 2s
Escape
Sleep 1.5s

Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item list.exe"
Enter
Wait
