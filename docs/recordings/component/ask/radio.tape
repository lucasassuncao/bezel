# ask radio: one radio field asked below a CLI's output.
#
# Record from the repository root: vhs docs/recordings/component/ask/radio.tape

Output docs/recordings/component/ask/radio.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        400
Set Theme         {"name":"Catppuccin Mocha on Mint","black":"#45475a","red":"#f38ba8","green":"#a6e3a1","yellow":"#f9e2af","blue":"#89b4fa","magenta":"#f5c2e7","cyan":"#94e2d5","white":"#bac2de","brightBlack":"#585b70","brightRed":"#f38ba8","brightGreen":"#a6e3a1","brightYellow":"#f9e2af","brightBlue":"#89b4fa","brightMagenta":"#f5c2e7","brightCyan":"#94e2d5","brightWhite":"#a6adc8","background":"#0B1B15","foreground":"#cdd6f4","selection":"#585b70","cursor":"#f5e0dc"}
Set Framerate     30
Set PlaybackSpeed 1.0
Set TypingSpeed   90ms
Set Padding       20

# -- Build and start, off camera -------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o ask-radio.exe ./examples/component/ask"
Enter
Wait
# clear runs first, so the command line is gone before the camera rolls, and
# the sleep after holds the last frame without a prompt under it.
Type "clear; ./ask-radio.exe radio; Start-Sleep 60"
Enter
Wait+Screen /Shell/
Show

Sleep 1.2s
Right
Sleep 600ms
Right
Sleep 600ms
Enter
Wait+Screen /shell set/
Sleep 2.5s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Ctrl+C
Sleep 500ms
Type "Remove-Item ask-radio.exe"
Enter
Wait
