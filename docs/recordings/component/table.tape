# table: rows scroll under the header with a scrollbar; w narrows and the description gives way first.
#
# Record from the repository root: vhs docs/recordings/component/table.tape

Output docs/recordings/component/table.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        640
Set Theme         {"name":"Catppuccin Mocha on Mint","black":"#45475a","red":"#f38ba8","green":"#a6e3a1","yellow":"#f9e2af","blue":"#89b4fa","magenta":"#f5c2e7","cyan":"#94e2d5","white":"#bac2de","brightBlack":"#585b70","brightRed":"#f38ba8","brightGreen":"#a6e3a1","brightYellow":"#f9e2af","brightBlue":"#89b4fa","brightMagenta":"#f5c2e7","brightCyan":"#94e2d5","brightWhite":"#a6adc8","background":"#0B1B15","foreground":"#cdd6f4","selection":"#585b70","cursor":"#f5e0dc"}
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
Type "go build -o table.exe ./examples/component/table"
Enter
Wait
Type "clear"
Enter
Type ".\table.exe"
Enter
Wait+Screen /DESCRIPTION/
Show
Sleep 2s

Down@250ms 6
Sleep 800ms
PageDown@600ms 2
Sleep 1.5s
Enter
Sleep 1.5s
PageUp@600ms 3
Sleep 1s

Type "w"
Sleep 2.5s
Type "w"
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item table.exe"
Enter
Wait
