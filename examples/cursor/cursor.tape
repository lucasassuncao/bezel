# cursor: three cursors, one per mode; tab moves the focus, m changes a mode.
#
# Record from the repository root: vhs examples/cursor/cursor.tape

Output examples/cursor/cursor.gif

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
Type "go build -o cursor.exe ./examples/cursor"
Enter
Wait
Type "clear"
Enter
Type ".\cursor.exe"
Enter
Wait+Screen /blinking/
Show
Sleep 2.5s

# -- The focus moves down, then the steady cursor turns hidden and back ---------

Tab
Sleep 1.5s
Type@1.5s "mm"
Sleep 1.5s
Tab
Sleep 2s

Hide
Type "q"
Sleep 1s

# -- Clean up, off camera ---------------------------------------------------------

Hide
Type "Remove-Item cursor.exe"
Enter
Wait
