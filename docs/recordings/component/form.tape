# form: one of each control, filled in and submitted.
#
# Record from the repository root: vhs docs/recordings/component/form.tape

Output docs/recordings/component/form.gif

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
Type "go build -o form.exe ./examples/component/form"
Enter
Wait
Type "clear"
Enter
Type ".\form.exe"
Enter
Wait+Screen /new dev container/
Show
Sleep 1.5s

# -- Name: typed into the input --------------------------------------------------------------

Type@120ms "api"
Sleep 1s

# -- Image: the select opens, the arrows move, enter picks -----------------------------------

Tab
Sleep 600ms
Enter
Wait+Screen /alpine/
Sleep 800ms
Down@400ms 2
Sleep 600ms
Enter
Sleep 1s

# -- Shell, docker and the port: a radio, a toggle and a checkbox -----------------------------

Tab
Sleep 500ms
Right@500ms 1
Sleep 800ms
Tab
Sleep 500ms
Space
Sleep 800ms
Tab
Sleep 500ms
Space
Sleep 1s

# -- Create: the status row reports the values -----------------------------------------------

Tab
Sleep 700ms
Enter
Wait+Screen /created api/
Sleep 3s

Hide
Ctrl+C
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item form.exe"
Enter
Wait
