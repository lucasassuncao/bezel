# themebrowser: the table a --list-themes flag prints, scrolled.
#
# Record from the repository root: vhs examples/themebrowser/themebrowser.tape

Output examples/themebrowser/themebrowser.gif

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
Type "go build -o themebrowser.exe ./examples/themebrowser"
Enter
Wait
Type "clear"
Enter
Show

Type ".\themebrowser.exe"
Sleep 400ms
Enter
Wait+Screen /Fruit/
Sleep 1.5s
Down@250ms 8
Sleep 1.5s
Type "q"
Sleep 1.5s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item themebrowser.exe"
Enter
Wait
