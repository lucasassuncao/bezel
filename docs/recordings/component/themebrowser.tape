# themebrowser: the table a --list-themes flag prints, scrolled, copied from and filtered.
#
# Record from the repository root: vhs docs/recordings/component/themebrowser.tape

Output docs/recordings/component/themebrowser.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        520
Set Theme         {"name":"Catppuccin Mocha on Mint","black":"#45475a","red":"#f38ba8","green":"#a6e3a1","yellow":"#f9e2af","blue":"#89b4fa","magenta":"#f5c2e7","cyan":"#94e2d5","white":"#bac2de","brightBlack":"#585b70","brightRed":"#f38ba8","brightGreen":"#a6e3a1","brightYellow":"#f9e2af","brightBlue":"#89b4fa","brightMagenta":"#f5c2e7","brightCyan":"#94e2d5","brightWhite":"#a6adc8","background":"#0B1B15","foreground":"#cdd6f4","selection":"#585b70","cursor":"#f5e0dc"}
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
Type "go build -o themebrowser.exe ./examples/component/themebrowser"
Enter
Wait
# clear runs first, so the command line is gone before the camera rolls.
Type "clear; .\themebrowser.exe"
Enter
Wait+Screen /Miscellaneous/
Show

Sleep 1.5s
Down@400ms 3
Sleep 1s
Enter
Sleep 1.5s
Type "/"
Sleep 500ms
Type "ter"
Sleep 1.5s
Enter
Sleep 2s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item themebrowser.exe"
Enter
Wait
