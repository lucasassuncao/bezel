# inline: a CLI that logs, asks one question below the logs, and draws a bar.
#
# Record from the repository root: vhs examples/inline/inline.tape

Output examples/inline/inline.gif

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
Type "go build -o inline.exe ./examples/inline"
Enter
Wait
Type "clear"
Enter
Show

Type ".\inline.exe"
Sleep 400ms
Enter
Wait+Screen /Download them now/
Sleep 1.5s
Type "y"
Wait+Screen /done/
Sleep 2.5s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item inline.exe"
Enter
Wait
