# input: a name that must not be empty, a command tab completes and a masked password.
#
# Record from the repository root: vhs docs/recordings/component/input.tape

Output docs/recordings/component/input.gif

Require ttyd
Require ffmpeg

Set Shell         "powershell"
Set FontSize      16
Set FontFamily    "JetBrainsMono Nerd Font"
Set Width         1000
Set Height        520
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
Type "go build -o input.exe ./examples/component/input"
Enter
Wait
Type "clear"
Enter
Type ".\input.exe"
Enter
Wait+Screen /deploy/
Show
Sleep 1s

# -- The name, left empty first ------------------------------------------------------

Down
Wait+Screen /a name is required/
Sleep 1.5s
Up
Type "ana"
Sleep 800ms

# -- The command, completed with tab ---------------------------------------------------

Down
Type "ro"
Sleep 1s
Tab
Sleep 1s

# -- The password, behind its mask ----------------------------------------------------

Down
Type "s3cret"
Sleep 1s
Enter
Wait+Screen /rollback/
Sleep 2.5s

Hide
Escape
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item input.exe"
Enter
Wait
