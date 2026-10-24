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
Set Theme         "Catppuccin Mocha"
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
