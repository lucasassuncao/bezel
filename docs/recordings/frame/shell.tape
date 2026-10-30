# shell: two tabs, two panes, a status that expires and a busy spinner.
#
# Record from the repository root: vhs docs/recordings/frame/shell.tape

Output docs/recordings/frame/shell.gif

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
Type "go build -o shell.exe ./examples/frame/shell"
Enter
Wait
Type "clear"
Enter
Type ".\shell.exe"
Enter
Wait+Screen /main.go/
Show
Sleep 1.5s

Down@300ms 2
Sleep 800ms

# -- A status that clears itself ----------------------------------------------------

Type "s"
Wait+Screen /saved/
Sleep 2.5s

# -- Busy, then idle -------------------------------------------------------------------

Type "b"
Wait+Screen /indexing/
Sleep 2.5s

# -- Tabs and panes --------------------------------------------------------------------

Tab
Wait+Screen /ideas/
Sleep 1.5s
Shift+Tab
Sleep 1.2s

# -- Every key, one keystroke away ---------------------------------------------------

Type "?"
Sleep 2.5s
Escape
Sleep 1s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item shell.exe"
Enter
Wait
