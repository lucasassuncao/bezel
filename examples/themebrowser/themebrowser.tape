# themebrowser: the table a --list-themes flag prints, scrolled.
#
# Record from the repository root: vhs examples/themebrowser/themebrowser.tape

Output examples/themebrowser/themebrowser.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        520
Set Theme         "Catppuccin Mocha"
Set Framerate     30
Set PlaybackSpeed 1.0
Set TypingSpeed   45ms
Set Padding       20

# -- Build and start, off camera -------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o themebrowser.exe ./examples/themebrowser"
Enter
Wait
# clear runs first, so the command line is gone before the camera rolls.
Type "clear; .\themebrowser.exe"
Enter
Wait+Screen /Fruit/
Show

Sleep 1.5s
Down@250ms 8
Sleep 1.5s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item themebrowser.exe"
Enter
Wait
