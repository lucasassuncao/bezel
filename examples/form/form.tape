# form: one of each control, filled in and submitted.
#
# Record from the repository root: vhs examples/form/form.tape

Output examples/form/form.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      14
Set Width         1200
Set Height        700
Set Theme         "Catppuccin Mocha"
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
Type "go build -o form.exe ./examples/form"
Enter
Wait
Type "clear"
Enter
Show

Type ".\form.exe"
Sleep 400ms
Enter
Wait+Screen /new dev container/
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

Ctrl+C
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item form.exe"
Enter
Wait
