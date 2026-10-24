# accordion: sections that open and close, sliding
#
# Record from the repository root: vhs docs/recordings/component/accordion.tape

Output docs/recordings/component/accordion.gif

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
Type "go build -o accordion.exe ./examples/component/accordion"
Enter
Wait
Type "clear"
Enter
Type ".\accordion.exe"
Enter
Wait+Screen /What is bezel/
Show
Sleep 1.2s

Enter
Sleep 1.5s
Down@400ms 2
Space
Sleep 1.5s
Up@400ms 2
Enter
Sleep 1.2s
Down@400ms 3
Enter
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item accordion.exe"
Enter
Wait
