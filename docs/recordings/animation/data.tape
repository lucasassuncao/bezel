# animation, data: the data tab: charts whose numbers grow, slide and warm
#
# Record from the repository root: vhs docs/recordings/animation/data.tape

Output docs/recordings/animation/data.gif

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

# -- Build and start, off camera -------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o anim-data.exe ./examples/animation"
Enter
Wait
Type "clear"
Enter
Type ".\anim-data.exe"
Enter
Wait+Screen /every animation/
Tab@600ms 4
Wait+Screen /heat/
Sleep 1s
Show

Sleep 1.8s
Type "p"
Sleep 600ms
Type "p"
Sleep 600ms
Type "p"
Sleep 600ms
Type "p"
Sleep 600ms
Type "n"
Sleep 1.4s
Type "n"
Sleep 1.4s
Type "w"
Sleep 1.4s
Type "w"
Sleep 1.4s
Type "g"
Sleep 2s

# -- Quit and clean up, off camera ------------------------------------------------

Hide
Type "q"
Sleep 1s
Type "Remove-Item anim-data.exe"
Enter
Wait
