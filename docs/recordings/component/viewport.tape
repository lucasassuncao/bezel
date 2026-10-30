# viewport: an access log with line numbers; scroll, jump between errors, pan, wrap.
#
# Record from the repository root: vhs docs/recordings/component/viewport.tape

Output docs/recordings/component/viewport.gif

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
Type "go build -o viewport.exe ./examples/component/viewport"
Enter
Wait
Type "clear"
Enter
Type ".\viewport.exe"
Enter
Wait+Screen /access.log/
Show
Sleep 1.5s

# -- Scroll, then from error to error ---------------------------------------------

Down@300ms 5
Sleep 0.5s
PageDown
Sleep 1s
Type@1.2s "nn"
Sleep 1.5s

# -- Pan to the end of the lines, then wrap them -------------------------------------

Right@600ms 3
Sleep 1s
Type "w"
Sleep 1.5s
Type "N"
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item viewport.exe"
Enter
Wait
