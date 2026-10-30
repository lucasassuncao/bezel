# tree: a directory tree: expand, collapse, and :reveal down to a path.
#
# Record from the repository root: vhs docs/recordings/component/tree.tape

Output docs/recordings/component/tree.gif

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
Type "go build -o tree.exe ./examples/component/tree"
Enter
Wait
Type "clear"
Enter
Type ".\tree.exe"
Enter
Wait+Screen /internal/
Show
Sleep 1.5s

Right
Sleep 1s
Down
Sleep 600ms
Right
Sleep 1.2s
Left
Sleep 1s
Up
Sleep 600ms
Left
Sleep 1.2s

Type ":reveal internal/tui/keys/keys.go"
Sleep 800ms
Enter
Wait+Screen /keys.go/
Sleep 2.5s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item tree.exe"
Enter
Wait
