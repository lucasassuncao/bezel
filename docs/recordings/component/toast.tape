# toast: notices that slide into the corner, stack and leave on their own
#
# Record from the repository root: vhs docs/recordings/component/toast.tape

Output docs/recordings/component/toast.gif

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
Set TypingSpeed   45ms
Set Padding       20

# -- Build, off camera ----------------------------------------------------------

Hide
# A NO_COLOR inherited from the parent shell would record in monochrome.
Type "Remove-Item Env:NO_COLOR -ErrorAction SilentlyContinue"
Enter
Wait
Type "go build -o toast.exe ./examples/component/toast"
Enter
Wait
Type "clear"
Enter
Type ".\toast.exe"
Enter
Wait+Screen /A toast says/
Show
Sleep 1.2s

Type "s"
Sleep 1.2s
Type "c"
Sleep 1.2s
Type "e"
Sleep 1.5s
Type "w"
Sleep 4.5s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item toast.exe"
Enter
Wait
