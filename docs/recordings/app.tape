# app: a task list, the app docs/EXAMPLE.md builds step by step.
#
# Record from the repository root: vhs docs/recordings/app.tape

Output docs/recordings/app.gif

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
Set TypingSpeed   60ms
Set Padding       20

# -- Build, off camera ----------------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o app.exe ./examples/app"
Enter
Wait
Type "clear"
Enter
Type ".\app.exe"
Enter
Wait+Screen /record the demo/
Show
Sleep 1.5s

# -- Move and mark a task done -------------------------------------------------------

Down@400ms 1
Sleep 600ms
Space
Wait+Screen /2\/4 done/
Sleep 1.5s

# -- A new task, through a prompt ------------------------------------------------------

Type "n"
Wait+Screen /New task/
Sleep 800ms
Type "ship the docs"
Sleep 600ms
Enter
Wait+Screen /added ship the docs/
Sleep 1.5s

# -- Delete, after a confirm -----------------------------------------------------------

Down@300ms 3
Sleep 600ms
Type "d"
Wait+Screen /Delete task/
Sleep 1.5s
Type "y"
Wait+Screen /deleted/
Sleep 1.5s

# -- Every key, in the help --------------------------------------------------------------

Type "?"
Sleep 2.5s
Escape
Sleep 1s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item app.exe"
Enter
Wait
