# palette: the : command line: completion, arguments, and the errors it gives.
#
# Record from the repository root: vhs docs/recordings/frame/palette.tape

Output docs/recordings/frame/palette.gif

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
Type "go build -o palette.exe ./examples/frame/palette"
Enter
Wait
Type "clear"
Enter
Type ".\palette.exe"
Enter
Wait+Screen /apple/
Show
Sleep 1.5s

# -- Complete a name, pass an argument -----------------------------------------------

Type ":"
Sleep 1s
Type "go"
Sleep 1s
Tab
Sleep 800ms
Type "mango"
Sleep 800ms
Enter
Sleep 2s

Type ":theme strawberry"
Sleep 800ms
Enter
Sleep 2s

# -- What it cannot run ------------------------------------------------------------------

Type ":xyz"
Sleep 600ms
Enter
Wait+Screen /unknown command/
Sleep 2s

Type ":delete"
Sleep 600ms
Enter
Wait+Screen /not available/
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item palette.exe"
Enter
Wait
