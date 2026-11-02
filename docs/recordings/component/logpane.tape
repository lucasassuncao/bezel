# logpane: a deploy logs through slog into logger.Pane; scrolling up holds it still and counts what arrives, paging back down to the end follows again.
#
# Record from the repository root: vhs docs/recordings/component/logpane.tape

Output docs/recordings/component/logpane.gif

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
Type "go build -o logpane.exe ./examples/component/logpane"
Enter
Wait
Type "clear"
Enter
Type ".\logpane.exe"
Enter
Wait+Screen /following/
Show
Sleep 6s

Up@300ms 8
Sleep 4s
PageDown@300ms 4
Sleep 4s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item logpane.exe"
Enter
Wait
